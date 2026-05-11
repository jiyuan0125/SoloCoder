use std::collections::HashMap;
use std::sync::Mutex;

use crate::error::DormError;
use crate::models::*;

pub struct Repository {
    buildings: Mutex<HashMap<String, Building>>,
    rooms: Mutex<HashMap<String, Room>>,
    beds: Mutex<HashMap<String, Bed>>,
    students: Mutex<HashMap<String, Student>>,
    swap_requests: Mutex<HashMap<String, SwapRequest>>,
}

impl Repository {
    pub fn new() -> Self {
        Self {
            buildings: Mutex::new(HashMap::new()),
            rooms: Mutex::new(HashMap::new()),
            beds: Mutex::new(HashMap::new()),
            students: Mutex::new(HashMap::new()),
            swap_requests: Mutex::new(HashMap::new()),
        }
    }

    pub fn create_building(&self, building: Building) -> Building {
        let mut buildings = self.buildings.lock().unwrap();
        let id = building.id.clone();
        buildings.insert(id, building.clone());
        building
    }

    pub fn get_building(&self, id: &str) -> Result<Building, DormError> {
        let buildings = self.buildings.lock().unwrap();
        buildings
            .get(id)
            .cloned()
            .ok_or_else(|| DormError::BuildingNotFound(id.to_string()))
    }

    pub fn list_buildings(&self) -> Vec<Building> {
        let buildings = self.buildings.lock().unwrap();
        let mut result: Vec<Building> = buildings.values().cloned().collect();
        result.sort_by(|a, b| a.name.cmp(&b.name));
        result
    }

    pub fn create_room(&self, room: Room) -> Result<Room, DormError> {
        let building = self.get_building(&room.building_id)?;
        
        if room.floor_number > building.floor_count {
            return Err(DormError::FloorExceedsMax {
                floor: room.floor_number,
                max_floors: building.floor_count,
            });
        }

        let rooms = self.rooms.lock().unwrap();
        let exists = rooms.values().any(|r| {
            r.building_id == room.building_id 
                && r.floor_number == room.floor_number 
                && r.room_number == room.room_number
        });
        drop(rooms);

        if exists {
            return Err(DormError::RoomNumberExists {
                building_id: room.building_id.clone(),
                room_number: room.room_number,
            });
        }

        let mut rooms = self.rooms.lock().unwrap();
        let id = room.id.clone();
        rooms.insert(id, room.clone());
        
        let mut beds = self.beds.lock().unwrap();
        for bed_number in 1..=room.bed_count() {
            let bed = Bed::new(room.id.clone(), bed_number);
            beds.insert(bed.id.clone(), bed);
        }

        Ok(room)
    }

    pub fn get_room(&self, id: &str) -> Result<Room, DormError> {
        let rooms = self.rooms.lock().unwrap();
        rooms
            .get(id)
            .cloned()
            .ok_or_else(|| DormError::RoomNotFound(id.to_string()))
    }

    pub fn list_rooms(&self, building_id: Option<&str>) -> Vec<Room> {
        let rooms = self.rooms.lock().unwrap();
        let mut result: Vec<Room> = rooms
            .values()
            .filter(|r| building_id.map_or(true, |b| r.building_id == b))
            .cloned()
            .collect();
        result.sort_by(|a, b| {
            a.building_id.cmp(&b.building_id)
                .then_with(|| a.floor_number.cmp(&b.floor_number))
                .then_with(|| a.room_number.cmp(&b.room_number))
        });
        result
    }

    pub fn update_room(&self, room: Room) -> Result<Room, DormError> {
        let mut rooms = self.rooms.lock().unwrap();
        if rooms.contains_key(&room.id) {
            rooms.insert(room.id.clone(), room.clone());
            Ok(room)
        } else {
            Err(DormError::RoomNotFound(room.id))
        }
    }

    pub fn get_bed(&self, id: &str) -> Result<Bed, DormError> {
        let beds = self.beds.lock().unwrap();
        beds
            .get(id)
            .cloned()
            .ok_or_else(|| DormError::BedNotFound(id.to_string()))
    }

    pub fn list_beds(&self, room_id: Option<&str>) -> Vec<Bed> {
        let beds = self.beds.lock().unwrap();
        let mut result: Vec<Bed> = beds
            .values()
            .filter(|b| room_id.map_or(true, |r| b.room_id == r))
            .cloned()
            .collect();
        result.sort_by(|a, b| {
            a.room_id.cmp(&b.room_id)
                .then_with(|| a.bed_number.cmp(&b.bed_number))
        });
        result
    }

    pub fn update_bed(&self, bed: Bed) -> Result<Bed, DormError> {
        let mut beds = self.beds.lock().unwrap();
        if beds.contains_key(&bed.id) {
            beds.insert(bed.id.clone(), bed.clone());
            Ok(bed)
        } else {
            Err(DormError::BedNotFound(bed.id))
        }
    }

    pub fn create_student(&self, student: Student) -> Student {
        let mut students = self.students.lock().unwrap();
        let id = student.id.clone();
        students.insert(id, student.clone());
        student
    }

    pub fn get_student(&self, id: &str) -> Result<Student, DormError> {
        let students = self.students.lock().unwrap();
        students
            .get(id)
            .cloned()
            .ok_or_else(|| DormError::StudentNotFound(id.to_string()))
    }

    pub fn get_student_by_student_id(&self, student_id: &str) -> Option<Student> {
        let students = self.students.lock().unwrap();
        students
            .values()
            .find(|s| s.student_id == student_id)
            .cloned()
    }

    pub fn list_students(&self, department: Option<&str>, status: Option<StudentStatus>) -> Vec<Student> {
        let students = self.students.lock().unwrap();
        let mut result: Vec<Student> = students
            .values()
            .filter(|s| {
                department.map_or(true, |d| s.department == d)
                    && status.as_ref().map_or(true, |st| &s.status == st)
            })
            .cloned()
            .collect();
        result.sort_by(|a, b| a.student_id.cmp(&b.student_id));
        result
    }

    pub fn update_student(&self, student: Student) -> Result<Student, DormError> {
        let mut students = self.students.lock().unwrap();
        if students.contains_key(&student.id) {
            students.insert(student.id.clone(), student.clone());
            Ok(student)
        } else {
            Err(DormError::StudentNotFound(student.id))
        }
    }

    pub fn create_swap_request(&self, swap_request: SwapRequest) -> SwapRequest {
        let mut swap_requests = self.swap_requests.lock().unwrap();
        let id = swap_request.id.clone();
        swap_requests.insert(id, swap_request.clone());
        swap_request
    }

    pub fn get_swap_request(&self, id: &str) -> Result<SwapRequest, DormError> {
        let swap_requests = self.swap_requests.lock().unwrap();
        swap_requests
            .get(id)
            .cloned()
            .ok_or_else(|| DormError::SwapRequestNotFound(id.to_string()))
    }

    pub fn list_swap_requests(&self, status: Option<SwapRequestStatus>) -> Vec<SwapRequest> {
        let swap_requests = self.swap_requests.lock().unwrap();
        let mut result: Vec<SwapRequest> = swap_requests
            .values()
            .filter(|s| status.map_or(true, |st| s.status == st))
            .cloned()
            .collect();
        result.sort_by(|a, b| a.created_at.cmp(&b.created_at));
        result
    }

    pub fn update_swap_request(&self, swap_request: SwapRequest) -> Result<SwapRequest, DormError> {
        let mut swap_requests = self.swap_requests.lock().unwrap();
        if swap_requests.contains_key(&swap_request.id) {
            swap_requests.insert(swap_request.id.clone(), swap_request.clone());
            Ok(swap_request)
        } else {
            Err(DormError::SwapRequestNotFound(swap_request.id))
        }
    }

    pub fn has_pending_swap_request(&self, student_id: &str) -> bool {
        let swap_requests = self.swap_requests.lock().unwrap();
        swap_requests.values().any(|s| {
            s.status == SwapRequestStatus::Pending
                && (s.requester_id == student_id || s.target_id == student_id)
        })
    }
}

impl Default for Repository {
    fn default() -> Self {
        Self::new()
    }
}
