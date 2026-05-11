use chrono::{NaiveDate, Utc};
use uuid::Uuid;

use crate::models::{Report, ReportInput, TestInput, TestItem, Patient};
use crate::reference::{ReferenceRegistry, evaluate_test};

pub fn generate_report(input: ReportInput, registry: &ReferenceRegistry) -> Report {
    let patient = Patient::new(
        input.patient_name,
        input.gender,
        input.age,
    );
    
    let report_date = input.report_date.unwrap_or_else(|| Utc::now().date_naive());
    
    let items: Vec<TestItem> = input
        .test_results
        .iter()
        .map(|test_input| {
            evaluate_test_item(
                test_input,
                patient.gender,
                patient.age,
                registry,
            )
        })
        .collect();
    
    let conclusion = generate_conclusion(&items);
    
    Report {
        id: Uuid::new_v4(),
        patient,
        report_date,
        items,
        conclusion,
        doctor_notes: input.doctor_notes,
        diagnosis_suggestion: input.diagnosis_suggestion,
    }
}

fn evaluate_test_item(
    input: &TestInput,
    gender: crate::models::Gender,
    age: u32,
    registry: &ReferenceRegistry,
) -> TestItem {
    evaluate_test(
        &input.name,
        &input.value,
        input.unit.as_deref(),
        gender,
        age,
        registry,
    )
}

fn generate_conclusion(items: &[TestItem]) -> String {
    let abnormalities: Vec<&TestItem> = items
        .iter()
        .filter(|item| item.abnormality.is_some())
        .collect();
    
    if abnormalities.is_empty() {
        return "所有检验项目结果正常".to_string();
    }
    
    let parts: Vec<String> = abnormalities
        .iter()
        .map(|item| {
            let ab = item.abnormality.as_ref().unwrap();
            let direction = match ab {
                crate::models::Abnormality::Low => "偏低",
                crate::models::Abnormality::High => "偏高",
            };
            format!("{}({})", item.name, direction)
        })
        .collect();
    
    format!("异常项目：{}", parts.join("、"))
}

pub fn report_matches_query(
    report: &Report,
    patient_name: Option<&str>,
    start_date: Option<NaiveDate>,
    end_date: Option<NaiveDate>,
) -> bool {
    if let Some(name) = patient_name {
        if !report.patient.name.contains(name) {
            return false;
        }
    }
    
    if let Some(start) = start_date {
        if report.report_date < start {
            return false;
        }
    }
    
    if let Some(end) = end_date {
        if report.report_date > end {
            return false;
        }
    }
    
    true
}
