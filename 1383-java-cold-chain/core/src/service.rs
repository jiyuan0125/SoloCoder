use crate::error::ColdChainError;
use crate::models::{
    Alert, BrokenChainEvent, CargoTypeStats, Shipment, TemperatureReading, VehicleMonthlyStats,
};
use crate::storage::Storage;
use crate::types::{ShipmentStatus, VehicleId, CargoType};
use chrono::{DateTime, Datelike, Utc};
use std::sync::Arc;
use uuid::Uuid;

pub struct ColdChainService {
    storage: Arc<dyn Storage>,
}

impl ColdChainService {
    pub fn new(storage: Arc<dyn Storage>) -> Self {
        Self { storage }
    }

    pub fn create_shipment(
        &self,
        cargo_type: CargoType,
        vehicle_id: VehicleId,
        temp_min: f64,
        temp_max: f64,
    ) -> Shipment {
        let shipment = Shipment::new(cargo_type, vehicle_id, temp_min, temp_max);
        self.storage.save_shipment(shipment.clone());
        shipment
    }

    pub fn get_shipment(&self, id: &Uuid) -> Result<Shipment, ColdChainError> {
        self.storage
            .get_shipment(id)
            .ok_or_else(|| ColdChainError::ShipmentNotFound(id.to_string()))
    }

    pub fn list_shipments(&self) -> Vec<Shipment> {
        self.storage.list_shipments()
    }

    pub fn start_shipment(&self, id: &Uuid) -> Result<Shipment, ColdChainError> {
        let mut shipment = self.get_shipment(id)?;
        
        if !shipment.status.can_transition_to(&ShipmentStatus::InTransit) {
            return Err(ColdChainError::InvalidStateTransition(format!(
                "Cannot transition from {} to InTransit",
                shipment.status
            )));
        }

        shipment.status = ShipmentStatus::InTransit;
        shipment.started_at = Some(Utc::now());
        self.storage.save_shipment(shipment.clone());
        
        Ok(shipment)
    }

    pub fn report_temperature(
        &self,
        shipment_id: &Uuid,
        temperature: f64,
    ) -> Result<TemperatureReading, ColdChainError> {
        let shipment = self.get_shipment(shipment_id)?;

        if shipment.status != ShipmentStatus::InTransit {
            return Err(ColdChainError::ShipmentNotInTransit);
        }

        let alert_level = shipment.check_temperature(temperature);
        let now = Utc::now();

        let mut reading = TemperatureReading::new(*shipment_id, temperature);
        reading.alert_level = alert_level;
        reading.recorded_at = now;
        self.storage.save_temperature_reading(reading.clone());

        if let Some(level) = alert_level {
            let alert = Alert::new(*shipment_id, level, temperature);
            self.storage.save_alert(alert);

            if let Some(first_out_of_range) = self.storage.get_first_out_of_range_time(shipment_id) {
                let duration = now.signed_duration_since(first_out_of_range);
                if duration.num_minutes() >= 30 && !shipment.is_broken_chain {
                    self.mark_broken_chain(&shipment, first_out_of_range);
                }
            } else {
                self.storage.save_first_out_of_range_time(shipment_id, now);
            }
        } else {
            self.storage.clear_first_out_of_range_time(shipment_id);
        }

        Ok(reading)
    }

    fn mark_broken_chain(&self, shipment: &Shipment, first_out_of_range: DateTime<Utc>) {
        let mut updated_shipment = shipment.clone();
        updated_shipment.is_broken_chain = true;
        updated_shipment.broken_chain_detected_at = Some(Utc::now());
        self.storage.save_shipment(updated_shipment);

        let event = BrokenChainEvent::new(
            shipment.id,
            shipment.vehicle_id.clone(),
            shipment.cargo_type.clone(),
            first_out_of_range,
        );
        self.storage.save_broken_chain_event(event);
    }

    pub fn complete_shipment(&self, id: &Uuid) -> Result<Shipment, ColdChainError> {
        let mut shipment = self.get_shipment(id)?;

        if !shipment.status.can_transition_to(&ShipmentStatus::PendingAcceptance) {
            return Err(ColdChainError::InvalidStateTransition(format!(
                "Cannot transition from {} to PendingAcceptance",
                shipment.status
            )));
        }

        shipment.status = ShipmentStatus::PendingAcceptance;
        shipment.completed_at = Some(Utc::now());
        self.storage.save_shipment(shipment.clone());

        Ok(shipment)
    }

    pub fn accept_shipment(&self, id: &Uuid) -> Result<Shipment, ColdChainError> {
        let mut shipment = self.get_shipment(id)?;

        if shipment.is_broken_chain {
            return Err(ColdChainError::BrokenChainCannotAccept);
        }

        if !shipment.status.can_transition_to(&ShipmentStatus::Completed) {
            return Err(ColdChainError::InvalidStateTransition(format!(
                "Cannot transition from {} to Completed",
                shipment.status
            )));
        }

        shipment.status = ShipmentStatus::Completed;
        self.storage.save_shipment(shipment.clone());

        Ok(shipment)
    }

    pub fn reject_shipment(&self, id: &Uuid) -> Result<Shipment, ColdChainError> {
        let mut shipment = self.get_shipment(id)?;

        if !shipment.status.can_transition_to(&ShipmentStatus::Rejected) {
            return Err(ColdChainError::InvalidStateTransition(format!(
                "Cannot transition from {} to Rejected",
                shipment.status
            )));
        }

        shipment.status = ShipmentStatus::Rejected;
        self.storage.save_shipment(shipment.clone());

        Ok(shipment)
    }

    pub fn list_alerts(&self, shipment_id: Option<&Uuid>) -> Vec<Alert> {
        match shipment_id {
            Some(id) => self.storage.get_alerts_by_shipment(id),
            None => self.storage.list_alerts(),
        }
    }

    pub fn get_vehicle_monthly_stats(
        &self,
        vehicle_id: &VehicleId,
        year: i32,
        month: u32,
    ) -> VehicleMonthlyStats {
        let events = self.storage.get_broken_chain_events_by_vehicle(vehicle_id);
        
        let broken_chain_count = events
            .iter()
            .filter(|e| {
                let detected = e.detected_at;
                detected.year() == year && detected.month() == month
            })
            .count() as u32;

        let shipments = self.storage.list_shipments();
        let total_shipments = shipments
            .iter()
            .filter(|s| {
                s.vehicle_id == *vehicle_id
                    && s.started_at.map(|t| t.year() == year && t.month() == month).unwrap_or(false)
            })
            .count() as u32;

        VehicleMonthlyStats {
            vehicle_id: vehicle_id.clone(),
            year,
            month,
            broken_chain_count,
            total_shipments,
        }
    }

    pub fn get_cargo_type_stats(&self, cargo_type: &CargoType) -> CargoTypeStats {
        let shipments = self.storage.list_shipments();
        
        let filtered: Vec<_> = shipments
            .iter()
            .filter(|s| s.cargo_type == *cargo_type)
            .collect();

        let total_shipments = filtered.len() as u32;
        let rejected_shipments = filtered
            .iter()
            .filter(|s| s.status == ShipmentStatus::Rejected)
            .count() as u32;

        let loss_rate = if total_shipments > 0 {
            rejected_shipments as f64 / total_shipments as f64
        } else {
            0.0
        };

        CargoTypeStats {
            cargo_type: cargo_type.clone(),
            total_shipments,
            rejected_shipments,
            loss_rate,
        }
    }
}
