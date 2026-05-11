use crate::types::{AlertLevel, ShipmentStatus, CargoType, VehicleId};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TemperatureRange {
    pub min: f64,
    pub max: f64,
}

impl TemperatureRange {
    pub fn new(min: f64, max: f64) -> Self {
        Self { min, max }
    }

    pub fn contains(&self, temp: f64) -> bool {
        temp >= self.min && temp <= self.max
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CargoTypeConfig {
    pub cargo_type: CargoType,
    pub temperature_range: TemperatureRange,
    pub tolerance_range: TemperatureRange,
}

impl CargoTypeConfig {
    pub fn new(cargo_type: CargoType, temp_min: f64, temp_max: f64) -> Self {
        let tolerance_min = temp_min - 2.0;
        let tolerance_max = temp_max + 2.0;
        Self {
            cargo_type,
            temperature_range: TemperatureRange::new(temp_min, temp_max),
            tolerance_range: TemperatureRange::new(tolerance_min, tolerance_max),
        }
    }

    pub fn with_tolerance(
        cargo_type: CargoType,
        temp_min: f64,
        temp_max: f64,
        tolerance_min: f64,
        tolerance_max: f64,
    ) -> Self {
        Self {
            cargo_type,
            temperature_range: TemperatureRange::new(temp_min, temp_max),
            tolerance_range: TemperatureRange::new(tolerance_min, tolerance_max),
        }
    }

    pub fn check_temperature(&self, temp: f64) -> Option<AlertLevel> {
        if self.temperature_range.contains(temp) {
            None
        } else if self.tolerance_range.contains(temp) {
            Some(AlertLevel::Warning)
        } else {
            Some(AlertLevel::Critical)
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Shipment {
    pub id: Uuid,
    pub cargo_type: CargoType,
    pub vehicle_id: VehicleId,
    pub temperature_range: TemperatureRange,
    pub tolerance_range: TemperatureRange,
    pub status: ShipmentStatus,
    pub created_at: DateTime<Utc>,
    pub started_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
    pub is_broken_chain: bool,
    pub broken_chain_detected_at: Option<DateTime<Utc>>,
}

impl Shipment {
    pub fn new(
        cargo_type: CargoType,
        vehicle_id: VehicleId,
        temp_min: f64,
        temp_max: f64,
    ) -> Self {
        let tolerance_min = temp_min - 2.0;
        let tolerance_max = temp_max + 2.0;
        Self {
            id: Uuid::new_v4(),
            cargo_type,
            vehicle_id,
            temperature_range: TemperatureRange::new(temp_min, temp_max),
            tolerance_range: TemperatureRange::new(tolerance_min, tolerance_max),
            status: ShipmentStatus::Pending,
            created_at: Utc::now(),
            started_at: None,
            completed_at: None,
            is_broken_chain: false,
            broken_chain_detected_at: None,
        }
    }

    pub fn check_temperature(&self, temp: f64) -> Option<AlertLevel> {
        if self.temperature_range.contains(temp) {
            None
        } else if self.tolerance_range.contains(temp) {
            Some(AlertLevel::Warning)
        } else {
            Some(AlertLevel::Critical)
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TemperatureReading {
    pub id: Uuid,
    pub shipment_id: Uuid,
    pub temperature: f64,
    pub recorded_at: DateTime<Utc>,
    pub alert_level: Option<AlertLevel>,
}

impl TemperatureReading {
    pub fn new(shipment_id: Uuid, temperature: f64) -> Self {
        Self {
            id: Uuid::new_v4(),
            shipment_id,
            temperature,
            recorded_at: Utc::now(),
            alert_level: None,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Alert {
    pub id: Uuid,
    pub shipment_id: Uuid,
    pub level: AlertLevel,
    pub temperature: f64,
    pub recorded_at: DateTime<Utc>,
}

impl Alert {
    pub fn new(shipment_id: Uuid, level: AlertLevel, temperature: f64) -> Self {
        Self {
            id: Uuid::new_v4(),
            shipment_id,
            level,
            temperature,
            recorded_at: Utc::now(),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BrokenChainEvent {
    pub id: Uuid,
    pub shipment_id: Uuid,
    pub vehicle_id: VehicleId,
    pub cargo_type: CargoType,
    pub detected_at: DateTime<Utc>,
    pub first_out_of_range_at: DateTime<Utc>,
}

impl BrokenChainEvent {
    pub fn new(
        shipment_id: Uuid,
        vehicle_id: VehicleId,
        cargo_type: CargoType,
        first_out_of_range_at: DateTime<Utc>,
    ) -> Self {
        Self {
            id: Uuid::new_v4(),
            shipment_id,
            vehicle_id,
            cargo_type,
            detected_at: Utc::now(),
            first_out_of_range_at,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct VehicleMonthlyStats {
    pub vehicle_id: VehicleId,
    pub year: i32,
    pub month: u32,
    pub broken_chain_count: u32,
    pub total_shipments: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CargoTypeStats {
    pub cargo_type: CargoType,
    pub total_shipments: u32,
    pub rejected_shipments: u32,
    pub loss_rate: f64,
}
