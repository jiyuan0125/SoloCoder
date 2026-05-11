use chrono::Utc;
use uuid::Uuid;
use crate::models::*;
use crate::store::InMemoryStore;

#[derive(Clone)]
pub struct CertificationService {
    store: InMemoryStore,
}

impl CertificationService {
    pub fn new(store: InMemoryStore) -> Self {
        Self { store }
    }

    pub fn create_department(&self, name: &str) -> Result<Department, CertError> {
        let dept = Department::new(name);
        self.store.add_department(dept)
    }

    pub fn list_departments(&self) -> Vec<Department> {
        self.store.list_departments()
    }

    pub fn get_department(&self, id: &Uuid) -> Option<Department> {
        self.store.get_department(id)
    }

    pub fn create_employee(&self, name: &str, department_id: Uuid) -> Result<Employee, CertError> {
        let employee = Employee::new(name, department_id);
        self.store.add_employee(employee)
    }

    pub fn list_employees(&self) -> Vec<Employee> {
        self.store.list_employees()
    }

    pub fn get_employee(&self, id: &Uuid) -> Option<Employee> {
        self.store.get_employee(id)
    }

    pub fn create_certificate(
        &self,
        employee_id: Uuid,
        certificate_number: &str,
        issuer: &str,
        issue_date: chrono::NaiveDate,
        expiry_date: chrono::NaiveDate,
        certificate_type: &str,
        required_hours_for_renewal: u32,
    ) -> Result<Certificate, CertError> {
        let cert = Certificate::new(
            employee_id,
            certificate_number,
            issuer,
            issue_date,
            expiry_date,
            certificate_type,
            required_hours_for_renewal,
        );
        self.store.add_certificate(cert)
    }

    pub fn list_certificates(&self) -> Vec<Certificate> {
        self.store.list_certificates()
    }

    pub fn get_certificate(&self, id: &Uuid) -> Option<Certificate> {
        self.store.get_certificate(id)
    }

    pub fn get_employee_certificates(&self, employee_id: &Uuid) -> Vec<Certificate> {
        self.store.get_certificates_by_employee(employee_id)
    }

    pub fn revoke_certificate(&self, certificate_id: Uuid) -> Result<Certificate, CertError> {
        self.store.update_certificate(&certificate_id, |cert| {
            cert.revoke();
        })
    }

    pub fn calculate_employee_training_hours(&self, employee_id: &Uuid) -> u32 {
        let records = self.store.list_training_records();
        records.iter()
            .filter(|r| r.attendee_ids.contains(employee_id))
            .map(|r| r.hours)
            .sum()
    }

    pub fn renew_certificate(
        &self,
        certificate_id: Uuid,
        new_expiry_date: chrono::NaiveDate,
    ) -> Result<CertificateRenewalResult, CertError> {
        let cert = self.store.get_certificate(&certificate_id)
            .ok_or(CertError::CertificateNotFound)?;
        let total_hours = self.calculate_employee_training_hours(&cert.employee_id);
        let result = self.store.update_certificate(&certificate_id, |cert| {
            let _ = cert.renew(new_expiry_date, total_hours);
        })?;
        let hours_insufficient = total_hours < result.required_hours_for_renewal;
        if result.status == CertificateStatus::Revoked {
            Ok(CertificateRenewalResult::CannotRenewRevoked)
        } else if hours_insufficient {
            Ok(CertificateRenewalResult::RenewedWithInsufficientHours {
                actual_hours: total_hours,
                required_hours: result.required_hours_for_renewal,
            })
        } else {
            Ok(CertificateRenewalResult::RenewedSuccessfully)
        }
    }

    pub fn create_training_record(
        &self,
        name: &str,
        date: chrono::NaiveDate,
        hours: u32,
        attendee_ids: Vec<Uuid>,
    ) -> Result<TrainingRecord, CertError> {
        let record = TrainingRecord::new(name, date, hours, attendee_ids);
        self.store.add_training_record(record)
    }

    pub fn list_training_records(&self) -> Vec<TrainingRecord> {
        self.store.list_training_records()
    }

    pub fn get_training_record(&self, id: &Uuid) -> Option<TrainingRecord> {
        self.store.get_training_record(id)
    }

    pub fn get_department_certification_rates(&self) -> Vec<DepartmentCertificationRate> {
        let departments = self.list_departments();
        let employees = self.list_employees();
        let certificates = self.list_certificates();
        
        departments.iter().map(|dept| {
            let dept_employees: Vec<_> = employees.iter()
                .filter(|e| e.department_id == dept.id)
                .collect();
            let total = dept_employees.len();
            
            let certified_count = dept_employees.iter()
                .filter(|e| {
                    certificates.iter().any(|c| {
                        c.employee_id == e.id && 
                        c.status != CertificateStatus::Revoked
                    })
                })
                .count();
            
            let rate = if total > 0 {
                certified_count as f64 / total as f64
            } else {
                0.0
            };
            
            DepartmentCertificationRate {
                department_id: dept.id,
                department_name: dept.name.clone(),
                total_employees: total,
                certified_employees: certified_count,
                certification_rate: rate,
            }
        }).collect()
    }

    pub fn get_certificate_type_expiry_rates(&self) -> Vec<CertificateTypeExpiryRate> {
        let certificates = self.list_certificates();
        let mut type_map: std::collections::HashMap<String, (usize, usize)> = std::collections::HashMap::new();
        
        for cert in &certificates {
            let entry = type_map.entry(cert.certificate_type.clone()).or_insert((0, 0));
            entry.0 += 1;
            if cert.status == CertificateStatus::Expired {
                entry.1 += 1;
            }
        }
        
        type_map.into_iter().map(|(cert_type, (total, expired))| {
            let rate = if total > 0 {
                expired as f64 / total as f64
            } else {
                0.0
            };
            CertificateTypeExpiryRate {
                certificate_type: cert_type,
                total_certificates: total,
                expired_count: expired,
                expiry_rate: rate,
            }
        }).collect()
    }

    pub fn get_expiring_certificates_alerts(&self) -> Vec<ExpiringCertificatesAlert> {
        let today = Utc::now().date_naive();
        let employees = self.list_employees();
        let certificates = self.list_certificates();
        
        let mut employee_certs: std::collections::HashMap<Uuid, Vec<Certificate>> = std::collections::HashMap::new();
        
        for cert in certificates {
            let days_until = (cert.expiry_date - today).num_days();
            if (cert.status == CertificateStatus::ExpiringSoon || cert.status == CertificateStatus::Expired) 
                && cert.status != CertificateStatus::Revoked {
                employee_certs.entry(cert.employee_id)
                    .or_default()
                    .push(cert);
            }
        }
        
        employee_certs.into_iter().filter_map(|(emp_id, certs)| {
            employees.iter().find(|e| e.id == emp_id).map(|emp| {
                ExpiringCertificatesAlert {
                    employee_id: emp.id,
                    employee_name: emp.name.clone(),
                    certificates: certs,
                }
            })
        }).collect()
    }
}
