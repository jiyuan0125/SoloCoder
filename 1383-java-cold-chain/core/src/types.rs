use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ShipmentStatus {
    Pending,
    InTransit,
    PendingAcceptance,
    Completed,
    Rejected,
}

impl ShipmentStatus {
    pub fn can_transition_to(&self, next: &ShipmentStatus) -> bool {
        match (self, next) {
            (ShipmentStatus::Pending, ShipmentStatus::InTransit) => true,
            (ShipmentStatus::InTransit, ShipmentStatus::PendingAcceptance) => true,
            (ShipmentStatus::PendingAcceptance, ShipmentStatus::Completed) => true,
            (ShipmentStatus::PendingAcceptance, ShipmentStatus::Rejected) => true,
            _ => false,
        }
    }
}

impl std::fmt::Display for ShipmentStatus {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ShipmentStatus::Pending => write!(f, "Pending"),
            ShipmentStatus::InTransit => write!(f, "InTransit"),
            ShipmentStatus::PendingAcceptance => write!(f, "PendingAcceptance"),
            ShipmentStatus::Completed => write!(f, "Completed"),
            ShipmentStatus::Rejected => write!(f, "Rejected"),
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum AlertLevel {
    Warning,
    Critical,
}

impl std::fmt::Display for AlertLevel {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            AlertLevel::Warning => write!(f, "Warning"),
            AlertLevel::Critical => write!(f, "Critical"),
        }
    }
}

pub type CargoType = String;
pub type VehicleId = String;
