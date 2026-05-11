use chrono::NaiveDate;
use parking_lot::RwLock;
use std::collections::HashMap;
use std::sync::Arc;

use crate::error::{LeaveError, Result};
use crate::models::*;

#[derive(Clone)]
pub struct LeaveService {
    employees: Arc<RwLock<HashMap<String, Employee>>>,
    balances: Arc<RwLock<HashMap<String, LeaveBalance>>>,
    requests: Arc<RwLock<HashMap<String, LeaveRequest>>>,
    employee_requests: Arc<RwLock<HashMap<String, Vec<String>>>>,
}

impl LeaveService {
    pub fn new() -> Self {
        Self {
            employees: Arc::new(RwLock::new(HashMap::new())),
            balances: Arc::new(RwLock::new(HashMap::new())),
            requests: Arc::new(RwLock::new(HashMap::new())),
            employee_requests: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn add_employee(&self, name: &str, hire_date: NaiveDate) -> Employee {
        let employee = Employee::new(name, hire_date);
        let emp_id = employee.id.clone();
        let balance = LeaveBalance::new();

        self.employees.write().insert(emp_id.clone(), employee.clone());
        self.balances.write().insert(emp_id.clone(), balance);
        self.employee_requests.write().insert(emp_id.clone(), Vec::new());

        employee
    }

    pub fn get_employee(&self, employee_id: &str) -> Result<Employee> {
        self.employees
            .read()
            .get(employee_id)
            .cloned()
            .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))
    }

    pub fn list_employees(&self) -> Vec<Employee> {
        self.employees.read().values().cloned().collect()
    }

    pub fn initialize_annual_leave(&self, employee_id: &str, year: i32) -> Result<u32> {
        let employee = self.get_employee(employee_id)?;
        let service_years = calculate_service_years(employee.hire_date, NaiveDate::from_ymd_opt(year, 12, 31).unwrap());
        let entitlement = if service_years >= 1.0 {
            calculate_annual_leave_entitlement(service_years)
        } else {
            calculate_prorated_annual_leave(employee.hire_date, year)
        };

        let pool = AnnualLeavePool::new(year, entitlement, None);

        let mut balances = self.balances.write();
        let balance = balances
            .get_mut(employee_id)
            .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))?;

        balance.annual_leave.retain(|p| p.year != year);
        balance.annual_leave.push(pool);
        balance.annual_leave.sort_by_key(|p| (p.expires_on.is_some(), p.year));

        Ok(entitlement)
    }

    pub fn perform_year_end_carryover(&self, employee_id: &str, from_year: i32) -> Result<u32> {
        let _today = chrono::Local::now().date_naive();
        let mut balances = self.balances.write();
        let balance = balances
            .get_mut(employee_id)
            .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))?;

        let carryover_amount = if let Some(from_pool) = balance
            .annual_leave
            .iter_mut()
            .find(|p| p.year == from_year && p.expires_on.is_none())
        {
            let remaining = from_pool.remaining_days();
            let carryover = std::cmp::min(remaining, max_carryover_days());
            from_pool.used_days = from_pool.total_days;
            carryover
        } else {
            0
        };

        if carryover_amount > 0 {
            let to_year = from_year + 1;
            let expiry = NaiveDate::from_ymd_opt(to_year, 12, 31).unwrap();
            let carryover_pool = AnnualLeavePool::new(from_year, carryover_amount, Some(expiry));
            
            if !balance.annual_leave.iter().any(|p| p.year == to_year) {
                let employee = self.employees.read().get(employee_id).cloned().unwrap();
                let service_years = calculate_service_years(
                    employee.hire_date,
                    NaiveDate::from_ymd_opt(to_year, 12, 31).unwrap(),
                );
                let entitlement = calculate_annual_leave_entitlement(service_years);
                balance.annual_leave.push(AnnualLeavePool::new(to_year, entitlement, None));
            }

            balance.annual_leave.push(carryover_pool);
            balance.annual_leave.sort_by_key(|p| (p.expires_on.is_none(), p.year));
        }

        Ok(carryover_amount)
    }

    pub fn get_available_annual_leave(&self, employee_id: &str, today: NaiveDate) -> Result<u32> {
        let balances = self.balances.read();
        let balance = balances
            .get(employee_id)
            .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))?;

        let total = balance
            .annual_leave
            .iter()
            .filter(|p| !p.is_expired(today))
            .map(|p| p.remaining_days())
            .sum();

        Ok(total)
    }

    pub fn get_leave_balance(&self, employee_id: &str) -> Result<LeaveBalance> {
        self.balances
            .read()
            .get(employee_id)
            .cloned()
            .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))
    }

    pub fn apply_leave(
        &self,
        employee_id: &str,
        leave_type: LeaveType,
        start_date: NaiveDate,
        end_date: NaiveDate,
        sick_leave_proof: Option<String>,
    ) -> Result<LeaveRequest> {
        let today = chrono::Local::now().date_naive();
        self.get_employee(employee_id)?;

        let mut request = LeaveRequest::new(employee_id, leave_type, start_date, end_date, sick_leave_proof)?;

        match leave_type {
            LeaveType::Annual => {
                let usage = self.consume_annual_leave(employee_id, request.total_days, today)?;
                request.annual_leave_usage = usage;
            }
            LeaveType::Personal => {
                let mut balances = self.balances.write();
                let balance = balances
                    .get_mut(employee_id)
                    .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))?;
                if balance.personal_leave < request.total_days {
                    return Err(LeaveError::InsufficientLeaveBalance {
                        need: request.total_days,
                        available: balance.personal_leave,
                    });
                }
                balance.personal_leave -= request.total_days;
            }
            LeaveType::Sick => {
                let mut balances = self.balances.write();
                let balance = balances
                    .get_mut(employee_id)
                    .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))?;
                if balance.sick_leave < request.total_days {
                    return Err(LeaveError::InsufficientLeaveBalance {
                        need: request.total_days,
                        available: balance.sick_leave,
                    });
                }
                balance.sick_leave -= request.total_days;
            }
            LeaveType::Compensatory => {
                let mut balances = self.balances.write();
                let balance = balances
                    .get_mut(employee_id)
                    .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))?;
                if balance.compensatory_leave < request.total_days {
                    return Err(LeaveError::InsufficientLeaveBalance {
                        need: request.total_days,
                        available: balance.compensatory_leave,
                    });
                }
                balance.compensatory_leave -= request.total_days;
            }
        }

        let request_id = request.id.clone();
        let emp_id = employee_id.to_string();

        self.requests.write().insert(request_id.clone(), request.clone());
        self.employee_requests
            .write()
            .entry(emp_id)
            .or_default()
            .push(request_id);

        Ok(request)
    }

    fn consume_annual_leave(
        &self,
        employee_id: &str,
        days: u32,
        today: NaiveDate,
    ) -> Result<Vec<AnnualLeaveUsage>> {
        let mut balances = self.balances.write();
        let balance = balances
            .get_mut(employee_id)
            .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))?;

        let available: u32 = balance
            .annual_leave
            .iter()
            .filter(|p| !p.is_expired(today))
            .map(|p| p.remaining_days())
            .sum();

        if available < days {
            return Err(LeaveError::InsufficientLeaveBalance {
                need: days,
                available,
            });
        }

        let mut pool_indices: Vec<usize> = balance
            .annual_leave
            .iter()
            .enumerate()
            .filter(|(_, p)| !p.is_expired(today) && p.remaining_days() > 0)
            .map(|(i, _)| i)
            .collect();

        pool_indices.sort_by(|a, b| {
            let a_pool = &balance.annual_leave[*a];
            let b_pool = &balance.annual_leave[*b];
            let a_expiry = a_pool.expires_on.unwrap_or(NaiveDate::MAX);
            let b_expiry = b_pool.expires_on.unwrap_or(NaiveDate::MAX);
            a_expiry
                .cmp(&b_expiry)
                .then_with(|| a_pool.year.cmp(&b_pool.year))
        });

        let mut remaining_needed = days;
        let mut usage_records = Vec::new();

        for idx in pool_indices {
            if remaining_needed == 0 {
                break;
            }

            let pool = &mut balance.annual_leave[idx];
            let available_in_pool = pool.remaining_days();
            let consume = std::cmp::min(remaining_needed, available_in_pool);

            pool.used_days += consume;

            usage_records.push(AnnualLeaveUsage {
                year: pool.year,
                is_carryover: pool.expires_on.is_some(),
                days_used: consume,
            });

            remaining_needed -= consume;
        }

        Ok(usage_records)
    }

    pub fn cancel_leave(&self, request_id: &str) -> Result<LeaveRequest> {
        let today = chrono::Local::now().date_naive();
        let mut requests = self.requests.write();
        let request = requests
            .get_mut(request_id)
            .ok_or_else(|| LeaveError::LeaveNotFound(request_id.to_string()))?;

        if request.has_started(today) {
            return Err(LeaveError::LeaveAlreadyStarted);
        }

        if request.status == LeaveStatus::Cancelled {
            return Ok(request.clone());
        }

        let employee_id = request.employee_id.clone();
        let leave_type = request.leave_type;
        let total_days = request.total_days;
        let annual_usage = request.annual_leave_usage.clone();

        request.status = LeaveStatus::Cancelled;

        let mut balances = self.balances.write();
        let balance = balances
            .get_mut(&employee_id)
            .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.clone()))?;

        match leave_type {
            LeaveType::Annual => {
                for usage in annual_usage {
                    if let Some(pool) = balance.annual_leave.iter_mut().find(|p| {
                        p.year == usage.year && p.expires_on.is_some() == usage.is_carryover
                    }) {
                        pool.used_days = pool.used_days.saturating_sub(usage.days_used);
                    }
                }
            }
            LeaveType::Personal => {
                balance.personal_leave += total_days;
            }
            LeaveType::Sick => {
                balance.sick_leave += total_days;
            }
            LeaveType::Compensatory => {
                balance.compensatory_leave += total_days;
            }
        }

        Ok(request.clone())
    }

    pub fn add_leave_balance(
        &self,
        employee_id: &str,
        leave_type: LeaveType,
        days: u32,
    ) -> Result<()> {
        let mut balances = self.balances.write();
        let balance = balances
            .get_mut(employee_id)
            .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))?;

        match leave_type {
            LeaveType::Personal => {
                balance.personal_leave += days;
            }
            LeaveType::Sick => {
                balance.sick_leave += days;
            }
            LeaveType::Compensatory => {
                balance.compensatory_leave += days;
            }
            LeaveType::Annual => {
                return Err(LeaveError::AnnualLeaveCalculationError(
                    "年假应通过 initialize_annual_leave 或 perform_year_end_carryover 管理".to_string(),
                ));
            }
        }

        Ok(())
    }

    pub fn get_leave_request(&self, request_id: &str) -> Result<LeaveRequest> {
        self.requests
            .read()
            .get(request_id)
            .cloned()
            .ok_or_else(|| LeaveError::LeaveNotFound(request_id.to_string()))
    }

    pub fn list_employee_leaves(&self, employee_id: &str) -> Result<Vec<LeaveRequest>> {
        let employee_requests = self.employee_requests.read();
        let request_ids = employee_requests
            .get(employee_id)
            .ok_or_else(|| LeaveError::EmployeeNotFound(employee_id.to_string()))?;

        let requests = self.requests.read();
        let mut leaves = Vec::new();
        for id in request_ids {
            if let Some(req) = requests.get(id) {
                leaves.push(req.clone());
            }
        }

        leaves.sort_by(|a, b| b.start_date.cmp(&a.start_date));
        Ok(leaves)
    }

    pub fn list_all_leaves(&self) -> Vec<LeaveRequest> {
        self.requests.read().values().cloned().collect()
    }
}

impl Default for LeaveService {
    fn default() -> Self {
        Self::new()
    }
}
