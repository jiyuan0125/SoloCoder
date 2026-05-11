use chrono::{DateTime, Utc, NaiveDate, Datelike};
use serde::{Serialize, Deserialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum CertificateStatus {
    Valid,
    ExpiringSoon,
    Expired,
    Revoked,
}

impl CertificateStatus {
    pub fn to_string(&self) -> &'static str {
        match self {
            CertificateStatus::Valid => "有效",
            CertificateStatus::ExpiringSoon => "即将到期",
            CertificateStatus::Expired => "已过期",
            CertificateStatus::Revoked => "已注销",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Department {
    pub id: Uuid,
    pub name: String,
    pub created_at: DateTime<Utc>,
}

impl Department {
    pub fn new(name: impl Into<String>) -> Self {
        Self {
            id: Uuid::new_v4(),
            name: name.into(),
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Employee {
    pub id: Uuid,
    pub name: String,
    pub department_id: Uuid,
    pub created_at: DateTime<Utc>,
}

impl Employee {
    pub fn new(name: impl Into<String>, department_id: Uuid) -> Self {
        Self {
            id: Uuid::new_v4(),
            name: name.into(),
            department_id,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Certificate {
    pub id: Uuid,
    pub employee_id: Uuid,
    pub certificate_number: String,
    pub issuer: String,
    pub issue_date: NaiveDate,
    pub expiry_date: NaiveDate,
    pub certificate_type: String,
    pub required_hours_for_renewal: u32,
    pub hours_insufficient_flag: bool,
    pub status: CertificateStatus,
    pub created_at: DateTime<Utc>,
}

impl Certificate {
    pub fn new(
        employee_id: Uuid,
        certificate_number: impl Into<String>,
        issuer: impl Into<String>,
        issue_date: NaiveDate,
        expiry_date: NaiveDate,
        certificate_type: impl Into<String>,
        required_hours_for_renewal: u32,
    ) -> Self {
        let status = Self::calculate_status(&expiry_date, &Utc::now().date_naive());
        Self {
            id: Uuid::new_v4(),
            employee_id,
            certificate_number: certificate_number.into(),
            issuer: issuer.into(),
            issue_date,
            expiry_date,
            certificate_type: certificate_type.into(),
            required_hours_for_renewal,
            hours_insufficient_flag: false,
            status,
            created_at: Utc::now(),
        }
    }

    pub fn calculate_status(expiry_date: &NaiveDate, today: &NaiveDate) -> CertificateStatus {
        let days_until_expiry = (*expiry_date - *today).num_days();
        if days_until_expiry < 0 {
            CertificateStatus::Expired
        } else if days_until_expiry <= 90 {
            CertificateStatus::ExpiringSoon
        } else {
            CertificateStatus::Valid
        }
    }

    pub fn update_status(&mut self, today: &NaiveDate) {
        if self.status != CertificateStatus::Revoked {
            self.status = Self::calculate_status(&self.expiry_date, today);
        }
    }

    pub fn renew(
        &mut self,
        new_expiry_date: NaiveDate,
        total_completed_hours: u32,
    ) -> CertificateRenewalResult {
        if self.status == CertificateStatus::Revoked {
            return CertificateRenewalResult::CannotRenewRevoked;
        }
        
        let hours_insufficient = total_completed_hours < self.required_hours_for_renewal;
        self.issue_date = self.expiry_date;
        self.expiry_date = new_expiry_date;
        self.hours_insufficient_flag = hours_insufficient;
        self.status = CertificateStatus::Valid;
        
        if hours_insufficient {
            CertificateRenewalResult::RenewedWithInsufficientHours {
                actual_hours: total_completed_hours,
                required_hours: self.required_hours_for_renewal,
            }
        } else {
            CertificateRenewalResult::RenewedSuccessfully
        }
    }

    pub fn revoke(&mut self) {
        self.status = CertificateStatus::Revoked;
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum CertificateRenewalResult {
    RenewedSuccessfully,
    RenewedWithInsufficientHours {
        actual_hours: u32,
        required_hours: u32,
    },
    CannotRenewRevoked,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrainingRecord {
    pub id: Uuid,
    pub name: String,
    pub date: NaiveDate,
    pub hours: u32,
    pub attendee_ids: Vec<Uuid>,
    pub created_at: DateTime<Utc>,
}

impl TrainingRecord {
    pub fn new(
        name: impl Into<String>,
        date: NaiveDate,
        hours: u32,
        attendee_ids: Vec<Uuid>,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            name: name.into(),
            date,
            hours,
            attendee_ids,
            created_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DepartmentCertificationRate {
    pub department_id: Uuid,
    pub department_name: String,
    pub total_employees: usize,
    pub certified_employees: usize,
    pub certification_rate: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CertificateTypeExpiryRate {
    pub certificate_type: String,
    pub total_certificates: usize,
    pub expired_count: usize,
    pub expiry_rate: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExpiringCertificatesAlert {
    pub employee_id: Uuid,
    pub employee_name: String,
    pub certificates: Vec<Certificate>,
}

#[derive(Debug, Clone, thiserror::Error)]
pub enum CertError {
    #[error("部门不存在")]
    DepartmentNotFound,
    #[error("员工不存在")]
    EmployeeNotFound,
    #[error("证书不存在")]
    CertificateNotFound,
    #[error("培训记录不存在")]
    TrainingRecordNotFound,
    #[error("部门名称已存在")]
    DepartmentNameExists,
    #[error("证书编号已存在")]
    CertificateNumberExists,
}
