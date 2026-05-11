use chrono::NaiveDate;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum Gender {
    Male,
    Female,
}

impl Gender {
    pub fn as_str(&self) -> &'static str {
        match self {
            Gender::Male => "男",
            Gender::Female => "女",
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum TestResultValue {
    Quantitative(f64),
    Qualitative(String),
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub enum Abnormality {
    Low,
    High,
}

impl Abnormality {
    pub fn as_str(&self) -> &'static str {
        match self {
            Abnormality::Low => "↓",
            Abnormality::High => "↑",
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct TestItem {
    pub name: String,
    pub value: TestResultValue,
    pub unit: Option<String>,
    pub reference_text: Option<String>,
    pub abnormality: Option<Abnormality>,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Patient {
    pub id: Uuid,
    pub name: String,
    pub gender: Gender,
    pub age: u32,
}

impl Patient {
    pub fn new(name: String, gender: Gender, age: u32) -> Self {
        Self {
            id: Uuid::new_v4(),
            name,
            gender,
            age,
        }
    }
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct Report {
    pub id: Uuid,
    pub patient: Patient,
    pub report_date: NaiveDate,
    pub items: Vec<TestItem>,
    pub conclusion: String,
    pub doctor_notes: Option<String>,
    pub diagnosis_suggestion: Option<String>,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct ReportInput {
    pub patient_name: String,
    pub gender: Gender,
    pub age: u32,
    pub report_date: Option<NaiveDate>,
    pub test_results: Vec<TestInput>,
    pub doctor_notes: Option<String>,
    pub diagnosis_suggestion: Option<String>,
}

#[derive(Debug, Clone, PartialEq, Serialize, Deserialize)]
pub struct TestInput {
    pub name: String,
    pub value: TestResultValue,
    pub unit: Option<String>,
}
