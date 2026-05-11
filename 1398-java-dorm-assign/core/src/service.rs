use crate::error::DormError;
use crate::models::*;
use crate::repository::Repository;
use std::str::FromStr;
use std::sync::Arc;

pub struct DormService {
    repo: Arc<Repository>,
}

impl DormService {
    pub fn new(repo: Arc<Repository>) -> Self {
        Self { repo }
    }

    pub fn create_building(&self, name: String, floor_count: u32, rooms_per_floor: u32) -> Building {
        let building = Building::new(name, floor_count, rooms_per_floor);
        self.repo.create_building(building)
    }

    pub fn get_building(&self, id: &str) -> Result<Building, DormError> {
        self.repo.get_building(id)
    }

    pub fn list_buildings(&self) -> Vec<Building> {
        self.repo.list_buildings()
    }

    pub fn create_room(
        &self,
        building_id: String,
        floor_number: u32,
        room_number: u32,
        room_type: RoomType,
    ) -> Result<Room, DormError> {
        let room = Room::new(building_id, floor_number, room_number, room_type);
        self.repo.create_room(room)
    }

    pub fn get_room(&self, id: &str) -> Result<Room, DormError> {
        self.repo.get_room(id)
    }

    pub fn list_rooms(&self, building_id: Option<&str>) -> Vec<Room> {
        self.repo.list_rooms(building_id)
    }

    pub fn set_room_maintenance(&self, room_id: &str, under_maintenance: bool) -> Result<Room, DormError> {
        let mut room = self.repo.get_room(room_id)?;
        room.status = if under_maintenance {
            RoomStatus::UnderMaintenance
        } else {
            RoomStatus::Normal
        };
        self.repo.update_room(room)
    }

    pub fn update_hygiene_score(&self, room_id: &str, score: &str) -> Result<Room, DormError> {
        let hygiene_score = HygieneScore::from_str(score)?;
        let mut room = self.repo.get_room(room_id)?;
        room.hygiene_score = Some(hygiene_score);
        self.repo.update_room(room)
    }

    pub fn list_beds(&self, room_id: Option<&str>) -> Vec<Bed> {
        self.repo.list_beds(room_id)
    }

    pub fn create_student(&self, student_id: String, name: String, department: String) -> Student {
        let student = Student::new(student_id, name, department);
        self.repo.create_student(student)
    }

    pub fn get_student(&self, id: &str) -> Result<Student, DormError> {
        self.repo.get_student(id)
    }

    pub fn list_students(&self, department: Option<&str>, status: Option<StudentStatus>) -> Vec<Student> {
        self.repo.list_students(department, status)
    }

    pub fn allocate_department(&self, department: &str) -> Result<AllocationResult, DormError> {
        let unassigned_students = self.repo.list_students(
            Some(department),
            Some(StudentStatus::Unassigned),
        );

        if unassigned_students.is_empty() {
            return Ok(AllocationResult {
                department: department.to_string(),
                total_students: 0,
                assigned_count: 0,
                unassigned_count: 0,
                assigned_students: vec![],
                unassigned_students: vec![],
                message: "No unassigned students in this department".to_string(),
            });
        }

        let available_beds = self.find_available_beds_by_department(department);
        let total_students = unassigned_students.len();
        let available_count = available_beds.len();
        
        let mut assigned_students = Vec::new();
        let mut unassigned_students_list = Vec::new();
        
        let students_to_assign = std::cmp::min(total_students, available_count);
        
        for i in 0..total_students {
            let student = &unassigned_students[i];
            if i < students_to_assign {
                let bed = &available_beds[i];
                self.assign_bed_to_student(student, bed)?;
                assigned_students.push(student.student_id.clone());
            } else {
                unassigned_students_list.push(student.student_id.clone());
            }
        }

        let message = if total_students > available_count {
            format!(
                "Insufficient beds. Assigned {}/{} students. {} students remain unassigned.",
                students_to_assign, total_students, total_students - students_to_assign
            )
        } else {
            format!("Successfully assigned all {} students", total_students)
        };

        Ok(AllocationResult {
            department: department.to_string(),
            total_students,
            assigned_count: students_to_assign,
            unassigned_count: total_students - students_to_assign,
            assigned_students,
            unassigned_students: unassigned_students_list,
            message,
        })
    }

    fn find_available_beds_by_department(&self, department: &str) -> Vec<Bed> {
        let all_beds = self.repo.list_beds(None);
        
        let assigned_students = self.repo.list_students(
            Some(department),
            Some(StudentStatus::Assigned),
        );
        
        let mut preferred_buildings = std::collections::HashSet::new();
        let mut preferred_floors = std::collections::HashMap::new();
        
        for student in &assigned_students {
            if let Some(bed_id) = &student.bed_id {
                if let Ok(bed) = self.repo.get_bed(bed_id) {
                    if let Ok(room) = self.repo.get_room(&bed.room_id) {
                        preferred_buildings.insert(room.building_id.clone());
                        preferred_floors
                            .entry(room.building_id.clone())
                            .or_insert_with(std::collections::HashSet::new)
                            .insert(room.floor_number);
                    }
                }
            }
        }

        let mut available_beds: Vec<Bed> = all_beds
            .into_iter()
            .filter(|bed| !bed.is_occupied())
            .filter(|bed| {
                if let Ok(room) = self.repo.get_room(&bed.room_id) {
                    room.status == RoomStatus::Normal
                } else {
                    false
                }
            })
            .collect();

        available_beds.sort_by(|a, b| {
            let room_a = self.repo.get_room(&a.room_id).ok();
            let room_b = self.repo.get_room(&b.room_id).ok();
            
            if room_a.is_none() || room_b.is_none() {
                return std::cmp::Ordering::Equal;
            }
            
            let room_a = room_a.unwrap();
            let room_b = room_b.unwrap();
            
            let a_in_preferred_building = preferred_buildings.contains(&room_a.building_id);
            let b_in_preferred_building = preferred_buildings.contains(&room_b.building_id);
            
            if a_in_preferred_building != b_in_preferred_building {
                return a_in_preferred_building.cmp(&b_in_preferred_building).reverse();
            }
            
            if a_in_preferred_building {
                let a_in_preferred_floor = preferred_floors
                    .get(&room_a.building_id)
                    .map_or(false, |floors| floors.contains(&room_a.floor_number));
                let b_in_preferred_floor = preferred_floors
                    .get(&room_b.building_id)
                    .map_or(false, |floors| floors.contains(&room_b.floor_number));
                
                if a_in_preferred_floor != b_in_preferred_floor {
                    return a_in_preferred_floor.cmp(&b_in_preferred_floor).reverse();
                }
            }
            
            room_a.building_id.cmp(&room_b.building_id)
                .then_with(|| room_a.floor_number.cmp(&room_b.floor_number))
                .then_with(|| room_a.room_number.cmp(&room_b.room_number))
                .then_with(|| a.bed_number.cmp(&b.bed_number))
        });

        available_beds
    }

    fn assign_bed_to_student(&self, student: &Student, bed: &Bed) -> Result<(), DormError> {
        if student.is_assigned() {
            return Err(DormError::StudentAlreadyAssigned(student.id.clone()));
        }
        
        if bed.is_occupied() {
            return Err(DormError::BedAlreadyOccupied(bed.id.clone()));
        }

        let room = self.repo.get_room(&bed.room_id)?;
        if room.status == RoomStatus::UnderMaintenance {
            return Err(DormError::RoomUnderMaintenance(room.id));
        }

        let mut updated_student = student.clone();
        updated_student.bed_id = Some(bed.id.clone());
        updated_student.status = StudentStatus::Assigned;
        self.repo.update_student(updated_student)?;

        let mut updated_bed = bed.clone();
        updated_bed.student_id = Some(student.id.clone());
        self.repo.update_bed(updated_bed)?;

        Ok(())
    }

    pub fn check_out(&self, student_id: &str) -> Result<(), DormError> {
        let mut student = self.repo.get_student(student_id)?;
        
        let bed_id = student.bed_id.clone()
            .ok_or_else(|| DormError::NoBedAssigned(student.id.clone()))?;
        
        let mut bed = self.repo.get_bed(&bed_id)?;
        
        bed.student_id = None;
        self.repo.update_bed(bed)?;
        
        student.bed_id = None;
        student.status = StudentStatus::Unassigned;
        self.repo.update_student(student)?;
        
        Ok(())
    }

    pub fn create_swap_request(&self, requester_id: &str, target_id: &str) -> Result<SwapRequest, DormError> {
        if requester_id == target_id {
            return Err(DormError::CannotSwapWithSelf);
        }

        let requester = self.repo.get_student(requester_id)?;
        let target = self.repo.get_student(target_id)?;

        let requester_bed_id = requester.bed_id.clone()
            .ok_or_else(|| DormError::NoBedAssigned(requester.id.clone()))?;
        let target_bed_id = target.bed_id.clone()
            .ok_or_else(|| DormError::NoBedAssigned(target.id.clone()))?;

        if self.repo.has_pending_swap_request(requester_id) {
            return Err(DormError::HasPendingSwapRequest(requester.id.clone()));
        }
        if self.repo.has_pending_swap_request(target_id) {
            return Err(DormError::HasPendingSwapRequest(target.id.clone()));
        }

        let swap_request = SwapRequest::new(
            requester.id.clone(),
            target.id.clone(),
            requester_bed_id,
            target_bed_id,
        );

        Ok(self.repo.create_swap_request(swap_request))
    }

    pub fn respond_swap_request(&self, request_id: &str, target_id: &str, accept: bool) -> Result<SwapRequest, DormError> {
        let mut swap_request = self.repo.get_swap_request(request_id)?;
        
        if swap_request.is_expired() {
            swap_request.status = SwapRequestStatus::Expired;
            self.repo.update_swap_request(swap_request.clone())?;
            return Err(DormError::SwapRequestExpired(swap_request.id));
        }

        if swap_request.status != SwapRequestStatus::Pending {
            return Err(DormError::SwapRequestNotPending(swap_request.id));
        }

        if swap_request.target_id != target_id {
            return Err(DormError::NotSwapTarget(target_id.to_string()));
        }

        if accept {
            self.execute_swap(&swap_request)?;
            swap_request.status = SwapRequestStatus::Confirmed;
        } else {
            swap_request.status = SwapRequestStatus::Rejected;
        }

        self.repo.update_swap_request(swap_request.clone())?;
        Ok(swap_request)
    }

    fn execute_swap(&self, swap_request: &SwapRequest) -> Result<(), DormError> {
        let mut requester = self.repo.get_student(&swap_request.requester_id)?;
        let mut target = self.repo.get_student(&swap_request.target_id)?;
        let mut requester_bed = self.repo.get_bed(&swap_request.requester_bed_id)?;
        let mut target_bed = self.repo.get_bed(&swap_request.target_bed_id)?;

        let old_requester_bed_id = requester.bed_id.clone();
        let old_target_bed_id = target.bed_id.clone();

        requester.bed_id = Some(target_bed.id.clone());
        target.bed_id = Some(requester_bed.id.clone());

        requester_bed.student_id = Some(target.id.clone());
        target_bed.student_id = Some(requester.id.clone());

        self.repo.update_student(requester)?;
        self.repo.update_student(target)?;
        self.repo.update_bed(requester_bed)?;
        self.repo.update_bed(target_bed)?;

        if let Some(bed_id) = old_requester_bed_id {
            let _ = self.repo.get_bed(&bed_id);
        }
        if let Some(bed_id) = old_target_bed_id {
            let _ = self.repo.get_bed(&bed_id);
        }

        Ok(())
    }

    pub fn get_swap_request(&self, id: &str) -> Result<SwapRequest, DormError> {
        self.repo.get_swap_request(id)
    }

    pub fn list_swap_requests(&self, status: Option<SwapRequestStatus>) -> Vec<SwapRequest> {
        self.repo.list_swap_requests(status)
    }

    pub fn cleanup_expired_swap_requests(&self) -> usize {
        let pending_requests = self.repo.list_swap_requests(Some(SwapRequestStatus::Pending));
        let mut cleaned = 0;

        for mut request in pending_requests {
            if request.is_expired() {
                request.status = SwapRequestStatus::Expired;
                if self.repo.update_swap_request(request).is_ok() {
                    cleaned += 1;
                }
            }
        }

        cleaned
    }

    pub fn get_available_beds_count(&self) -> usize {
        self.repo
            .list_beds(None)
            .into_iter()
            .filter(|bed| !bed.is_occupied())
            .filter(|bed| {
                if let Ok(room) = self.repo.get_room(&bed.room_id) {
                    room.status == RoomStatus::Normal
                } else {
                    false
                }
            })
            .count()
    }

    pub fn get_bed(&self, id: &str) -> Result<Bed, DormError> {
        self.repo.get_bed(id)
    }
}
