use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum PrescriptionStatus {
    PendingReview,
    Approved,
    Rejected,
    Dispensed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MedicineRecord {
    pub id: Uuid,
    pub name: String,
    pub dosage: String,
    pub frequency: String,
    pub days: u32,
    pub route: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Prescription {
    pub id: Uuid,
    pub doctor_id: Uuid,
    pub doctor_name: String,
    pub patient_id: Uuid,
    pub patient_name: String,
    pub medicines: Vec<MedicineRecord>,
    pub status: PrescriptionStatus,
    pub created_at: DateTime<Utc>,
    pub reviewed_at: Option<DateTime<Utc>>,
    pub dispensed_at: Option<DateTime<Utc>>,
    pub review_notes: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Patient {
    pub id: Uuid,
    pub name: String,
    pub allergies: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Medicine {
    pub id: Uuid,
    pub name: String,
    pub stock: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Doctor {
    pub id: Uuid,
    pub name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DrugInteraction {
    pub drug1: String,
    pub drug2: String,
    pub severity: InteractionSeverity,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum InteractionSeverity {
    Warning,
    Severe,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum CheckWarning {
    Severe(String),
    Warning(String),
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PrescriptionCheckResult {
    pub warnings: Vec<CheckWarning>,
    pub is_safe: bool,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DoctorMonthlyStats {
    pub doctor_id: Uuid,
    pub doctor_name: String,
    pub month: String,
    pub prescription_count: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MedicineUsageStats {
    pub medicine_name: String,
    pub usage_count: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreatePrescriptionRequest {
    pub doctor_id: Uuid,
    pub doctor_name: String,
    pub patient_id: Uuid,
    pub patient_name: String,
    pub medicines: Vec<MedicineRecordRequest>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MedicineRecordRequest {
    pub name: String,
    pub dosage: String,
    pub frequency: String,
    pub days: u32,
    pub route: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ReviewPrescriptionRequest {
    pub prescription_id: Uuid,
    pub approved: bool,
    pub notes: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DispensePrescriptionRequest {
    pub prescription_id: Uuid,
}

impl Prescription {
    pub fn new(
        doctor_id: Uuid,
        doctor_name: String,
        patient_id: Uuid,
        patient_name: String,
        medicines: Vec<MedicineRecord>,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            doctor_id,
            doctor_name,
            patient_id,
            patient_name,
            medicines,
            status: PrescriptionStatus::PendingReview,
            created_at: Utc::now(),
            reviewed_at: None,
            dispensed_at: None,
            review_notes: None,
        }
    }
}

impl MedicineRecord {
    pub fn new(
        name: String,
        dosage: String,
        frequency: String,
        days: u32,
        route: String,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            dosage,
            frequency,
            days,
            route,
        }
    }
}
