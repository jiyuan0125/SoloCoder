use chrono::{Duration, Utc};
use thiserror::Error;
use uuid::Uuid;

use crate::models::{
    CheckWarning, CreatePrescriptionRequest, DispensePrescriptionRequest, DoctorMonthlyStats,
    InteractionSeverity, MedicineRecord, MedicineUsageStats, Prescription, PrescriptionCheckResult,
    PrescriptionStatus, ReviewPrescriptionRequest,
};
use crate::storage::InMemoryStorage;

#[derive(Debug, Error)]
pub enum PrescriptionError {
    #[error("处方未找到")]
    NotFound,
    #[error("处方状态不允许此操作")]
    InvalidState,
    #[error("药品库存不足")]
    InsufficientStock,
    #[error("药品不存在")]
    MedicineNotFound,
    #[error("处方包含严重警告，无法审核通过")]
    HasSevereWarnings,
}

pub struct PrescriptionService {
    storage: InMemoryStorage,
}

impl PrescriptionService {
    pub fn new(storage: InMemoryStorage) -> Self {
        Self { storage }
    }

    pub fn create_prescription(
        &self,
        request: CreatePrescriptionRequest,
    ) -> Result<(Prescription, PrescriptionCheckResult), PrescriptionError> {
        for med in &request.medicines {
            let medicine = self
                .storage
                .get_medicine_by_name(&med.name)
                .ok_or(PrescriptionError::MedicineNotFound)?;
            if medicine.stock == 0 {
                return Err(PrescriptionError::InsufficientStock);
            }
        }

        let medicines: Vec<MedicineRecord> = request
            .medicines
            .into_iter()
            .map(|m| MedicineRecord::new(m.name, m.dosage, m.frequency, m.days, m.route))
            .collect();

        let prescription = Prescription::new(
            request.doctor_id,
            request.doctor_name,
            request.patient_id,
            request.patient_name,
            medicines,
        );

        let check_result = self.check_prescription(&prescription);

        self.storage.add_prescription(prescription.clone());

        Ok((prescription, check_result))
    }

    pub fn check_prescription(&self, prescription: &Prescription) -> PrescriptionCheckResult {
        let mut warnings = Vec::new();

        self.check_allergies(prescription, &mut warnings);
        self.check_duplicate_medication(prescription, &mut warnings);
        self.check_drug_interactions(prescription, &mut warnings);

        let has_severe = warnings
            .iter()
            .any(|w| matches!(w, CheckWarning::Severe(_)));

        PrescriptionCheckResult {
            warnings,
            is_safe: !has_severe,
        }
    }

    fn check_allergies(&self, prescription: &Prescription, warnings: &mut Vec<CheckWarning>) {
        if let Some(patient) = self.storage.get_patient(prescription.patient_id) {
            for med in &prescription.medicines {
                if patient.allergies.contains(&med.name) {
                    warnings.push(CheckWarning::Severe(format!(
                        "患者对药品 '{}' 过敏！",
                        med.name
                    )));
                }
            }
        }
    }

    fn check_duplicate_medication(
        &self,
        prescription: &Prescription,
        warnings: &mut Vec<CheckWarning>,
    ) {
        let now = Utc::now();
        let seven_days_ago = now - Duration::days(7);

        let patient_prescriptions = self.storage.get_patient_prescriptions(prescription.patient_id);

        for med in &prescription.medicines {
            for existing in &patient_prescriptions {
                if existing.id == prescription.id {
                    continue;
                }

                if existing.created_at < seven_days_ago {
                    continue;
                }

                for existing_med in &existing.medicines {
                    if existing_med.name == med.name {
                        let existing_end =
                            existing.created_at + Duration::days(existing_med.days as i64);
                        if existing_end > now {
                            warnings.push(CheckWarning::Warning(format!(
                                "药品 '{}' 在过去7天内已开过，且上次疗程尚未结束（预计结束于 {}）",
                                med.name,
                                existing_end.format("%Y-%m-%d")
                            )));
                        }
                    }
                }
            }
        }
    }

    fn check_drug_interactions(
        &self,
        prescription: &Prescription,
        warnings: &mut Vec<CheckWarning>,
    ) {
        let interactions = self.storage.get_drug_interactions();
        let medicine_names: Vec<&str> = prescription
            .medicines
            .iter()
            .map(|m| m.name.as_str())
            .collect();

        for interaction in &interactions {
            let has_drug1 = medicine_names.contains(&interaction.drug1.as_str());
            let has_drug2 = medicine_names.contains(&interaction.drug2.as_str());

            if has_drug1 && has_drug2 {
                match interaction.severity {
                    InteractionSeverity::Severe => {
                        warnings.push(CheckWarning::Severe(format!(
                            "药物相互作用（严重）: {} - {}",
                            interaction.drug1, interaction.description
                        )));
                    }
                    InteractionSeverity::Warning => {
                        warnings.push(CheckWarning::Warning(format!(
                            "药物相互作用（警告）: {} - {}",
                            interaction.drug1, interaction.description
                        )));
                    }
                }
            }
        }
    }

    pub fn review_prescription(
        &self,
        request: ReviewPrescriptionRequest,
    ) -> Result<Prescription, PrescriptionError> {
        let mut prescription = self
            .storage
            .get_prescription(request.prescription_id)
            .ok_or(PrescriptionError::NotFound)?;

        if prescription.status != PrescriptionStatus::PendingReview {
            return Err(PrescriptionError::InvalidState);
        }

        let check_result = self.check_prescription(&prescription);

        if request.approved {
            if !check_result.is_safe {
                return Err(PrescriptionError::HasSevereWarnings);
            }

            for med in &prescription.medicines {
                let medicine = self
                    .storage
                    .get_medicine_by_name(&med.name)
                    .ok_or(PrescriptionError::MedicineNotFound)?;
                if medicine.stock == 0 {
                    return Err(PrescriptionError::InsufficientStock);
                }
            }

            prescription.status = PrescriptionStatus::Approved;
        } else {
            prescription.status = PrescriptionStatus::Rejected;
        }

        prescription.reviewed_at = Some(Utc::now());
        prescription.review_notes = request.notes;

        self.storage.update_prescription(prescription.clone());

        Ok(prescription)
    }

    pub fn dispense_prescription(
        &self,
        request: DispensePrescriptionRequest,
    ) -> Result<Prescription, PrescriptionError> {
        let mut prescription = self
            .storage
            .get_prescription(request.prescription_id)
            .ok_or(PrescriptionError::NotFound)?;

        if prescription.status != PrescriptionStatus::Approved {
            return Err(PrescriptionError::InvalidState);
        }

        for med in &prescription.medicines {
            let medicine = self
                .storage
                .get_medicine_by_name(&med.name)
                .ok_or(PrescriptionError::MedicineNotFound)?;
            if medicine.stock < 1 {
                return Err(PrescriptionError::InsufficientStock);
            }
            self.storage
                .update_medicine_stock(medicine.id, medicine.stock - 1);
        }

        prescription.status = PrescriptionStatus::Dispensed;
        prescription.dispensed_at = Some(Utc::now());

        self.storage.update_prescription(prescription.clone());

        Ok(prescription)
    }

    pub fn get_prescription(&self, id: Uuid) -> Option<Prescription> {
        self.storage.get_prescription(id)
    }

    pub fn get_all_prescriptions(&self) -> Vec<Prescription> {
        self.storage.get_all_prescriptions()
    }

    pub fn get_doctor_monthly_stats(&self, doctor_id: Uuid) -> Option<DoctorMonthlyStats> {
        let doctor = self.storage.get_doctor(doctor_id)?;
        let now = Utc::now();
        let month = now.format("%Y-%m").to_string();

        let count = self
            .storage
            .get_all_prescriptions()
            .iter()
            .filter(|p| {
                p.doctor_id == doctor_id
                    && p.created_at.format("%Y-%m").to_string() == month
            })
            .count() as u32;

        Some(DoctorMonthlyStats {
            doctor_id,
            doctor_name: doctor.name,
            month,
            prescription_count: count,
        })
    }

    pub fn get_all_doctors_monthly_stats(&self) -> Vec<DoctorMonthlyStats> {
        self.storage
            .get_all_doctors()
            .into_iter()
            .filter_map(|d| self.get_doctor_monthly_stats(d.id))
            .collect()
    }

    pub fn get_medicine_usage_stats(&self) -> Vec<MedicineUsageStats> {
        let mut usage: std::collections::HashMap<String, u32> = std::collections::HashMap::new();

        for prescription in self.storage.get_all_prescriptions() {
            if prescription.status == PrescriptionStatus::Dispensed {
                for med in &prescription.medicines {
                    *usage.entry(med.name.clone()).or_insert(0) += 1;
                }
            }
        }

        let mut stats: Vec<MedicineUsageStats> = usage
            .into_iter()
            .map(|(name, count)| MedicineUsageStats {
                medicine_name: name,
                usage_count: count,
            })
            .collect();

        stats.sort_by(|a, b| b.usage_count.cmp(&a.usage_count));

        stats
    }

    pub fn get_all_medicines(&self) -> Vec<crate::models::Medicine> {
        self.storage.get_all_medicines()
    }

    pub fn get_all_patients(&self) -> Vec<crate::models::Patient> {
        self.storage.get_all_patients()
    }

    pub fn get_all_doctors(&self) -> Vec<crate::models::Doctor> {
        self.storage.get_all_doctors()
    }
}
