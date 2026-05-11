use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::error::AppError;
use crate::models::*;

pub type StoreRef = Arc<RwLock<InMemoryStore>>;

#[derive(Debug, Default)]
pub struct InMemoryStore {
    departments: HashMap<Uuid, Department>,
    employees: HashMap<Uuid, Employee>,
    assessments: HashMap<Uuid, Assessment>,
    succession_plans: HashMap<Uuid, SuccessionPlan>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }
    
    pub fn new_ref() -> StoreRef {
        Arc::new(RwLock::new(Self::new()))
    }
    
    pub fn add_department(&mut self, dept: Department) {
        self.departments.insert(dept.id, dept);
    }
    
    pub fn get_department(&self, id: &Uuid) -> Option<Department> {
        self.departments.get(id).cloned()
    }
    
    pub fn get_all_departments(&self) -> Vec<Department> {
        self.departments.values().cloned().collect()
    }
    
    pub fn add_employee(&mut self, emp: Employee) -> Result<(), AppError> {
        if !self.departments.contains_key(&emp.department_id) {
            return Err(AppError::DepartmentNotFound(emp.department_id.to_string()));
        }
        self.employees.insert(emp.id, emp);
        Ok(())
    }
    
    pub fn get_employee(&self, id: &Uuid) -> Option<Employee> {
        self.employees.get(id).cloned()
    }
    
    pub fn get_employees_by_department(&self, dept_id: &Uuid) -> Vec<Employee> {
        self.employees
            .values()
            .filter(|e| e.department_id == *dept_id)
            .cloned()
            .collect()
    }
    
    pub fn get_all_employees(&self) -> Vec<Employee> {
        self.employees.values().cloned().collect()
    }
    
    pub fn add_assessment(
        &mut self,
        emp_id: Uuid,
        year: i32,
        performance: PerformanceLevel,
        potential: PotentialLevel,
    ) -> Result<Assessment, AppError> {
        if !self.employees.contains_key(&emp_id) {
            return Err(AppError::EmployeeNotFound(emp_id.to_string()));
        }
        
        let exists = self.assessments.values().any(|a| a.employee_id == emp_id && a.year == year);
        if exists {
            return Err(AppError::AssessmentAlreadyExistsForYear(year));
        }
        
        let assessment = Assessment::new(emp_id, year, performance, potential);
        self.assessments.insert(assessment.id, assessment.clone());
        Ok(assessment)
    }
    
    pub fn get_assessments_by_employee(&self, emp_id: &Uuid) -> Vec<Assessment> {
        let mut result: Vec<_> = self.assessments
            .values()
            .filter(|a| a.employee_id == *emp_id)
            .cloned()
            .collect();
        result.sort_by(|a, b| b.year.cmp(&a.year));
        result
    }
    
    pub fn get_latest_assessment(&self, emp_id: &Uuid, year: Option<i32>) -> Option<Assessment> {
        let assessments = self.get_assessments_by_employee(emp_id);
        match year {
            Some(y) => assessments.into_iter().find(|a| a.year == y),
            None => assessments.into_iter().next(),
        }
    }
    
    pub fn add_succession_plan(&mut self, plan: SuccessionPlan) -> Result<(), AppError> {
        if !self.departments.contains_key(&plan.department_id) {
            return Err(AppError::DepartmentNotFound(plan.department_id.to_string()));
        }
        self.succession_plans.insert(plan.id, plan);
        Ok(())
    }
    
    pub fn get_succession_plan(&self, id: &Uuid) -> Option<SuccessionPlan> {
        self.succession_plans.get(id).cloned()
    }
    
    pub fn get_succession_plans_by_department(&self, dept_id: &Uuid) -> Vec<SuccessionPlan> {
        self.succession_plans
            .values()
            .filter(|p| p.department_id == *dept_id)
            .cloned()
            .collect()
    }
    
    pub fn get_all_succession_plans(&self) -> Vec<SuccessionPlan> {
        self.succession_plans.values().cloned().collect()
    }
    
    pub fn update_succession_plan(&mut self, plan: SuccessionPlan) -> Result<(), AppError> {
        if !self.succession_plans.contains_key(&plan.id) {
            return Err(AppError::PositionNotFound(plan.id.to_string()));
        }
        self.succession_plans.insert(plan.id, plan);
        Ok(())
    }
    
    pub fn get_grid_distribution(&self, year: i32, dept_id: Option<Uuid>) -> GridDistribution {
        let mut counts = GridCounts::default();
        
        for assessment in self.assessments.values().filter(|a| a.year == year) {
            if let Some(dept) = dept_id {
                if let Some(emp) = self.employees.get(&assessment.employee_id) {
                    if emp.department_id == dept {
                        counts.increment(assessment.zone);
                    }
                }
            } else {
                counts.increment(assessment.zone);
            }
        }
        
        GridDistribution {
            year,
            department_id: dept_id,
            counts,
        }
    }
    
    pub fn get_employees_by_zone(&self, year: i32, zone: GridZone, dept_id: Option<Uuid>) -> Vec<(Employee, Assessment)> {
        let mut result = Vec::new();
        
        for assessment in self.assessments.values().filter(|a| a.year == year && a.zone == zone) {
            if let Some(emp) = self.employees.get(&assessment.employee_id) {
                if let Some(dept) = dept_id {
                    if emp.department_id == dept {
                        result.push((emp.clone(), assessment.clone()));
                    }
                } else {
                    result.push((emp.clone(), assessment.clone()));
                }
            }
        }
        
        result
    }
    
    pub fn get_succession_warnings(&self) -> Vec<String> {
        self.succession_plans
            .values()
            .filter_map(|p| p.warning_message())
            .collect()
    }
}
