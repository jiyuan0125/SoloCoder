use std::collections::{HashMap, VecDeque};

pub mod models {
    use serde::{Deserialize, Serialize};

    #[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
    pub struct EmployeeId(pub String);

    #[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
    pub struct CourseId(pub String);

    #[derive(Debug, Clone, Serialize, Deserialize)]
    pub struct Course {
        pub id: CourseId,
        pub name: String,
        pub capacity: u32,
        pub registration_start: u64,
        pub registration_end: u64,
    }

    #[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq)]
    pub enum RegistrationStatus {
        Enrolled,
        WaitingList { position: u32 },
        PendingConfirmation { deadline: u64 },
        Cancelled,
    }

    #[derive(Debug, Clone, Serialize, Deserialize)]
    pub struct Registration {
        pub employee_id: EmployeeId,
        pub course_id: CourseId,
        pub status: RegistrationStatus,
        pub created_at: u64,
        pub updated_at: u64,
    }

    #[derive(Debug, Clone, Serialize, Deserialize)]
    pub struct Notification {
        pub id: String,
        pub employee_id: EmployeeId,
        pub course_id: CourseId,
        pub message: String,
        pub created_at: u64,
        pub read: bool,
    }
}

pub mod service {
    use super::models::*;
    use super::*;

    const CONFIRMATION_WINDOW_HOURS: u64 = 24;

    #[derive(Debug, Clone, Default)]
    pub struct TrainingService {
        courses: HashMap<CourseId, Course>,
        registrations: HashMap<CourseId, Vec<Registration>>,
        waiting_lists: HashMap<CourseId, VecDeque<EmployeeId>>,
        notifications: Vec<Notification>,
        next_notification_id: u64,
        current_time: u64,
    }

    impl TrainingService {
        pub fn new() -> Self {
            Self::default()
        }

        pub fn set_time(&mut self, time: u64) {
            self.current_time = time;
        }

        pub fn get_time(&self) -> u64 {
            self.current_time
        }

        pub fn create_course(
            &mut self,
            id: CourseId,
            name: String,
            capacity: u32,
            registration_start: u64,
            registration_end: u64,
        ) -> Result<Course, String> {
            if registration_end <= registration_start {
                return Err("Registration end must be after start".to_string());
            }
            if capacity == 0 {
                return Err("Capacity must be greater than zero".to_string());
            }

            let course = Course {
                id: id.clone(),
                name,
                capacity,
                registration_start,
                registration_end,
            };

            if self.courses.contains_key(&id) {
                return Err("Course already exists".to_string());
            }

            self.courses.insert(id.clone(), course.clone());
            self.registrations.insert(id.clone(), Vec::new());
            self.waiting_lists.insert(id, VecDeque::new());

            Ok(course)
        }

        pub fn get_course(&self, id: &CourseId) -> Option<&Course> {
            self.courses.get(id)
        }

        pub fn list_courses(&self) -> Vec<Course> {
            self.courses.values().cloned().collect()
        }

        pub fn register(&mut self, employee_id: EmployeeId, course_id: CourseId) -> Result<Registration, String> {
            let course = self.courses.get(&course_id)
                .ok_or_else(|| "Course not found".to_string())?;

            let now = self.current_time;
            if now < course.registration_start {
                return Err("Registration not started yet".to_string());
            }
            if now >= course.registration_end {
                return Err("Registration already ended".to_string());
            }

            let course_registrations = self.registrations.get_mut(&course_id)
                .ok_or_else(|| "Course registrations not found".to_string())?;

            if let Some(existing) = course_registrations.iter().find(|r| r.employee_id == employee_id) {
                match existing.status {
                    RegistrationStatus::Enrolled => {
                        return Err("Already registered for this course".to_string());
                    }
                    RegistrationStatus::PendingConfirmation { .. } => {
                        return Err("Registration pending confirmation".to_string());
                    }
                    RegistrationStatus::WaitingList { .. } => {
                        return Err("Already on waiting list".to_string());
                    }
                    RegistrationStatus::Cancelled => {}
                }
            }

            let enrolled_count = course_registrations.iter()
                .filter(|r| matches!(r.status, RegistrationStatus::Enrolled))
                .count() as u32;

            let registration = if enrolled_count < course.capacity {
                Registration {
                    employee_id: employee_id.clone(),
                    course_id: course_id.clone(),
                    status: RegistrationStatus::Enrolled,
                    created_at: now,
                    updated_at: now,
                }
            } else {
                let waiting_list = self.waiting_lists.get_mut(&course_id)
                    .ok_or_else(|| "Waiting list not found".to_string())?;
                
                waiting_list.push_back(employee_id.clone());
                let position = waiting_list.len() as u32;

                Registration {
                    employee_id: employee_id.clone(),
                    course_id: course_id.clone(),
                    status: RegistrationStatus::WaitingList { position },
                    created_at: now,
                    updated_at: now,
                }
            };

            course_registrations.push(registration.clone());

            Ok(registration)
        }

        pub fn cancel_registration(&mut self, employee_id: EmployeeId, course_id: CourseId) -> Result<(), String> {
            let _course = self.courses.get(&course_id)
                .ok_or_else(|| "Course not found".to_string())?;

            let course_registrations = self.registrations.get_mut(&course_id)
                .ok_or_else(|| "Course registrations not found".to_string())?;

            let registration = course_registrations.iter_mut()
                .find(|r| r.employee_id == employee_id)
                .ok_or_else(|| "Registration not found".to_string())?;

            match &registration.status {
                RegistrationStatus::Enrolled => {
                    registration.status = RegistrationStatus::Cancelled;
                    registration.updated_at = self.current_time;
                    self.try_promote_next(&course_id);
                }
                RegistrationStatus::WaitingList { .. } => {
                    registration.status = RegistrationStatus::Cancelled;
                    registration.updated_at = self.current_time;
                    
                    let waiting_list = self.waiting_lists.get_mut(&course_id)
                        .ok_or_else(|| "Waiting list not found".to_string())?;
                    
                    if let Some(pos) = waiting_list.iter().position(|e| e == &employee_id) {
                        waiting_list.remove(pos);
                    }
                    
                    self.update_waiting_list_positions(&course_id);
                }
                RegistrationStatus::PendingConfirmation { .. } => {
                    registration.status = RegistrationStatus::Cancelled;
                    registration.updated_at = self.current_time;
                    self.try_promote_next(&course_id);
                }
                RegistrationStatus::Cancelled => {
                    return Err("Registration already cancelled".to_string());
                }
            }

            Ok(())
        }

        fn try_promote_next(&mut self, course_id: &CourseId) {
            let course = match self.courses.get(course_id) {
                Some(c) => c,
                None => return,
            };

            if self.current_time >= course.registration_end {
                return;
            }

            let course_registrations = match self.registrations.get_mut(course_id) {
                Some(r) => r,
                None => return,
            };

            let enrolled_count = course_registrations.iter()
                .filter(|r| matches!(r.status, RegistrationStatus::Enrolled))
                .count() as u32;

            if enrolled_count >= course.capacity {
                return;
            }

            let waiting_list = match self.waiting_lists.get_mut(course_id) {
                Some(w) => w,
                None => return,
            };

            while let Some(next_employee) = waiting_list.pop_front() {
                let registration = course_registrations.iter_mut()
                    .find(|r| r.employee_id == next_employee);

                if let Some(reg) = registration {
                    let deadline = self.current_time + CONFIRMATION_WINDOW_HOURS * 3600;
                    reg.status = RegistrationStatus::PendingConfirmation { deadline };
                    reg.updated_at = self.current_time;

                    self.add_notification(
                        next_employee.clone(),
                        course_id.clone(),
                        format!("You have been promoted from the waiting list for course '{}'. Please confirm within {} hours to secure your spot.", course.name, CONFIRMATION_WINDOW_HOURS),
                    );

                    self.update_waiting_list_positions(course_id);
                    break;
                }
            }
        }

        fn update_waiting_list_positions(&mut self, course_id: &CourseId) {
            let waiting_list = match self.waiting_lists.get(course_id) {
                Some(w) => w,
                None => return,
            };

            let course_registrations = match self.registrations.get_mut(course_id) {
                Some(r) => r,
                None => return,
            };

            for (idx, employee_id) in waiting_list.iter().enumerate() {
                if let Some(reg) = course_registrations.iter_mut().find(|r| r.employee_id == *employee_id) {
                    if matches!(reg.status, RegistrationStatus::WaitingList { .. }) {
                        reg.status = RegistrationStatus::WaitingList { position: (idx + 1) as u32 };
                    }
                }
            }
        }

        pub fn confirm_registration(&mut self, employee_id: EmployeeId, course_id: CourseId) -> Result<Registration, String> {
            let now = self.current_time;
            let course_registrations = self.registrations.get_mut(&course_id)
                .ok_or_else(|| "Course registrations not found".to_string())?;

            let registration = course_registrations.iter_mut()
                .find(|r| r.employee_id == employee_id)
                .ok_or_else(|| "Registration not found".to_string())?;

            match registration.status {
                RegistrationStatus::PendingConfirmation { deadline } => {
                    if now > deadline {
                        let waiting_list = self.waiting_lists.get_mut(&course_id)
                            .ok_or_else(|| "Waiting list not found".to_string())?;
                        
                        waiting_list.push_back(employee_id.clone());
                        registration.status = RegistrationStatus::WaitingList {
                            position: waiting_list.len() as u32,
                        };
                        registration.updated_at = now;
                        self.update_waiting_list_positions(&course_id);
                        
                        return Err("Confirmation window expired. You have been moved to the end of the waiting list.".to_string());
                    }

                    registration.status = RegistrationStatus::Enrolled;
                    registration.updated_at = now;
                    Ok(registration.clone())
                }
                RegistrationStatus::Enrolled => Err("Already enrolled".to_string()),
                RegistrationStatus::WaitingList { .. } => Err("Not in confirmation state".to_string()),
                RegistrationStatus::Cancelled => Err("Registration cancelled".to_string()),
            }
        }

        pub fn check_expired_confirmations(&mut self) {
            let now = self.current_time;
            let course_ids: Vec<CourseId> = self.courses.keys().cloned().collect();

            for course_id in course_ids {
                let course = match self.courses.get(&course_id) {
                    Some(c) => c,
                    None => continue,
                };

                if now >= course.registration_end {
                    continue;
                }

                let expired_employees: Vec<EmployeeId> = self.registrations.get(&course_id)
                    .map(|regs| {
                        regs.iter()
                            .filter_map(|r| {
                                if let RegistrationStatus::PendingConfirmation { deadline } = r.status {
                                    if now > deadline {
                                        Some(r.employee_id.clone())
                                    } else {
                                        None
                                    }
                                } else {
                                    None
                                }
                            })
                            .collect()
                    })
                    .unwrap_or_default();

                for employee_id in expired_employees {
                    if let Some(waiting_list) = self.waiting_lists.get_mut(&course_id) {
                        waiting_list.push_back(employee_id.clone());
                        let new_position = waiting_list.len() as u32;
                        
                        if let Some(regs) = self.registrations.get_mut(&course_id) {
                            if let Some(reg) = regs.iter_mut().find(|r| r.employee_id == employee_id) {
                                reg.status = RegistrationStatus::WaitingList {
                                    position: new_position,
                                };
                                reg.updated_at = now;
                            }
                        }
                    }
                    
                    self.update_waiting_list_positions(&course_id);
                    self.try_promote_next(&course_id);
                }
            }
        }

        pub fn get_registrations(&self, course_id: &CourseId) -> Vec<Registration> {
            self.registrations.get(course_id)
                .cloned()
                .unwrap_or_default()
        }

        pub fn get_employee_registrations(&self, employee_id: &EmployeeId) -> Vec<Registration> {
            self.registrations.values()
                .flat_map(|regs| regs.iter())
                .filter(|r| r.employee_id == *employee_id)
                .cloned()
                .collect()
        }

        pub fn get_waiting_list(&self, course_id: &CourseId) -> Vec<EmployeeId> {
            self.waiting_lists.get(course_id)
                .map(|w| w.iter().cloned().collect())
                .unwrap_or_default()
        }

        fn add_notification(&mut self, employee_id: EmployeeId, course_id: CourseId, message: String) {
            let notification = Notification {
                id: format!("notif-{}", self.next_notification_id),
                employee_id,
                course_id,
                message,
                created_at: self.current_time,
                read: false,
            };
            self.next_notification_id += 1;
            self.notifications.push(notification);
        }

        pub fn get_notifications(&self, employee_id: &EmployeeId) -> Vec<Notification> {
            self.notifications.iter()
                .filter(|n| n.employee_id == *employee_id)
                .cloned()
                .collect()
        }

        pub fn mark_notification_read(&mut self, notification_id: &str) -> Result<(), String> {
            let notification = self.notifications.iter_mut()
                .find(|n| n.id == notification_id)
                .ok_or_else(|| "Notification not found".to_string())?;
            notification.read = true;
            Ok(())
        }
    }
}
