use std::sync::Arc;
use uuid::Uuid;
use chrono::Utc;

use crate::models::*;
use crate::store::InMemoryStore;
use crate::utils::*;
use crate::errors::SystemError;

pub struct OvertimeLeaveService {
    store: Arc<InMemoryStore>,
}

impl OvertimeLeaveService {
    pub fn new(store: Arc<InMemoryStore>) -> Self {
        Self { store }
    }

    pub fn create_employee(&self, req: CreateEmployeeRequest) -> Employee {
        let employee = Employee {
            id: Uuid::new_v4(),
            name: req.name,
            monthly_salary: req.monthly_salary,
        };
        self.store.create_employee(employee)
    }

    pub fn get_employee(&self, id: Uuid) -> Result<Employee, SystemError> {
        self.store.get_employee(id).ok_or(SystemError::EmployeeNotFound(id))
    }

    pub fn list_employees(&self) -> Vec<Employee> {
        self.store.list_employees()
    }

    pub fn create_overtime(&self, req: CreateOvertimeRequest) -> Result<Overtime, SystemError> {
        if req.start_time >= req.end_time {
            return Err(SystemError::InvalidOvertimePeriod);
        }

        let employee = self.store.get_employee(req.employee_id)
            .ok_or(SystemError::EmployeeNotFound(req.employee_id))?;

        let submitted_at = Utc::now();
        
        if is_overtime_submission_expired(req.end_time, submitted_at) {
            return Err(SystemError::OvertimeExpired(Uuid::new_v4()));
        }

        let is_holiday = self.store.is_holiday(req.start_time);
        let overtime_type = determine_overtime_type(req.start_time, req.end_time, is_holiday);

        if overtime_type == OvertimeType::Weekend && req.weekend_compensation.is_none() {
            return Err(SystemError::WeekendCompensationRequired);
        }

        if overtime_type == OvertimeType::Holiday {
            if req.weekend_compensation.is_some() {
                return Err(SystemError::HolidayCannotBeLeave);
            }
        }

        let duration_hours = calculate_duration(req.start_time, req.end_time)?;

        let hourly_wage = calculate_hourly_wage(employee.monthly_salary);
        let overtime_pay = if overtime_type == OvertimeType::Holiday || 
            (overtime_type == OvertimeType::Weekend && 
             req.weekend_compensation == Some(WeekendCompensation::OvertimePay)) {
            Some(calculate_overtime_pay(hourly_wage, overtime_type, duration_hours))
        } else {
            None
        };

        let overtime = Overtime {
            id: Uuid::new_v4(),
            employee_id: req.employee_id,
            start_time: req.start_time,
            end_time: req.end_time,
            overtime_type,
            weekend_compensation: req.weekend_compensation,
            status: OvertimeStatus::Pending,
            submitted_at,
            duration_hours,
            overtime_pay,
        };

        Ok(self.store.create_overtime(overtime))
    }

    pub fn approve_overtime(&self, req: ApproveOvertimeRequest) -> Result<Overtime, SystemError> {
        let mut overtime = self.store.get_overtime(req.overtime_id)
            .ok_or(SystemError::OvertimeNotFound(req.overtime_id))?;

        if overtime.status != OvertimeStatus::Pending {
            return Err(SystemError::OvertimeAlreadyProcessed(req.overtime_id));
        }

        if is_overtime_submission_expired(overtime.end_time, overtime.submitted_at) {
            overtime.status = OvertimeStatus::Expired;
            return Ok(self.store.update_overtime(overtime));
        }

        overtime.status = OvertimeStatus::Approved;

        if overtime.overtime_type == OvertimeType::Workday ||
            (overtime.overtime_type == OvertimeType::Weekend && 
             overtime.weekend_compensation == Some(WeekendCompensation::Leave)) {
            self.store.update_leave_balance(overtime.employee_id, overtime.duration_hours);
        }

        Ok(self.store.update_overtime(overtime))
    }

    pub fn reject_overtime(&self, req: RejectOvertimeRequest) -> Result<Overtime, SystemError> {
        let mut overtime = self.store.get_overtime(req.overtime_id)
            .ok_or(SystemError::OvertimeNotFound(req.overtime_id))?;

        if overtime.status != OvertimeStatus::Pending {
            return Err(SystemError::OvertimeAlreadyProcessed(req.overtime_id));
        }

        overtime.status = OvertimeStatus::Rejected;
        Ok(self.store.update_overtime(overtime))
    }

    pub fn list_overtimes(&self, employee_id: Option<Uuid>) -> Vec<Overtime> {
        self.store.list_overtimes(employee_id)
    }

    pub fn create_leave(&self, req: CreateLeaveRequest) -> Result<Leave, SystemError> {
        if req.start_time >= req.end_time {
            return Err(SystemError::InvalidLeavePeriod);
        }

        let _employee = self.store.get_employee(req.employee_id)
            .ok_or(SystemError::EmployeeNotFound(req.employee_id))?;

        let duration_hours = calculate_duration(req.start_time, req.end_time)?;
        let current_balance = self.store.get_leave_balance(req.employee_id);

        if current_balance < duration_hours {
            return Err(SystemError::InsufficientLeaveBalance(req.employee_id, duration_hours, current_balance));
        }

        let leave = Leave {
            id: Uuid::new_v4(),
            employee_id: req.employee_id,
            start_time: req.start_time,
            end_time: req.end_time,
            duration_hours,
            status: LeaveStatus::Pending,
            submitted_at: Utc::now(),
        };

        Ok(self.store.create_leave(leave))
    }

    pub fn approve_leave(&self, req: ApproveLeaveRequest) -> Result<Leave, SystemError> {
        let mut leave = self.store.get_leave(req.leave_id)
            .ok_or(SystemError::LeaveNotFound(req.leave_id))?;

        if leave.status != LeaveStatus::Pending {
            return Err(SystemError::LeaveAlreadyProcessed(req.leave_id));
        }

        let current_balance = self.store.get_leave_balance(leave.employee_id);
        if current_balance < leave.duration_hours {
            return Err(SystemError::InsufficientLeaveBalance(leave.employee_id, leave.duration_hours, current_balance));
        }

        leave.status = LeaveStatus::Approved;
        self.store.update_leave_balance(leave.employee_id, -leave.duration_hours);

        Ok(self.store.update_leave(leave))
    }

    pub fn reject_leave(&self, req: RejectLeaveRequest) -> Result<Leave, SystemError> {
        let mut leave = self.store.get_leave(req.leave_id)
            .ok_or(SystemError::LeaveNotFound(req.leave_id))?;

        if leave.status != LeaveStatus::Pending {
            return Err(SystemError::LeaveAlreadyProcessed(req.leave_id));
        }

        leave.status = LeaveStatus::Rejected;
        Ok(self.store.update_leave(leave))
    }

    pub fn cancel_leave(&self, req: CancelLeaveRequest) -> Result<Leave, SystemError> {
        let mut leave = self.store.get_leave(req.leave_id)
            .ok_or(SystemError::LeaveNotFound(req.leave_id))?;

        if leave.status != LeaveStatus::Approved {
            return Err(SystemError::LeaveAlreadyProcessed(req.leave_id));
        }

        if is_leave_past(leave.end_time) {
            return Err(SystemError::CannotCancelPastLeave(req.leave_id));
        }

        leave.status = LeaveStatus::Cancelled;
        self.store.update_leave_balance(leave.employee_id, leave.duration_hours);

        Ok(self.store.update_leave(leave))
    }

    pub fn list_leaves(&self, employee_id: Option<Uuid>) -> Vec<Leave> {
        self.store.list_leaves(employee_id)
    }

    pub fn create_holiday(&self, req: CreateHolidayRequest) -> Holiday {
        let holiday = Holiday {
            id: Uuid::new_v4(),
            date: req.date,
            name: req.name,
        };
        self.store.create_holiday(holiday)
    }

    pub fn list_holidays(&self) -> Vec<Holiday> {
        self.store.list_holidays()
    }

    pub fn get_leave_balance(&self, employee_id: Uuid) -> Result<f64, SystemError> {
        let _employee = self.store.get_employee(employee_id)
            .ok_or(SystemError::EmployeeNotFound(employee_id))?;
        Ok(self.store.get_leave_balance(employee_id))
    }
}
