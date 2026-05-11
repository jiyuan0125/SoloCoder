use crate::models::{Alert, BrokenChainEvent, Shipment, TemperatureReading};
use crate::types::VehicleId;
use chrono::{DateTime, Utc};
use std::collections::HashMap;
use std::sync::Mutex;
use uuid::Uuid;

pub trait Storage: Send + Sync {
    fn save_shipment(&self, shipment: Shipment);
    fn get_shipment(&self, id: &Uuid) -> Option<Shipment>;
    fn list_shipments(&self) -> Vec<Shipment>;

    fn save_temperature_reading(&self, reading: TemperatureReading);
    fn get_readings_by_shipment(&self, shipment_id: &Uuid) -> Vec<TemperatureReading>;

    fn save_alert(&self, alert: Alert);
    fn get_alerts_by_shipment(&self, shipment_id: &Uuid) -> Vec<Alert>;
    fn list_alerts(&self) -> Vec<Alert>;

    fn save_broken_chain_event(&self, event: BrokenChainEvent);
    fn get_broken_chain_events_by_vehicle(&self, vehicle_id: &VehicleId) -> Vec<BrokenChainEvent>;

    fn save_first_out_of_range_time(&self, shipment_id: &Uuid, time: DateTime<Utc>);
    fn get_first_out_of_range_time(&self, shipment_id: &Uuid) -> Option<DateTime<Utc>>;
    fn clear_first_out_of_range_time(&self, shipment_id: &Uuid);
}

pub struct InMemoryStorage {
    shipments: Mutex<HashMap<Uuid, Shipment>>,
    readings: Mutex<HashMap<Uuid, Vec<TemperatureReading>>>,
    alerts: Mutex<Vec<Alert>>,
    broken_chain_events: Mutex<HashMap<VehicleId, Vec<BrokenChainEvent>>>,
    first_out_of_range_times: Mutex<HashMap<Uuid, DateTime<Utc>>>,
}

impl InMemoryStorage {
    pub fn new() -> Self {
        Self {
            shipments: Mutex::new(HashMap::new()),
            readings: Mutex::new(HashMap::new()),
            alerts: Mutex::new(Vec::new()),
            broken_chain_events: Mutex::new(HashMap::new()),
            first_out_of_range_times: Mutex::new(HashMap::new()),
        }
    }
}

impl Default for InMemoryStorage {
    fn default() -> Self {
        Self::new()
    }
}

impl Storage for InMemoryStorage {
    fn save_shipment(&self, shipment: Shipment) {
        let mut shipments = self.shipments.lock().unwrap();
        shipments.insert(shipment.id, shipment);
    }

    fn get_shipment(&self, id: &Uuid) -> Option<Shipment> {
        let shipments = self.shipments.lock().unwrap();
        shipments.get(id).cloned()
    }

    fn list_shipments(&self) -> Vec<Shipment> {
        let shipments = self.shipments.lock().unwrap();
        shipments.values().cloned().collect()
    }

    fn save_temperature_reading(&self, reading: TemperatureReading) {
        let mut readings = self.readings.lock().unwrap();
        readings
            .entry(reading.shipment_id)
            .or_default()
            .push(reading);
    }

    fn get_readings_by_shipment(&self, shipment_id: &Uuid) -> Vec<TemperatureReading> {
        let readings = self.readings.lock().unwrap();
        readings.get(shipment_id).cloned().unwrap_or_default()
    }

    fn save_alert(&self, alert: Alert) {
        let mut alerts = self.alerts.lock().unwrap();
        alerts.push(alert);
    }

    fn get_alerts_by_shipment(&self, shipment_id: &Uuid) -> Vec<Alert> {
        let alerts = self.alerts.lock().unwrap();
        alerts
            .iter()
            .filter(|a| a.shipment_id == *shipment_id)
            .cloned()
            .collect()
    }

    fn list_alerts(&self) -> Vec<Alert> {
        let alerts = self.alerts.lock().unwrap();
        alerts.clone()
    }

    fn save_broken_chain_event(&self, event: BrokenChainEvent) {
        let mut events = self.broken_chain_events.lock().unwrap();
        events
            .entry(event.vehicle_id.clone())
            .or_default()
            .push(event);
    }

    fn get_broken_chain_events_by_vehicle(&self, vehicle_id: &VehicleId) -> Vec<BrokenChainEvent> {
        let events = self.broken_chain_events.lock().unwrap();
        events.get(vehicle_id).cloned().unwrap_or_default()
    }

    fn save_first_out_of_range_time(&self, shipment_id: &Uuid, time: DateTime<Utc>) {
        let mut times = self.first_out_of_range_times.lock().unwrap();
        times.entry(*shipment_id).or_insert(time);
    }

    fn get_first_out_of_range_time(&self, shipment_id: &Uuid) -> Option<DateTime<Utc>> {
        let times = self.first_out_of_range_times.lock().unwrap();
        times.get(shipment_id).copied()
    }

    fn clear_first_out_of_range_time(&self, shipment_id: &Uuid) {
        let mut times = self.first_out_of_range_times.lock().unwrap();
        times.remove(shipment_id);
    }
}
