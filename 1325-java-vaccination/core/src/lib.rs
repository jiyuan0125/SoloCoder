use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use thiserror::Error;
use uuid::Uuid;

pub mod models {
    use super::*;

    #[derive(Debug, Clone, Serialize, Deserialize, PartialEq, Eq, Hash)]
    pub struct ContraindicationTag(pub String);

    #[derive(Debug, Clone, Serialize, Deserialize)]
    pub struct VaccineDoseSchedule {
        pub dose_number: u32,
        pub min_interval_days: u32,
    }

    #[derive(Debug, Clone, Serialize, Deserialize)]
    pub struct Vaccine {
        pub id: Uuid,
        pub name: String,
        pub total_doses: u32,
        pub dose_schedules: Vec<VaccineDoseSchedule>,
        pub contraindications: Vec<ContraindicationTag>,
    }

    #[derive(Debug, Clone, Serialize, Deserialize)]
    pub struct Person {
        pub id: Uuid,
        pub name: String,
        pub contraindications: Vec<ContraindicationTag>,
    }

    #[derive(Debug, Clone, Serialize, Deserialize)]
    pub struct VaccinationRecord {
        pub id: Uuid,
        pub person_id: Uuid,
        pub vaccine_id: Uuid,
        pub dose_number: u32,
        pub vaccination_date: DateTime<Utc>,
        pub voided: bool,
        pub created_at: DateTime<Utc>,
    }

    #[derive(Debug, Clone, Serialize, Deserialize)]
    pub struct VaccinationValidationResult {
        pub can_vaccinate: bool,
        pub interval_ok: bool,
        pub days_remaining: Option<i64>,
        pub contraindication_warnings: Vec<String>,
        pub message: String,
    }

    #[derive(Debug, Error)]
    pub enum VaccinationError {
        #[error("Vaccine not found")]
        VaccineNotFound,
        #[error("Person not found")]
        PersonNotFound,
        #[error("Invalid dose number: expected max {max}, got {got}")]
        InvalidDoseNumber { max: u32, got: u32 },
        #[error("Interval not met: need {required} days, only {elapsed} days elapsed")]
        IntervalNotMet { required: u32, elapsed: i64 },
        #[error("Record not found")]
        RecordNotFound,
        #[error("Record already voided")]
        RecordAlreadyVoided,
    }
}

pub mod service {
    use super::*;
    use models::*;

    #[derive(Debug, Default, Clone)]
    pub struct InMemoryStore {
        pub vaccines: HashMap<Uuid, Vaccine>,
        pub persons: HashMap<Uuid, Person>,
        pub records: HashMap<Uuid, VaccinationRecord>,
        pub records_by_person: HashMap<Uuid, Vec<Uuid>>,
        pub records_by_vaccine: HashMap<Uuid, Vec<Uuid>>,
    }

    impl InMemoryStore {
        pub fn new() -> Self {
            Self::default()
        }

        pub fn add_vaccine(&mut self, vaccine: Vaccine) {
            self.vaccines.insert(vaccine.id, vaccine);
        }

        pub fn get_vaccine(&self, id: &Uuid) -> Option<&Vaccine> {
            self.vaccines.get(id)
        }

        pub fn list_vaccines(&self) -> Vec<Vaccine> {
            self.vaccines.values().cloned().collect()
        }

        pub fn add_person(&mut self, person: Person) {
            let person_id = person.id;
            self.persons.insert(person_id, person);
            self.records_by_person.entry(person_id).or_default();
        }

        pub fn get_person(&self, id: &Uuid) -> Option<&Person> {
            self.persons.get(id)
        }

        pub fn list_persons(&self) -> Vec<Person> {
            self.persons.values().cloned().collect()
        }

        pub fn add_record(&mut self, record: VaccinationRecord) {
            let record_id = record.id;
            let person_id = record.person_id;
            let vaccine_id = record.vaccine_id;

            self.records.insert(record_id, record);
            self.records_by_person
                .entry(person_id)
                .or_default()
                .push(record_id);
            self.records_by_vaccine
                .entry(vaccine_id)
                .or_default()
                .push(record_id);
        }

        pub fn get_record(&self, id: &Uuid) -> Option<&VaccinationRecord> {
            self.records.get(id)
        }

        pub fn list_records(&self) -> Vec<VaccinationRecord> {
            self.records.values().cloned().collect()
        }

        pub fn list_records_for_person(&self, person_id: &Uuid) -> Vec<VaccinationRecord> {
            self.records_by_person
                .get(person_id)
                .map(|ids| ids.iter().filter_map(|id| self.records.get(id).cloned()).collect())
                .unwrap_or_default()
        }

        pub fn list_active_records_for_person_and_vaccine(
            &self,
            person_id: &Uuid,
            vaccine_id: &Uuid,
        ) -> Vec<VaccinationRecord> {
            self.records_by_person
                .get(person_id)
                .map(|ids| {
                    ids.iter()
                        .filter_map(|id| self.records.get(id))
                        .filter(|r| r.vaccine_id == *vaccine_id && !r.voided)
                        .cloned()
                        .collect()
                })
                .unwrap_or_default()
        }

        pub fn void_record(&mut self, id: &Uuid) -> Result<(), VaccinationError> {
            let record = self.records.get_mut(id).ok_or(VaccinationError::RecordNotFound)?;
            if record.voided {
                return Err(VaccinationError::RecordAlreadyVoided);
            }
            record.voided = true;
            Ok(())
        }
    }

    #[derive(Debug, Clone)]
    pub struct VaccinationService {
        store: InMemoryStore,
    }

    impl VaccinationService {
        pub fn new() -> Self {
            Self {
                store: InMemoryStore::new(),
            }
        }

        pub fn store(&self) -> &InMemoryStore {
            &self.store
        }

        pub fn store_mut(&mut self) -> &mut InMemoryStore {
            &mut self.store
        }

        pub fn validate_vaccination(
            &self,
            person_id: &Uuid,
            vaccine_id: &Uuid,
            vaccination_date: &DateTime<Utc>,
        ) -> Result<VaccinationValidationResult, VaccinationError> {
            let vaccine = self.store.get_vaccine(vaccine_id)
                .ok_or(VaccinationError::VaccineNotFound)?;
            let person = self.store.get_person(person_id)
                .ok_or(VaccinationError::PersonNotFound)?;

            let active_records = self.store.list_active_records_for_person_and_vaccine(
                person_id,
                vaccine_id,
            );

            let next_dose_number = active_records.len() as u32 + 1;
            if next_dose_number > vaccine.total_doses {
                return Err(VaccinationError::InvalidDoseNumber {
                    max: vaccine.total_doses,
                    got: next_dose_number,
                });
            }

            let mut interval_ok = true;
            let mut days_remaining: Option<i64> = None;
            let mut message = String::new();

            if let Some(last_record) = active_records.last() {
                let elapsed = (*vaccination_date - last_record.vaccination_date).num_days();
                let schedule = vaccine.dose_schedules.iter()
                    .find(|s| s.dose_number == next_dose_number);
                
                let min_interval = schedule.map(|s| s.min_interval_days).unwrap_or(0);
                
                if elapsed < min_interval as i64 {
                    interval_ok = false;
                    days_remaining = Some(min_interval as i64 - elapsed);
                    message = format!(
                        "间隔不足：需要{}天，已过{}天，还差{}天",
                        min_interval,
                        elapsed,
                        days_remaining.unwrap()
                    );
                }
            }

            let mut contraindication_warnings = Vec::new();
            for tag in &person.contraindications {
                if vaccine.contraindications.contains(tag) {
                    contraindication_warnings.push(format!(
                        "提醒：该人员有禁忌症标签 '{}'，不建议接种此疫苗",
                        tag.0
                    ));
                }
            }

            if interval_ok {
                if contraindication_warnings.is_empty() {
                    message = "可以接种".to_string();
                } else {
                    message = "可以接种，但有禁忌症提醒".to_string();
                }
            }

            Ok(VaccinationValidationResult {
                can_vaccinate: interval_ok,
                interval_ok,
                days_remaining,
                contraindication_warnings,
                message,
            })
        }

        pub fn record_vaccination(
            &mut self,
            person_id: Uuid,
            vaccine_id: Uuid,
            vaccination_date: DateTime<Utc>,
        ) -> Result<VaccinationRecord, VaccinationError> {
            let validation = self.validate_vaccination(&person_id, &vaccine_id, &vaccination_date)?;

            if !validation.interval_ok {
                return Err(VaccinationError::IntervalNotMet {
                    required: validation.days_remaining.map(|d| d as u32).unwrap_or(0),
                    elapsed: 0,
                });
            }

            let active_records = self.store.list_active_records_for_person_and_vaccine(
                &person_id,
                &vaccine_id,
            );
            let dose_number = active_records.len() as u32 + 1;

            let record = VaccinationRecord {
                id: Uuid::new_v4(),
                person_id,
                vaccine_id,
                dose_number,
                vaccination_date,
                voided: false,
                created_at: Utc::now(),
            };

            self.store.add_record(record.clone());
            Ok(record)
        }
    }
}

pub use models::*;
pub use service::*;
