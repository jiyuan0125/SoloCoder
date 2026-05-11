use chrono::Utc;
use uuid::Uuid;

use crate::error::AppError;
use crate::models::*;
use crate::store::{StoreRef, InMemoryStore};

pub struct TalentService {
    store: StoreRef,
}

impl TalentService {
    pub fn new(store: StoreRef) -> Self {
        Self { store }
    }
    
    pub async fn create_department(&self, name: String) -> Department {
        let dept = Department::new(name);
        let mut store = self.store.write().await;
        store.add_department(dept.clone());
        dept
    }
    
    pub async fn get_departments(&self) -> Vec<Department> {
        let store = self.store.read().await;
        store.get_all_departments()
    }
    
    pub async fn create_employee(
        &self,
        name: String,
        employee_number: String,
        department_id: Uuid,
        position: String,
    ) -> Result<Employee, AppError> {
        let emp = Employee::new(name, employee_number, department_id, position);
        let mut store = self.store.write().await;
        store.add_employee(emp.clone())?;
        Ok(emp)
    }
    
    pub async fn get_employees(&self, department_id: Option<Uuid>) -> Vec<Employee> {
        let store = self.store.read().await;
        match department_id {
            Some(id) => store.get_employees_by_department(&id),
            None => store.get_all_employees(),
        }
    }
    
    pub async fn create_assessment(
        &self,
        employee_id: Uuid,
        year: i32,
        performance_value: u8,
        potential_value: u8,
    ) -> Result<Assessment, AppError> {
        let performance = PerformanceLevel::from_value(performance_value)
            .ok_or(AppError::InvalidRating)?;
        let potential = PotentialLevel::from_value(potential_value)
            .ok_or(AppError::InvalidRating)?;
        
        let mut store = self.store.write().await;
        store.add_assessment(employee_id, year, performance, potential)
    }
    
    pub async fn get_assessments(&self, employee_id: Uuid) -> Vec<Assessment> {
        let store = self.store.read().await;
        store.get_assessments_by_employee(&employee_id)
    }
    
    pub async fn create_succession_plan(
        &self,
        position: String,
        department_id: Uuid,
        is_key_position: bool,
    ) -> Result<SuccessionPlan, AppError> {
        let plan = SuccessionPlan::new(position, department_id, is_key_position);
        let mut store = self.store.write().await;
        store.add_succession_plan(plan.clone())?;
        Ok(plan)
    }
    
    pub async fn get_succession_plans(&self, department_id: Option<Uuid>) -> Vec<SuccessionPlan> {
        let store = self.store.read().await;
        match department_id {
            Some(id) => store.get_succession_plans_by_department(&id),
            None => store.get_all_succession_plans(),
        }
    }
    
    pub async fn add_candidate_to_succession_plan(
        &self,
        plan_id: Uuid,
        employee_id: Uuid,
        assessment_year: Option<i32>,
        notes: Option<String>,
    ) -> Result<SuccessionPlan, AppError> {
        let mut store = self.store.write().await;
        
        let mut plan = store.get_succession_plan(&plan_id)
            .ok_or_else(|| AppError::PositionNotFound(plan_id.to_string()))?;
        
        let assessment = store.get_latest_assessment(&employee_id, assessment_year)
            .ok_or_else(|| AppError::EmployeeNotFound(employee_id.to_string()))?;
        
        let candidate = Candidate {
            employee_id,
            current_zone: assessment.zone,
            assessment_year: assessment.year,
            notes,
        };
        
        plan.candidates.push(candidate);
        plan.updated_at = Utc::now();
        
        store.update_succession_plan(plan.clone())?;
        Ok(plan)
    }
    
    pub async fn get_grid_distribution(
        &self,
        year: i32,
        department_id: Option<Uuid>,
    ) -> GridDistribution {
        let store = self.store.read().await;
        store.get_grid_distribution(year, department_id)
    }
    
    pub async fn get_employees_by_zone(
        &self,
        year: i32,
        zone: GridZone,
        department_id: Option<Uuid>,
    ) -> Vec<(Employee, Assessment)> {
        let store = self.store.read().await;
        store.get_employees_by_zone(year, zone, department_id)
    }
    
    pub async fn get_succession_warnings(&self) -> Vec<String> {
        let store = self.store.read().await;
        store.get_succession_warnings()
    }
}

impl Default for TalentService {
    fn default() -> Self {
        Self::new(InMemoryStore::new_ref())
    }
}
