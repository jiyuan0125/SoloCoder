use std::collections::HashMap;
use std::sync::Mutex;
use chrono::{Local, NaiveDate, NaiveTime, Duration};
use uuid::Uuid;

use crate::error::{DeskError, Result};
use crate::models::{
    add_working_days, now_naive, Desk, DeskStatus, DeskType, Department, Employee, 
    Notification, Reservation, ReservationStatus,
};

pub struct DeskService {
    departments: Mutex<HashMap<Uuid, Department>>,
    employees: Mutex<HashMap<Uuid, Employee>>,
    desks: Mutex<HashMap<Uuid, Desk>>,
    reservations: Mutex<HashMap<Uuid, Reservation>>,
    notifications: Mutex<HashMap<Uuid, Notification>>,
    desk_locks: Mutex<HashMap<Uuid, std::sync::Mutex<()>>>,
}

impl Default for DeskService {
    fn default() -> Self {
        Self::new()
    }
}

impl DeskService {
    pub fn new() -> Self {
        Self {
            departments: Mutex::new(HashMap::new()),
            employees: Mutex::new(HashMap::new()),
            desks: Mutex::new(HashMap::new()),
            reservations: Mutex::new(HashMap::new()),
            notifications: Mutex::new(HashMap::new()),
            desk_locks: Mutex::new(HashMap::new()),
        }
    }
    
    pub fn create_department(&self, name: String, floor: i32) -> Department {
        let dept = Department::new(name, floor);
        let mut depts = self.departments.lock().unwrap();
        depts.insert(dept.id, dept.clone());
        dept
    }
    
    pub fn get_department(&self, id: Uuid) -> Result<Department> {
        let depts = self.departments.lock().unwrap();
        depts.get(&id).cloned().ok_or(DeskError::DepartmentNotFound(id.to_string()))
    }
    
    pub fn list_departments(&self) -> Vec<Department> {
        let depts = self.departments.lock().unwrap();
        depts.values().cloned().collect()
    }
    
    pub fn create_desk(&self, code: String, desk_type: DeskType, floor: i32) -> Desk {
        let desk = Desk::new(code, desk_type, floor);
        let mut desks = self.desks.lock().unwrap();
        desks.insert(desk.id, desk.clone());
        self.desk_locks.lock().unwrap().insert(desk.id, std::sync::Mutex::new(()));
        desk
    }
    
    pub fn get_desk(&self, id: Uuid) -> Result<Desk> {
        let desks = self.desks.lock().unwrap();
        desks.get(&id).cloned().ok_or(DeskError::DeskNotFound(id.to_string()))
    }
    
    pub fn list_desks(&self) -> Vec<Desk> {
        let desks = self.desks.lock().unwrap();
        desks.values().cloned().collect()
    }
    
    pub fn create_employee(&self, name: String, department_id: Uuid) -> Result<(Employee, Desk)> {
        let dept = self.get_department(department_id)?;
        let employee = Employee::new(name, department_id);
        
        let desk = self.allocate_fixed_desk(dept.floor, department_id)?;
        
        let mut employees = self.employees.lock().unwrap();
        employees.insert(employee.id, employee.clone());
        
        Ok((employee, desk))
    }
    
    pub fn batch_create_employees(&self, requests: Vec<(String, Uuid)>) -> Result<Vec<(Employee, Desk)>> {
        let mut results = Vec::new();
        for (name, dept_id) in requests {
            let result = self.create_employee(name, dept_id)?;
            results.push(result);
        }
        Ok(results)
    }
    
    pub fn get_employee(&self, id: Uuid) -> Result<Employee> {
        let employees = self.employees.lock().unwrap();
        employees.get(&id).cloned().ok_or(DeskError::EmployeeNotFound(id.to_string()))
    }
    
    pub fn list_employees(&self) -> Vec<Employee> {
        let employees = self.employees.lock().unwrap();
        employees.values().cloned().collect()
    }
    
    pub fn employee_leave(&self, employee_id: Uuid) -> Result<()> {
        let mut employees = self.employees.lock().unwrap();
        let employee = employees.get_mut(&employee_id)
            .ok_or(DeskError::EmployeeNotFound(employee_id.to_string()))?;
        
        if !employee.is_active {
            return Err(DeskError::InvalidOperation("员工已离职".to_string()));
        }
        
        employee.is_active = false;
        employee.leave_date = Some(now_naive());
        let dept_id = employee.department_id;
        
        let mut desks = self.desks.lock().unwrap();
        for desk in desks.values_mut() {
            if desk.current_employee_id == Some(employee_id) && desk.desk_type == DeskType::Fixed {
                desk.status = DeskStatus::PendingRelease;
                desk.current_employee_id = None;
                desk.reserved_for_department = Some(dept_id);
                desk.reservation_expiry_date = Some(now_naive() + Duration::days(30));
                break;
            }
        }
        
        Ok(())
    }
    
    pub fn allocate_fixed_desk(&self, preferred_floor: i32, department_id: Uuid) -> Result<Desk> {
        let mut desks = self.desks.lock().unwrap();
        let today = now_naive();
        
        for desk in desks.values_mut() {
            if desk.desk_type == DeskType::Fixed {
                if let Some(expiry) = desk.reservation_expiry_date {
                    if expiry < today {
                        desk.reserved_for_department = None;
                        desk.reservation_expiry_date = None;
                        desk.status = DeskStatus::Available;
                    }
                }
            }
        }
        
        let available_desk = desks.values_mut()
            .filter(|d| d.desk_type == DeskType::Fixed)
            .filter(|d| {
                if d.status == DeskStatus::Available {
                    true
                } else if d.status == DeskStatus::PendingRelease {
                    d.reserved_for_department == Some(department_id)
                } else {
                    false
                }
            })
            .min_by_key(|d| {
                let floor_diff = (d.floor - preferred_floor).abs();
                if d.floor == preferred_floor {
                    (0, 0)
                } else if d.reserved_for_department == Some(department_id) {
                    (1, floor_diff)
                } else {
                    (2, floor_diff)
                }
            });
        
        match available_desk {
            Some(desk) => {
                desk.status = DeskStatus::Occupied;
                desk.reserved_for_department = None;
                desk.reservation_expiry_date = None;
                Ok(desk.clone())
            }
            None => Err(DeskError::NoAvailableFixedDesk),
        }
    }
    
    pub fn reserve_shared_desk(&self, employee_id: Uuid, desk_id: Uuid, date: NaiveDate) -> Result<Reservation> {
        let today = now_naive();
        let max_reservation_date = add_working_days(today, 5);
        
        if date < today || date > max_reservation_date {
            return Err(DeskError::ReservationTimeOutOfRange);
        }
        
        let employee = self.get_employee(employee_id)?;
        if !employee.is_active {
            return Err(DeskError::InvalidOperation("员工已离职".to_string()));
        }
        
        let desk = self.get_desk(desk_id)?;
        if desk.desk_type != DeskType::Shared {
            return Err(DeskError::InvalidOperation("只能预约共享工位".to_string()));
        }
        
        let reservations = self.reservations.lock().unwrap();
        let has_reservation = reservations.values()
            .any(|r| r.employee_id == employee_id && r.date == date && 
                 (r.status == ReservationStatus::Confirmed || r.status == ReservationStatus::CheckedIn));
        
        if has_reservation {
            return Err(DeskError::EmployeeAlreadyHasReservation);
        }
        
        let is_desk_taken = reservations.values()
            .any(|r| r.desk_id == desk_id && r.date == date && 
                 (r.status == ReservationStatus::Confirmed || r.status == ReservationStatus::CheckedIn));
        
        if is_desk_taken {
            return Err(DeskError::DeskAlreadyOccupied);
        }
        
        drop(reservations);
        
        let reservation = Reservation::new(desk_id, employee_id, date);
        let mut reservations = self.reservations.lock().unwrap();
        reservations.insert(reservation.id, reservation.clone());
        
        Ok(reservation)
    }
    
    pub fn check_in_reservation(&self, reservation_id: Uuid) -> Result<()> {
        let mut reservations = self.reservations.lock().unwrap();
        let reservation = reservations.get_mut(&reservation_id)
            .ok_or(DeskError::ReservationNotFound(reservation_id.to_string()))?;
        
        if reservation.status != ReservationStatus::Confirmed {
            return Err(DeskError::InvalidOperation("只能签到已确认的预约".to_string()));
        }
        
        let today = now_naive();
        if reservation.date < today {
            reservation.status = ReservationStatus::Expired;
            return Err(DeskError::ReservationExpired);
        }
        
        if reservation.date == today {
            let now = Local::now().time();
            let check_in_deadline = NaiveTime::from_hms_opt(9, 30, 0).unwrap();
            if now > check_in_deadline {
                reservation.status = ReservationStatus::Expired;
                return Err(DeskError::ReservationExpired);
            }
        }
        
        reservation.status = ReservationStatus::CheckedIn;
        reservation.check_in_time = Some(Local::now());
        
        Ok(())
    }
    
    pub fn cancel_reservation(&self, reservation_id: Uuid) -> Result<()> {
        let mut reservations = self.reservations.lock().unwrap();
        let reservation = reservations.get_mut(&reservation_id)
            .ok_or(DeskError::ReservationNotFound(reservation_id.to_string()))?;
        
        if reservation.status != ReservationStatus::Confirmed {
            return Err(DeskError::InvalidOperation("只能取消已确认的预约".to_string()));
        }
        
        reservation.status = ReservationStatus::Cancelled;
        Ok(())
    }
    
    pub fn list_reservations(&self) -> Vec<Reservation> {
        let reservations = self.reservations.lock().unwrap();
        reservations.values().cloned().collect()
    }
    
    pub fn get_reservation(&self, id: Uuid) -> Result<Reservation> {
        let reservations = self.reservations.lock().unwrap();
        reservations.get(&id).cloned().ok_or(DeskError::ReservationNotFound(id.to_string()))
    }
    
    pub fn batch_relocate_department(&self, department_id: Uuid, target_floor: i32) -> Result<Vec<(Uuid, Uuid)>> {
        let employees: Vec<Employee> = {
            let employees = self.employees.lock().unwrap();
            employees.values()
                .filter(|e| e.department_id == department_id && e.is_active)
                .cloned()
                .collect()
        };
        
        let mut old_new_desk_map = Vec::new();
        let mut allocated_desks = std::collections::HashSet::new();
        
        for emp in &employees {
            let old_desk = {
                let desks = self.desks.lock().unwrap();
                desks.values()
                    .find(|d| d.current_employee_id == Some(emp.id) && d.desk_type == DeskType::Fixed)
                    .cloned()
            };
            
            if let Some(old_desk) = old_desk {
                let new_desk = self.allocate_fixed_desk(target_floor, department_id)?;
                
                if allocated_desks.contains(&new_desk.id) {
                    return Err(DeskError::BatchConflict(format!("工位 {} 已被分配", new_desk.code)));
                }
                
                allocated_desks.insert(new_desk.id);
                
                let mut desks = self.desks.lock().unwrap();
                if let Some(old) = desks.get_mut(&old_desk.id) {
                    old.current_employee_id = None;
                    old.status = DeskStatus::Available;
                }
                
                if let Some(new) = desks.get_mut(&new_desk.id) {
                    new.current_employee_id = Some(emp.id);
                }
                
                old_new_desk_map.push((old_desk.id, new_desk.id));
            }
        }
        
        Ok(old_new_desk_map)
    }
    
    pub fn process_expired_reservations(&self) -> Vec<Reservation> {
        let now = Local::now();
        let today = now.date_naive();
        let current_time = now.time();
        let check_in_deadline = NaiveTime::from_hms_opt(9, 30, 0).unwrap();
        
        let mut expired = Vec::new();
        let mut reservations = self.reservations.lock().unwrap();
        
        for reservation in reservations.values_mut() {
            if reservation.status == ReservationStatus::Confirmed {
                if reservation.date < today {
                    reservation.status = ReservationStatus::Expired;
                    expired.push(reservation.clone());
                } else if reservation.date == today && current_time > check_in_deadline {
                    reservation.status = ReservationStatus::Expired;
                    expired.push(reservation.clone());
                }
            }
        }
        
        expired
    }
    
    pub fn process_pending_release_desks(&self) -> (Vec<Desk>, Vec<Notification>) {
        let today = now_naive();
        let three_days_later = today + Duration::days(3);
        let mut released = Vec::new();
        let mut notifications = Vec::new();
        
        let mut desks = self.desks.lock().unwrap();
        let departments = self.departments.lock().unwrap();
        
        for desk in desks.values_mut() {
            if desk.status == DeskStatus::PendingRelease {
                if let Some(expiry) = desk.reservation_expiry_date {
                    if expiry < today {
                        desk.status = DeskStatus::Available;
                        desk.reserved_for_department = None;
                        desk.reservation_expiry_date = None;
                        released.push(desk.clone());
                    } else if expiry == three_days_later {
                        if let Some(dept_id) = desk.reserved_for_department {
                            if let Some(dept) = departments.get(&dept_id) {
                                if let Some(manager_id) = dept.manager_id {
                                    let msg = format!("工位 {} 将于3天后释放，请及时安排新员工使用", desk.code);
                                    let notification = Notification::new(manager_id, msg);
                                    notifications.push(notification.clone());
                                    
                                    let mut notifs = self.notifications.lock().unwrap();
                                    notifs.insert(notification.id, notification);
                                }
                            }
                        }
                    }
                }
            }
        }
        
        (released, notifications)
    }
    
    pub fn get_employee_desk(&self, employee_id: Uuid) -> Option<Desk> {
        let desks = self.desks.lock().unwrap();
        desks.values()
            .find(|d| d.current_employee_id == Some(employee_id))
            .cloned()
    }
    
    pub fn list_available_shared_desks(&self, date: NaiveDate) -> Vec<Desk> {
        let desks = self.desks.lock().unwrap();
        let reservations = self.reservations.lock().unwrap();
        
        let reserved_desk_ids: std::collections::HashSet<Uuid> = reservations.values()
            .filter(|r| r.date == date && (r.status == ReservationStatus::Confirmed || r.status == ReservationStatus::CheckedIn))
            .map(|r| r.desk_id)
            .collect();
        
        desks.values()
            .filter(|d| d.desk_type == DeskType::Shared && d.status == DeskStatus::Available)
            .filter(|d| !reserved_desk_ids.contains(&d.id))
            .cloned()
            .collect()
    }
    
    pub fn assign_desk_to_employee(&self, desk_id: Uuid, employee_id: Uuid) -> Result<()> {
        let mut desks = self.desks.lock().unwrap();
        let desk = desks.get_mut(&desk_id)
            .ok_or(DeskError::DeskNotFound(desk_id.to_string()))?;
        
        if desk.current_employee_id.is_some() {
            return Err(DeskError::DeskAlreadyOccupied);
        }
        
        desk.current_employee_id = Some(employee_id);
        desk.status = DeskStatus::Occupied;
        
        Ok(())
    }
}
