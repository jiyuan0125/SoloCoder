use crate::{Claim, Vehicle};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

#[derive(Debug, Clone)]
pub struct InMemoryStore {
    vehicles: Arc<Mutex<HashMap<String, Vehicle>>>,
    claims: Arc<Mutex<HashMap<String, Claim>>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self {
            vehicles: Arc::new(Mutex::new(HashMap::new())),
            claims: Arc::new(Mutex::new(HashMap::new())),
        }
    }

    pub fn get_vehicle(&self, id: &str) -> Option<Vehicle> {
        let vehicles = self.vehicles.lock().unwrap();
        vehicles.get(id).cloned()
    }

    pub fn get_vehicle_by_plate(&self, plate_number: &str) -> Option<Vehicle> {
        let vehicles = self.vehicles.lock().unwrap();
        vehicles
            .values()
            .find(|v| v.plate_number == plate_number)
            .cloned()
    }

    pub fn list_vehicles(&self) -> Vec<Vehicle> {
        let vehicles = self.vehicles.lock().unwrap();
        vehicles.values().cloned().collect()
    }

    pub fn save_vehicle(&self, vehicle: Vehicle) {
        let mut vehicles = self.vehicles.lock().unwrap();
        vehicles.insert(vehicle.id.clone(), vehicle);
    }

    pub fn get_claim(&self, id: &str) -> Option<Claim> {
        let claims = self.claims.lock().unwrap();
        claims.get(id).cloned()
    }

    pub fn list_claims(&self) -> Vec<Claim> {
        let claims = self.claims.lock().unwrap();
        claims.values().cloned().collect()
    }

    pub fn list_claims_by_vehicle(&self, vehicle_id: &str) -> Vec<Claim> {
        let claims = self.claims.lock().unwrap();
        claims
            .values()
            .filter(|c| c.vehicle_id == vehicle_id)
            .cloned()
            .collect()
    }

    pub fn save_claim(&self, claim: Claim) {
        let mut claims = self.claims.lock().unwrap();
        claims.insert(claim.id.clone(), claim);
    }
}
