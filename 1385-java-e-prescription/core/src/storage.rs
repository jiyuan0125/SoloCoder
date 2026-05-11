use std::collections::HashMap;
use std::sync::{Arc, RwLock};

use crate::models::{Doctor, DrugInteraction, Medicine, Patient, Prescription};
use uuid::Uuid;

#[derive(Clone, Default)]
pub struct InMemoryStorage {
    inner: Arc<RwLock<StorageInner>>,
}

#[derive(Default)]
struct StorageInner {
    prescriptions: HashMap<Uuid, Prescription>,
    patients: HashMap<Uuid, Patient>,
    medicines: HashMap<Uuid, Medicine>,
    doctors: HashMap<Uuid, Doctor>,
    drug_interactions: Vec<DrugInteraction>,
}

impl InMemoryStorage {
    pub fn new() -> Self {
        let storage = Self::default();
        storage.seed_data();
        storage
    }

    fn seed_data(&self) {
        let mut inner = self.inner.write().unwrap();
        
        let doctor1 = Doctor {
            id: Uuid::new_v4(),
            name: "张医生".to_string(),
        };
        let doctor2 = Doctor {
            id: Uuid::new_v4(),
            name: "李医生".to_string(),
        };

        let patient1 = Patient {
            id: Uuid::new_v4(),
            name: "王患者".to_string(),
            allergies: vec!["青霉素".to_string()],
        };
        let patient2 = Patient {
            id: Uuid::new_v4(),
            name: "赵患者".to_string(),
            allergies: vec![],
        };

        let medicine1 = Medicine {
            id: Uuid::new_v4(),
            name: "阿莫西林".to_string(),
            stock: 100,
        };
        let medicine2 = Medicine {
            id: Uuid::new_v4(),
            name: "布洛芬".to_string(),
            stock: 200,
        };
        let medicine3 = Medicine {
            id: Uuid::new_v4(),
            name: "奥美拉唑".to_string(),
            stock: 150,
        };
        let medicine4 = Medicine {
            id: Uuid::new_v4(),
            name: "头孢氨苄".to_string(),
            stock: 50,
        };
        let medicine5 = Medicine {
            id: Uuid::new_v4(),
            name: "青霉素".to_string(),
            stock: 80,
        };

        let interactions = vec![
            DrugInteraction {
                drug1: "阿莫西林".to_string(),
                drug2: "奥美拉唑".to_string(),
                severity: crate::models::InteractionSeverity::Severe,
                description: "阿莫西林与奥美拉唑同时使用会降低阿莫西林的吸收效果".to_string(),
            },
            DrugInteraction {
                drug1: "布洛芬".to_string(),
                drug2: "头孢氨苄".to_string(),
                severity: crate::models::InteractionSeverity::Warning,
                description: "布洛芬可能增强头孢氨苄的肾毒性".to_string(),
            },
        ];

        inner.doctors.insert(doctor1.id, doctor1);
        inner.doctors.insert(doctor2.id, doctor2);
        inner.patients.insert(patient1.id, patient1);
        inner.patients.insert(patient2.id, patient2);
        inner.medicines.insert(medicine1.id, medicine1);
        inner.medicines.insert(medicine2.id, medicine2);
        inner.medicines.insert(medicine3.id, medicine3);
        inner.medicines.insert(medicine4.id, medicine4);
        inner.medicines.insert(medicine5.id, medicine5);
        inner.drug_interactions = interactions;
    }

    pub fn add_prescription(&self, prescription: Prescription) {
        let mut inner = self.inner.write().unwrap();
        inner.prescriptions.insert(prescription.id, prescription);
    }

    pub fn get_prescription(&self, id: Uuid) -> Option<Prescription> {
        let inner = self.inner.read().unwrap();
        inner.prescriptions.get(&id).cloned()
    }

    pub fn get_all_prescriptions(&self) -> Vec<Prescription> {
        let inner = self.inner.read().unwrap();
        inner.prescriptions.values().cloned().collect()
    }

    pub fn update_prescription(&self, prescription: Prescription) {
        let mut inner = self.inner.write().unwrap();
        inner.prescriptions.insert(prescription.id, prescription);
    }

    pub fn get_patient(&self, id: Uuid) -> Option<Patient> {
        let inner = self.inner.read().unwrap();
        inner.patients.get(&id).cloned()
    }

    pub fn get_all_patients(&self) -> Vec<Patient> {
        let inner = self.inner.read().unwrap();
        inner.patients.values().cloned().collect()
    }

    pub fn get_medicine_by_name(&self, name: &str) -> Option<Medicine> {
        let inner = self.inner.read().unwrap();
        inner
            .medicines
            .values()
            .find(|m| m.name == name)
            .cloned()
    }

    pub fn get_all_medicines(&self) -> Vec<Medicine> {
        let inner = self.inner.read().unwrap();
        inner.medicines.values().cloned().collect()
    }

    pub fn update_medicine_stock(&self, medicine_id: Uuid, new_stock: u32) {
        let mut inner = self.inner.write().unwrap();
        if let Some(medicine) = inner.medicines.get_mut(&medicine_id) {
            medicine.stock = new_stock;
        }
    }

    pub fn get_all_doctors(&self) -> Vec<Doctor> {
        let inner = self.inner.read().unwrap();
        inner.doctors.values().cloned().collect()
    }

    pub fn get_doctor(&self, id: Uuid) -> Option<Doctor> {
        let inner = self.inner.read().unwrap();
        inner.doctors.get(&id).cloned()
    }

    pub fn get_drug_interactions(&self) -> Vec<DrugInteraction> {
        let inner = self.inner.read().unwrap();
        inner.drug_interactions.clone()
    }

    pub fn get_patient_prescriptions(&self, patient_id: Uuid) -> Vec<Prescription> {
        let inner = self.inner.read().unwrap();
        inner
            .prescriptions
            .values()
            .filter(|p| p.patient_id == patient_id)
            .cloned()
            .collect()
    }
}
