use std::collections::HashMap;
use parking_lot::Mutex;
use uuid::Uuid;
use chrono::{DateTime, Utc};

use crate::models::{Employee, Overtime, Leave, Holiday, LeaveBalance};

#[derive(Default)]
pub struct InMemoryStore {
    employees: Mutex<HashMap<Uuid, Employee>>,
    overtimes: Mutex<HashMap<Uuid, Overtime>>,
    leaves: Mutex<HashMap<Uuid, Leave>>,
    holidays: Mutex<HashMap<Uuid, Holiday>>,
    leave_balances: Mutex<HashMap<Uuid, LeaveBalance>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn create_employee(&self, employee: Employee) -> Employee {
        let id = employee.id;
        self.employees.lock().insert(id, employee.clone());
        self.leave_balances.lock().insert(id, LeaveBalance {
            employee_id: id,
            balance_hours: 0.0,
        });
        employee
    }

    pub fn get_employee(&self, id: Uuid) -> Option<Employee> {
        self.employees.lock().get(&id).cloned()
    }

    pub fn list_employees(&self) -> Vec<Employee> {
        self.employees.lock().values().cloned().collect()
    }

    pub fn create_overtime(&self, overtime: Overtime) -> Overtime {
        self.overtimes.lock().insert(overtime.id, overtime.clone());
        overtime
    }

    pub fn get_overtime(&self, id: Uuid) -> Option<Overtime> {
        self.overtimes.lock().get(&id).cloned()
    }

    pub fn update_overtime(&self, overtime: Overtime) -> Overtime {
        self.overtimes.lock().insert(overtime.id, overtime.clone());
        overtime
    }

    pub fn list_overtimes(&self, employee_id: Option<Uuid>) -> Vec<Overtime> {
        let overtimes = self.overtimes.lock();
        overtimes.values()
            .filter(|o| employee_id.map_or(true, |id| o.employee_id == id))
            .cloned()
            .collect()
    }

    pub fn create_leave(&self, leave: Leave) -> Leave {
        self.leaves.lock().insert(leave.id, leave.clone());
        leave
    }

    pub fn get_leave(&self, id: Uuid) -> Option<Leave> {
        self.leaves.lock().get(&id).cloned()
    }

    pub fn update_leave(&self, leave: Leave) -> Leave {
        self.leaves.lock().insert(leave.id, leave.clone());
        leave
    }

    pub fn list_leaves(&self, employee_id: Option<Uuid>) -> Vec<Leave> {
        let leaves = self.leaves.lock();
        leaves.values()
            .filter(|l| employee_id.map_or(true, |id| l.employee_id == id))
            .cloned()
            .collect()
    }

    pub fn create_holiday(&self, holiday: Holiday) -> Holiday {
        self.holidays.lock().insert(holiday.id, holiday.clone());
        holiday
    }

    pub fn list_holidays(&self) -> Vec<Holiday> {
        self.holidays.lock().values().cloned().collect()
    }

    pub fn is_holiday(&self, date: DateTime<Utc>) -> bool {
        let holidays = self.holidays.lock();
        holidays.values().any(|h| {
            h.date.date_naive() == date.date_naive()
        })
    }

    pub fn get_leave_balance(&self, employee_id: Uuid) -> f64 {
        self.leave_balances.lock()
            .get(&employee_id)
            .map(|b| b.balance_hours)
            .unwrap_or(0.0)
    }

    pub fn update_leave_balance(&self, employee_id: Uuid, delta: f64) -> f64 {
        let mut balances = self.leave_balances.lock();
        let balance = balances.entry(employee_id).or_insert(LeaveBalance {
            employee_id,
            balance_hours: 0.0,
        });
        balance.balance_hours += delta;
        balance.balance_hours
    }
}
