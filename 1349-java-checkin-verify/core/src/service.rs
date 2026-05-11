use crate::error::CheckinError;
use crate::models::{
    Activity, CheckinRecord, CheckinReport, CheckinRequest, CheckinStatus, CheckoutRequest,
    CreateActivityRequest, CreateEmployeeRequest, Employee, GpsCoordinate,
};
use crate::utils::is_within_range;
use chrono::{DateTime, Utc};
use std::collections::HashMap;
use std::sync::{Arc, RwLock};
use uuid::Uuid;

#[derive(Clone, Default)]
pub struct CheckinService {
    inner: Arc<Inner>,
}

#[derive(Default)]
struct Inner {
    activities: RwLock<HashMap<Uuid, Activity>>,
    employees: RwLock<HashMap<Uuid, Employee>>,
    activity_employees: RwLock<HashMap<Uuid, Vec<Uuid>>>,
    records: RwLock<HashMap<(Uuid, Uuid), CheckinRecord>>,
    device_usage: RwLock<HashMap<String, Uuid>>,
}

impl CheckinService {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn create_employee(&self, req: CreateEmployeeRequest) -> Employee {
        let employee = Employee {
            id: Uuid::new_v4(),
            name: req.name,
            employee_number: req.employee_number,
        };
        self.inner
            .employees
            .write()
            .unwrap()
            .insert(employee.id, employee.clone());
        employee
    }

    pub fn get_employee(&self, id: Uuid) -> Option<Employee> {
        self.inner.employees.read().unwrap().get(&id).cloned()
    }

    pub fn list_employees(&self) -> Vec<Employee> {
        self.inner
            .employees
            .read()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    pub fn create_activity(&self, req: CreateActivityRequest) -> Activity {
        let activity = Activity {
            id: Uuid::new_v4(),
            name: req.name,
            checkin_start: req.checkin_start,
            checkin_end: req.checkin_end,
            checkout_deadline: req.checkout_deadline,
            location: req.location,
            allowed_distance_meters: req.allowed_distance_meters,
        };

        self.inner
            .activities
            .write()
            .unwrap()
            .insert(activity.id, activity.clone());

        self.inner
            .activity_employees
            .write()
            .unwrap()
            .insert(activity.id, req.employee_ids.clone());

        activity
    }

    pub fn get_activity(&self, id: Uuid) -> Option<Activity> {
        self.inner.activities.read().unwrap().get(&id).cloned()
    }

    pub fn list_activities(&self) -> Vec<Activity> {
        self.inner
            .activities
            .read()
            .unwrap()
            .values()
            .cloned()
            .collect()
    }

    pub fn get_activity_employees(&self, activity_id: Uuid) -> Option<Vec<Uuid>> {
        self.inner
            .activity_employees
            .read()
            .unwrap()
            .get(&activity_id)
            .cloned()
    }

    pub fn checkin(
        &self,
        activity_id: Uuid,
        req: CheckinRequest,
        now: DateTime<Utc>,
    ) -> Result<CheckinRecord, CheckinError> {
        let activity = self
            .get_activity(activity_id)
            .ok_or(CheckinError::ActivityNotFound)?;

        let _employee = self
            .get_activity_employees(activity_id)
            .ok_or(CheckinError::ActivityNotFound)?
            .iter()
            .find(|&&id| id == req.employee_id)
            .ok_or(CheckinError::EmployeeNotFound)?;

        if now < activity.checkin_start {
            return Err(CheckinError::ActivityNotStarted);
        }
        if now > activity.checkin_end {
            return Err(CheckinError::ActivityEnded);
        }

        if !is_within_range(&activity.location, &req.location, activity.allowed_distance_meters) {
            return Err(CheckinError::LocationOutOfRange);
        }

        let key = (activity_id, req.employee_id);
        let records_read = self.inner.records.read().unwrap();
        if let Some(existing) = records_read.get(&key) {
            if existing.checkin_time.is_some() {
                return Err(CheckinError::AlreadyCheckedIn);
            }
        }
        drop(records_read);

        let mut device_usage_write = self.inner.device_usage.write().unwrap();
        if let Some(&existing_employee) = device_usage_write.get(&req.device_id) {
            if existing_employee != req.employee_id {
                return Err(CheckinError::DeviceAlreadyUsedByOther);
            }
        }
        device_usage_write.insert(req.device_id.clone(), req.employee_id);
        drop(device_usage_write);

        let mut records_write = self.inner.records.write().unwrap();
        let record = records_write
            .entry(key)
            .or_insert_with(|| {
                CheckinRecord::new(activity_id, req.employee_id, req.device_id.clone())
            });

        record.checkin_time = Some(now);
        record.checkin_location = Some(GpsCoordinate {
            latitude: req.location.latitude,
            longitude: req.location.longitude,
        });
        record.status = CheckinStatus::NotCheckedOut;

        Ok(record.clone())
    }

    pub fn checkout(
        &self,
        activity_id: Uuid,
        req: CheckoutRequest,
        now: DateTime<Utc>,
    ) -> Result<CheckinRecord, CheckinError> {
        let activity = self
            .get_activity(activity_id)
            .ok_or(CheckinError::ActivityNotFound)?;

        let _employee = self
            .get_activity_employees(activity_id)
            .ok_or(CheckinError::ActivityNotFound)?
            .iter()
            .find(|&&id| id == req.employee_id)
            .ok_or(CheckinError::EmployeeNotFound)?;

        if !is_within_range(&activity.location, &req.location, activity.allowed_distance_meters) {
            return Err(CheckinError::LocationOutOfRange);
        }

        let key = (activity_id, req.employee_id);
        let mut records_write = self.inner.records.write().unwrap();
        let record = records_write
            .get_mut(&key)
            .ok_or(CheckinError::NotCheckedIn)?;

        if record.checkout_time.is_some() {
            return Err(CheckinError::AlreadyCheckedOut);
        }

        record.checkout_time = Some(now);
        record.checkout_location = Some(GpsCoordinate {
            latitude: req.location.latitude,
            longitude: req.location.longitude,
        });

        if now < activity.checkout_deadline {
            record.status = CheckinStatus::EarlyCheckout;
            Err(CheckinError::EarlyCheckout)
        } else {
            record.status = CheckinStatus::Normal;
            Ok(record.clone())
        }
    }

    pub fn get_record(&self, activity_id: Uuid, employee_id: Uuid) -> Option<CheckinRecord> {
        self.inner
            .records
            .read()
            .unwrap()
            .get(&(activity_id, employee_id))
            .cloned()
    }

    pub fn get_report(&self, activity_id: Uuid) -> Result<CheckinReport, CheckinError> {
        let activity = self
            .get_activity(activity_id)
            .ok_or(CheckinError::ActivityNotFound)?;

        let employees = self
            .get_activity_employees(activity_id)
            .ok_or(CheckinError::ActivityNotFound)?;

        let records = self.inner.records.read().unwrap();

        let total_count = employees.len();
        let mut checked_in_count = 0;
        let mut normal_count = 0;
        let mut not_checked_out_count = 0;
        let mut early_checkout_count = 0;

        for emp_id in &employees {
            let key = (activity.id, *emp_id);
            match records.get(&key) {
                Some(record) => {
                    if record.checkin_time.is_some() {
                        checked_in_count += 1;
                    }
                    match record.status {
                        CheckinStatus::Normal => normal_count += 1,
                        CheckinStatus::NotCheckedOut => not_checked_out_count += 1,
                        CheckinStatus::EarlyCheckout => early_checkout_count += 1,
                        CheckinStatus::NotCheckedIn => {}
                    }
                }
                None => {}
            }
        }

        let checkin_rate = if total_count > 0 {
            checked_in_count as f64 / total_count as f64
        } else {
            0.0
        };

        let not_checked_in_count = total_count - checked_in_count;

        Ok(CheckinReport {
            activity_id,
            total_count,
            checked_in_count,
            checkin_rate,
            normal_count,
            not_checked_out_count,
            early_checkout_count,
            not_checked_in_count,
        })
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use chrono::Duration;

    fn create_test_coords() -> GpsCoordinate {
        GpsCoordinate {
            latitude: 39.9042,
            longitude: 116.4074,
        }
    }

    fn create_test_activity(
        service: &CheckinService,
        employees: &[Uuid],
        offset_minutes: i64,
    ) -> Activity {
        let now = Utc::now();
        service.create_activity(CreateActivityRequest {
            name: "测试活动".to_string(),
            checkin_start: now - Duration::minutes(30),
            checkin_end: now + Duration::minutes(offset_minutes),
            checkout_deadline: now + Duration::minutes(120),
            location: create_test_coords(),
            allowed_distance_meters: 100.0,
            employee_ids: employees.to_vec(),
        })
    }

    #[test]
    fn test_checkin_success() {
        let service = CheckinService::new();
        let emp1 = service.create_employee(CreateEmployeeRequest {
            name: "张三".to_string(),
            employee_number: "001".to_string(),
        });
        let activity = create_test_activity(&service, &[emp1.id], 60);

        let result = service.checkin(
            activity.id,
            CheckinRequest {
                employee_id: emp1.id,
                device_id: "device1".to_string(),
                location: create_test_coords(),
            },
            Utc::now(),
        );

        assert!(result.is_ok());
        let record = result.unwrap();
        assert!(record.checkin_time.is_some());
        assert_eq!(record.status, CheckinStatus::NotCheckedOut);
    }

    #[test]
    fn test_checkin_location_out_of_range() {
        let service = CheckinService::new();
        let emp1 = service.create_employee(CreateEmployeeRequest {
            name: "张三".to_string(),
            employee_number: "001".to_string(),
        });
        let activity = create_test_activity(&service, &[emp1.id], 60);

        let result = service.checkin(
            activity.id,
            CheckinRequest {
                employee_id: emp1.id,
                device_id: "device1".to_string(),
                location: GpsCoordinate {
                    latitude: 31.2304,
                    longitude: 121.4737,
                },
            },
            Utc::now(),
        );

        assert!(matches!(result, Err(CheckinError::LocationOutOfRange)));
    }

    #[test]
    fn test_device_anti_proxy() {
        let service = CheckinService::new();
        let emp1 = service.create_employee(CreateEmployeeRequest {
            name: "张三".to_string(),
            employee_number: "001".to_string(),
        });
        let emp2 = service.create_employee(CreateEmployeeRequest {
            name: "李四".to_string(),
            employee_number: "002".to_string(),
        });
        let activity = create_test_activity(&service, &[emp1.id, emp2.id], 60);

        let result1 = service.checkin(
            activity.id,
            CheckinRequest {
                employee_id: emp1.id,
                device_id: "device1".to_string(),
                location: create_test_coords(),
            },
            Utc::now(),
        );
        assert!(result1.is_ok());

        let result2 = service.checkin(
            activity.id,
            CheckinRequest {
                employee_id: emp2.id,
                device_id: "device1".to_string(),
                location: create_test_coords(),
            },
            Utc::now(),
        );
        assert!(matches!(result2, Err(CheckinError::DeviceAlreadyUsedByOther)));
    }

    #[test]
    fn test_same_employee_different_device() {
        let service = CheckinService::new();
        let emp1 = service.create_employee(CreateEmployeeRequest {
            name: "张三".to_string(),
            employee_number: "001".to_string(),
        });
        let activity = create_test_activity(&service, &[emp1.id], 60);

        let result1 = service.checkin(
            activity.id,
            CheckinRequest {
                employee_id: emp1.id,
                device_id: "device1".to_string(),
                location: create_test_coords(),
            },
            Utc::now(),
        );
        assert!(result1.is_ok());

        let result2 = service.checkin(
            activity.id,
            CheckinRequest {
                employee_id: emp1.id,
                device_id: "device2".to_string(),
                location: create_test_coords(),
            },
            Utc::now(),
        );
        assert!(matches!(result2, Err(CheckinError::AlreadyCheckedIn)));
    }

    #[test]
    fn test_checkout_early() {
        let service = CheckinService::new();
        let emp1 = service.create_employee(CreateEmployeeRequest {
            name: "张三".to_string(),
            employee_number: "001".to_string(),
        });
        let activity = create_test_activity(&service, &[emp1.id], 60);

        service
            .checkin(
                activity.id,
                CheckinRequest {
                    employee_id: emp1.id,
                    device_id: "device1".to_string(),
                    location: create_test_coords(),
                },
                Utc::now(),
            )
            .unwrap();

        let result = service.checkout(
            activity.id,
            CheckoutRequest {
                employee_id: emp1.id,
                location: create_test_coords(),
            },
            Utc::now(),
        );
        assert!(matches!(result, Err(CheckinError::EarlyCheckout)));

        let record = service.get_record(activity.id, emp1.id).unwrap();
        assert_eq!(record.status, CheckinStatus::EarlyCheckout);
    }

    #[test]
    fn test_checkout_normal() {
        let service = CheckinService::new();
        let emp1 = service.create_employee(CreateEmployeeRequest {
            name: "张三".to_string(),
            employee_number: "001".to_string(),
        });
        let activity = create_test_activity(&service, &[emp1.id], 60);

        service
            .checkin(
                activity.id,
                CheckinRequest {
                    employee_id: emp1.id,
                    device_id: "device1".to_string(),
                    location: create_test_coords(),
                },
                Utc::now(),
            )
            .unwrap();

        let late_time = Utc::now() + Duration::minutes(150);
        let result = service.checkout(
            activity.id,
            CheckoutRequest {
                employee_id: emp1.id,
                location: create_test_coords(),
            },
            late_time,
        );
        assert!(result.is_ok());

        let record = result.unwrap();
        assert_eq!(record.status, CheckinStatus::Normal);
    }

    #[test]
    fn test_report() {
        let service = CheckinService::new();
        let emp1 = service.create_employee(CreateEmployeeRequest {
            name: "张三".to_string(),
            employee_number: "001".to_string(),
        });
        let emp2 = service.create_employee(CreateEmployeeRequest {
            name: "李四".to_string(),
            employee_number: "002".to_string(),
        });
        let emp3 = service.create_employee(CreateEmployeeRequest {
            name: "王五".to_string(),
            employee_number: "003".to_string(),
        });
        let activity = create_test_activity(&service, &[emp1.id, emp2.id, emp3.id], 60);

        service
            .checkin(
                activity.id,
                CheckinRequest {
                    employee_id: emp1.id,
                    device_id: "device1".to_string(),
                    location: create_test_coords(),
                },
                Utc::now(),
            )
            .unwrap();

        service
            .checkin(
                activity.id,
                CheckinRequest {
                    employee_id: emp2.id,
                    device_id: "device2".to_string(),
                    location: create_test_coords(),
                },
                Utc::now(),
            )
            .unwrap();

        let late_time = Utc::now() + Duration::minutes(150);
        service
            .checkout(
                activity.id,
                CheckoutRequest {
                    employee_id: emp2.id,
                    location: create_test_coords(),
                },
                late_time,
            )
            .unwrap();

        let report = service.get_report(activity.id).unwrap();
        assert_eq!(report.total_count, 3);
        assert_eq!(report.checked_in_count, 2);
        assert_eq!(report.normal_count, 1);
        assert_eq!(report.not_checked_out_count, 1);
        assert_eq!(report.early_checkout_count, 0);
        assert_eq!(report.not_checked_in_count, 1);
        assert!((report.checkin_rate - 2.0 / 3.0).abs() < 0.001);
    }

    #[test]
    fn test_concurrent_device_check() {
        use std::thread;

        let service = CheckinService::new();
        let emp1 = service.create_employee(CreateEmployeeRequest {
            name: "张三".to_string(),
            employee_number: "001".to_string(),
        });
        let emp2 = service.create_employee(CreateEmployeeRequest {
            name: "李四".to_string(),
            employee_number: "002".to_string(),
        });
        let activity = create_test_activity(&service, &[emp1.id, emp2.id], 60);

        let service1 = service.clone();
        let emp1_id = emp1.id;
        let activity_id = activity.id;
        let handle1 = thread::spawn(move || {
            service1.checkin(
                activity_id,
                CheckinRequest {
                    employee_id: emp1_id,
                    device_id: "device_shared".to_string(),
                    location: create_test_coords(),
                },
                Utc::now(),
            )
        });

        let service2 = service.clone();
        let emp2_id = emp2.id;
        let activity_id = activity.id;
        let handle2 = thread::spawn(move || {
            service2.checkin(
                activity_id,
                CheckinRequest {
                    employee_id: emp2_id,
                    device_id: "device_shared".to_string(),
                    location: create_test_coords(),
                },
                Utc::now(),
            )
        });

        let result1 = handle1.join().unwrap();
        let result2 = handle2.join().unwrap();

        let success_count = [&result1, &result2].iter().filter(|r| r.is_ok()).count();
        let proxy_count = [&result1, &result2]
            .iter()
            .filter(|r| matches!(r, Err(CheckinError::DeviceAlreadyUsedByOther)))
            .count();

        assert_eq!(success_count, 1);
        assert_eq!(proxy_count, 1);
    }
}
