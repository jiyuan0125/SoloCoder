use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ItemCategory {
    ClassA,
    ClassB { self_pay_ratio_percent: u32 },
    ClassC,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum HospitalLevel {
    Community,
    Level2,
    Level3,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum VisitType {
    Outpatient,
    Inpatient,
    SpecialDiseaseOutpatient,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum InsuredStatus {
    Working,
    Retired,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct MedicalItem {
    pub id: String,
    pub name: String,
    pub category: ItemCategory,
    pub price_cents: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Hospital {
    pub id: Uuid,
    pub name: String,
    pub level: HospitalLevel,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PolicyConfig {
    pub outpatient_deductible_cents: u64,
    pub outpatient_cap_cents: u64,
    pub inpatient_deductible_cents: u64,
    pub inpatient_cap_cents: u64,
    pub reimbursement_ratios: HashMap<HospitalLevel, HashMap<VisitType, u32>>,
    pub retired_bonus_percent: u32,
}

impl Default for PolicyConfig {
    fn default() -> Self {
        let mut ratios = HashMap::new();
        let mut community = HashMap::new();
        community.insert(VisitType::Outpatient, 70);
        community.insert(VisitType::Inpatient, 90);
        community.insert(VisitType::SpecialDiseaseOutpatient, 90);
        ratios.insert(HospitalLevel::Community, community);

        let mut level2 = HashMap::new();
        level2.insert(VisitType::Outpatient, 60);
        level2.insert(VisitType::Inpatient, 85);
        level2.insert(VisitType::SpecialDiseaseOutpatient, 85);
        ratios.insert(HospitalLevel::Level2, level2);

        let mut level3 = HashMap::new();
        level3.insert(VisitType::Outpatient, 50);
        level3.insert(VisitType::Inpatient, 80);
        level3.insert(VisitType::SpecialDiseaseOutpatient, 80);
        ratios.insert(HospitalLevel::Level3, level3);

        PolicyConfig {
            outpatient_deductible_cents: 10000,
            outpatient_cap_cents: 2000000,
            inpatient_deductible_cents: 15000,
            inpatient_cap_cents: 5000000,
            reimbursement_ratios: ratios,
            retired_bonus_percent: 5,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VisitItem {
    pub item_id: String,
    pub quantity: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SettlementRequest {
    pub patient_id: String,
    pub hospital_id: Uuid,
    pub visit_type: VisitType,
    pub insured_status: InsuredStatus,
    pub items: Vec<VisitItem>,
    pub policy_year: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ItemSettlementDetail {
    pub item_id: String,
    pub name: String,
    pub category: ItemCategory,
    pub unit_price_cents: u64,
    pub quantity: u32,
    pub total_cost_cents: u64,
    pub self_pay_first_cents: u64,
    pub reimbursable_cents: u64,
    pub final_reimbursed_cents: u64,
    pub final_self_pay_cents: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SettlementResult {
    pub request_id: Uuid,
    pub patient_id: String,
    pub hospital_id: Uuid,
    pub visit_type: VisitType,
    pub policy_year: u32,
    pub total_cost_cents: u64,
    pub total_self_pay_first_cents: u64,
    pub total_reimbursable_cents: u64,
    pub total_reimbursed_cents: u64,
    pub total_final_self_pay_cents: u64,
    pub details: Vec<ItemSettlementDetail>,
    pub year_to_date_outpatient_reimbursed_cents: u64,
    pub year_to_date_inpatient_reimbursed_cents: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PatientYearlyState {
    pub patient_id: String,
    pub policy_year: u32,
    pub outpatient_reimbursed_cents: u64,
    pub inpatient_reimbursed_cents: u64,
    pub outpatient_deductible_met_cents: u64,
    pub inpatient_deductible_met_cents: u64,
}

impl Default for PatientYearlyState {
    fn default() -> Self {
        PatientYearlyState {
            patient_id: String::new(),
            policy_year: 0,
            outpatient_reimbursed_cents: 0,
            inpatient_reimbursed_cents: 0,
            outpatient_deductible_met_cents: 0,
            inpatient_deductible_met_cents: 0,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreatePatientState {
    pub patient_id: String,
    pub policy_year: u32,
}
