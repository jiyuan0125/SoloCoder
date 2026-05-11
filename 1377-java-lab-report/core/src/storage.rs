use std::sync::Arc;
use std::collections::HashMap;

use chrono::NaiveDate;
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::models::{Report, ReportInput};
use crate::reference::ReferenceRegistry;
use crate::report::{generate_report, report_matches_query};

#[derive(Debug, Clone)]
pub struct ReportStorage {
    reports: Arc<RwLock<HashMap<Uuid, Report>>>,
    registry: Arc<ReferenceRegistry>,
}

impl ReportStorage {
    pub fn new() -> Self {
        Self {
            reports: Arc::new(RwLock::new(HashMap::new())),
            registry: Arc::new(ReferenceRegistry::new()),
        }
    }
    
    pub async fn create_report(&self, input: ReportInput) -> Report {
        let report = generate_report(input, &self.registry);
        let mut reports = self.reports.write().await;
        reports.insert(report.id, report.clone());
        report
    }
    
    pub async fn create_reports_batch(&self, inputs: Vec<ReportInput>) -> Vec<Report> {
        let mut reports_map = self.reports.write().await;
        let mut created = Vec::with_capacity(inputs.len());
        
        for input in inputs {
            let report = generate_report(input, &self.registry);
            reports_map.insert(report.id, report.clone());
            created.push(report);
        }
        
        created
    }
    
    pub async fn get_report(&self, id: &Uuid) -> Option<Report> {
        let reports = self.reports.read().await;
        reports.get(id).cloned()
    }
    
    pub async fn get_all_reports(&self) -> Vec<Report> {
        let reports = self.reports.read().await;
        reports.values().cloned().collect()
    }
    
    pub async fn search_reports(
        &self,
        patient_name: Option<&str>,
        start_date: Option<NaiveDate>,
        end_date: Option<NaiveDate>,
    ) -> Vec<Report> {
        let reports = self.reports.read().await;
        reports
            .values()
            .filter(|report| report_matches_query(report, patient_name, start_date, end_date))
            .cloned()
            .collect()
    }
    
    pub async fn update_doctor_info(
        &self,
        id: &Uuid,
        doctor_notes: Option<String>,
        diagnosis_suggestion: Option<String>,
    ) -> Option<Report> {
        let mut reports = self.reports.write().await;
        if let Some(report) = reports.get_mut(id) {
            if let Some(notes) = doctor_notes {
                report.doctor_notes = Some(notes);
            }
            if let Some(suggestion) = diagnosis_suggestion {
                report.diagnosis_suggestion = Some(suggestion);
            }
            return Some(report.clone());
        }
        None
    }
    
    pub async fn delete_report(&self, id: &Uuid) -> bool {
        let mut reports = self.reports.write().await;
        reports.remove(id).is_some()
    }
}

impl Default for ReportStorage {
    fn default() -> Self {
        Self::new()
    }
}
