use std::collections::HashMap;
use std::sync::{Arc, RwLock};
use chrono::Utc;
use uuid::Uuid;
use crate::models::*;

#[derive(Default, Clone)]
pub struct InMemoryStore {
    inner: Arc<RwLock<StoreInner>>,
}

#[derive(Default)]
struct StoreInner {
    departments: HashMap<Uuid, Department>,
    employees: HashMap<Uuid, Employee>,
    certificates: HashMap<Uuid, Certificate>,
    training_records: HashMap<Uuid, TrainingRecord>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn list_departments(&self) -> Vec<Department> {
        let inner = self.inner.read().unwrap();
        inner.departments.values().cloned().collect()
    }

    pub fn get_department(&self, id: &Uuid) -> Option<Department> {
        let inner = self.inner.read().unwrap();
        inner.departments.get(id).cloned()
    }

    pub fn get_department_by_name(&self, name: &str) -> Option<Department> {
        let inner = self.inner.read().unwrap();
        inner.departments.values()
            .find(|d| d.name == name)
            .cloned()
    }

    pub fn add_department(&self, dept: Department) -> Result<Department, CertError> {
        let mut inner = self.inner.write().unwrap();
        if inner.departments.values().any(|d| d.name == dept.name) {
            return Err(CertError::DepartmentNameExists);
        }
        inner.departments.insert(dept.id, dept.clone());
        Ok(dept)
    }

    pub fn list_employees(&self) -> Vec<Employee> {
        let inner = self.inner.read().unwrap();
        inner.employees.values().cloned().collect()
    }

    pub fn get_employee(&self, id: &Uuid) -> Option<Employee> {
        let inner = self.inner.read().unwrap();
        inner.employees.get(id).cloned()
    }

    pub fn get_employees_by_department(&self, dept_id: &Uuid) -> Vec<Employee> {
        let inner = self.inner.read().unwrap();
        inner.employees.values()
            .filter(|e| e.department_id == *dept_id)
            .cloned()
            .collect()
    }

    pub fn add_employee(&self, employee: Employee) -> Result<Employee, CertError> {
        let mut inner = self.inner.write().unwrap();
        if !inner.departments.contains_key(&employee.department_id) {
            return Err(CertError::DepartmentNotFound);
        }
        inner.employees.insert(employee.id, employee.clone());
        Ok(employee)
    }

    pub fn list_certificates(&self) -> Vec<Certificate> {
        let inner = self.inner.read().unwrap();
        let today = Utc::now().date_naive();
        inner.certificates.values().map(|c| {
            let mut cert = c.clone();
            cert.update_status(&today);
            cert
        }).collect()
    }

    pub fn get_certificate(&self, id: &Uuid) -> Option<Certificate> {
        let inner = self.inner.read().unwrap();
        let today = Utc::now().date_naive();
        inner.certificates.get(id).map(|c| {
            let mut cert = c.clone();
            cert.update_status(&today);
            cert
        })
    }

    pub fn get_certificates_by_employee(&self, employee_id: &Uuid) -> Vec<Certificate> {
        let inner = self.inner.read().unwrap();
        let today = Utc::now().date_naive();
        inner.certificates.values()
            .filter(|c| c.employee_id == *employee_id)
            .map(|c| {
                let mut cert = c.clone();
                cert.update_status(&today);
                cert
            })
            .collect()
    }

    pub fn add_certificate(&self, cert: Certificate) -> Result<Certificate, CertError> {
        let mut inner = self.inner.write().unwrap();
        if !inner.employees.contains_key(&cert.employee_id) {
            return Err(CertError::EmployeeNotFound);
        }
        if inner.certificates.values().any(|c| c.certificate_number == cert.certificate_number) {
            return Err(CertError::CertificateNumberExists);
        }
        inner.certificates.insert(cert.id, cert.clone());
        Ok(cert)
    }

    pub fn update_certificate<F>(&self, id: &Uuid, mut update_fn: F) -> Result<Certificate, CertError>
    where
        F: FnMut(&mut Certificate) -> (),
    {
        let mut inner = self.inner.write().unwrap();
        let cert = inner.certificates.get_mut(id)
            .ok_or(CertError::CertificateNotFound)?;
        update_fn(cert);
        Ok(cert.clone())
    }

    pub fn list_training_records(&self) -> Vec<TrainingRecord> {
        let inner = self.inner.read().unwrap();
        inner.training_records.values().cloned().collect()
    }

    pub fn get_training_record(&self, id: &Uuid) -> Option<TrainingRecord> {
        let inner = self.inner.read().unwrap();
        inner.training_records.get(id).cloned()
    }

    pub fn add_training_record(&self, record: TrainingRecord) -> Result<TrainingRecord, CertError> {
        let mut inner = self.inner.write().unwrap();
        for attendee_id in &record.attendee_ids {
            if !inner.employees.contains_key(attendee_id) {
                return Err(CertError::EmployeeNotFound);
            }
        }
        inner.training_records.insert(record.id, record.clone());
        Ok(record)
    }
}
