use crate::models::*;
use std::collections::HashMap;
use thiserror::Error;
use uuid::Uuid;

#[derive(Debug, Error)]
pub enum SettlementError {
    #[error("Hospital not found: {0}")]
    HospitalNotFound(Uuid),
    #[error("Medical item not found: {0}")]
    MedicalItemNotFound(String),
    #[error("Invalid hospital level configuration")]
    InvalidHospitalLevel,
    #[error("Invalid visit type configuration")]
    InvalidVisitType,
}

fn compute_self_pay_first(total_cost: u64, category: &ItemCategory) -> (u64, u64) {
    match category {
        ItemCategory::ClassA => (0, total_cost),
        ItemCategory::ClassB { self_pay_ratio_percent } => {
            let self_pay = (total_cost as u128 * *self_pay_ratio_percent as u128 / 100) as u64;
            (self_pay, total_cost - self_pay)
        }
        ItemCategory::ClassC => (total_cost, 0),
    }
}

fn get_effective_visit_type(visit_type: &VisitType) -> VisitType {
    match visit_type {
        VisitType::SpecialDiseaseOutpatient => VisitType::Inpatient,
        other => *other,
    }
}

fn get_reimbursement_ratio(
    config: &PolicyConfig,
    hospital_level: HospitalLevel,
    visit_type: VisitType,
    insured_status: InsuredStatus,
) -> Result<u32, SettlementError> {
    let level_map = config
        .reimbursement_ratios
        .get(&hospital_level)
        .ok_or(SettlementError::InvalidHospitalLevel)?;

    let effective_visit = get_effective_visit_type(&visit_type);
    let base_ratio = *level_map
        .get(&effective_visit)
        .ok_or(SettlementError::InvalidVisitType)?;

    let final_ratio = match insured_status {
        InsuredStatus::Retired => {
            let r = base_ratio + config.retired_bonus_percent;
            if r > 100 { 100 } else { r }
        }
        InsuredStatus::Working => base_ratio,
    };

    Ok(final_ratio)
}

pub struct SettlementContext<'a> {
    pub hospitals: &'a HashMap<Uuid, Hospital>,
    pub medical_items: &'a HashMap<String, MedicalItem>,
    pub policy: &'a PolicyConfig,
}

pub fn process_settlement(
    ctx: &SettlementContext,
    request: &SettlementRequest,
    patient_state: &mut PatientYearlyState,
) -> Result<SettlementResult, SettlementError> {
    let hospital = ctx
        .hospitals
        .get(&request.hospital_id)
        .ok_or(SettlementError::HospitalNotFound(request.hospital_id))?;

    let ratio = get_reimbursement_ratio(
        ctx.policy,
        hospital.level,
        request.visit_type,
        request.insured_status,
    )?;

    let effective_visit = get_effective_visit_type(&request.visit_type);
    let is_inpatient_path = matches!(effective_visit, VisitType::Inpatient);

    let (deductible, cap, state_reimbursed, state_deductible_met) = if is_inpatient_path {
        (
            ctx.policy.inpatient_deductible_cents,
            ctx.policy.inpatient_cap_cents,
            &mut patient_state.inpatient_reimbursed_cents,
            &mut patient_state.inpatient_deductible_met_cents,
        )
    } else {
        (
            ctx.policy.outpatient_deductible_cents,
            ctx.policy.outpatient_cap_cents,
            &mut patient_state.outpatient_reimbursed_cents,
            &mut patient_state.outpatient_deductible_met_cents,
        )
    };

    let remaining_cap = if cap > *state_reimbursed {
        cap - *state_reimbursed
    } else {
        0
    };

    let mut remaining_deductible = if deductible > *state_deductible_met {
        deductible - *state_deductible_met
    } else {
        0
    };

    let mut details = Vec::new();
    let mut total_cost = 0u64;
    let mut total_self_pay_first = 0u64;
    let mut total_reimbursable = 0u64;

    for visit_item in &request.items {
        let item = ctx
            .medical_items
            .get(&visit_item.item_id)
            .ok_or(SettlementError::MedicalItemNotFound(
                visit_item.item_id.clone(),
            ))?;

        let item_total = item.price_cents * visit_item.quantity as u64;
        let (self_pay_first, reimbursable) = compute_self_pay_first(item_total, &item.category);

        total_cost += item_total;
        total_self_pay_first += self_pay_first;
        total_reimbursable += reimbursable;

        details.push(ItemSettlementDetail {
            item_id: item.id.clone(),
            name: item.name.clone(),
            category: item.category,
            unit_price_cents: item.price_cents,
            quantity: visit_item.quantity,
            total_cost_cents: item_total,
            self_pay_first_cents: self_pay_first,
            reimbursable_cents: reimbursable,
            final_reimbursed_cents: 0,
            final_self_pay_cents: item_total,
        });
    }

    let mut pool_for_reimbursement = total_reimbursable;

    if remaining_deductible > 0 && pool_for_reimbursement > 0 {
        if pool_for_reimbursement <= remaining_deductible {
            let used = pool_for_reimbursement;
            *state_deductible_met += used;
            pool_for_reimbursement = 0;
        } else {
            let used = remaining_deductible;
            *state_deductible_met += used;
            pool_for_reimbursement -= used;
        }
    }

    let mut total_reimbursed = 0u64;
    if pool_for_reimbursement > 0 && remaining_cap > 0 {
        let tentative = (pool_for_reimbursement as u128 * ratio as u128 / 100) as u64;
        let actual = if tentative <= remaining_cap {
            tentative
        } else {
            remaining_cap
        };
        total_reimbursed = actual;
        *state_reimbursed += actual;
    }

    let total_final_self_pay = total_cost - total_reimbursed;

    let mut remaining_to_allocate = total_reimbursed;
    let mut remaining_reimbursable_portion = pool_for_reimbursement;

    for detail in details.iter_mut() {
        if detail.reimbursable_cents == 0 || remaining_reimbursable_portion == 0 {
            continue;
        }

        let used_in_deductible = if remaining_deductible > 0 {
            if detail.reimbursable_cents <= remaining_deductible {
                let u = detail.reimbursable_cents;
                remaining_deductible -= u;
                u
            } else {
                let u = remaining_deductible;
                remaining_deductible = 0;
                u
            }
        } else {
            0
        };

        let eligible_for_reimbursement = detail.reimbursable_cents - used_in_deductible;
        if eligible_for_reimbursement == 0 {
            continue;
        }

        if remaining_reimbursable_portion == 0 {
            continue;
        }

        let share_for_item = (eligible_for_reimbursement as u128
            * remaining_to_allocate as u128
            / remaining_reimbursable_portion as u128) as u64;

        detail.final_reimbursed_cents = share_for_item;
        detail.final_self_pay_cents = detail.total_cost_cents - share_for_item;

        remaining_to_allocate -= share_for_item;
        remaining_reimbursable_portion -= eligible_for_reimbursement;
    }

    Ok(SettlementResult {
        request_id: Uuid::new_v4(),
        patient_id: request.patient_id.clone(),
        hospital_id: request.hospital_id,
        visit_type: request.visit_type,
        policy_year: request.policy_year,
        total_cost_cents: total_cost,
        total_self_pay_first_cents: total_self_pay_first,
        total_reimbursable_cents: total_reimbursable,
        total_reimbursed_cents: total_reimbursed,
        total_final_self_pay_cents: total_final_self_pay,
        details,
        year_to_date_outpatient_reimbursed_cents: patient_state.outpatient_reimbursed_cents,
        year_to_date_inpatient_reimbursed_cents: patient_state.inpatient_reimbursed_cents,
    })
}
