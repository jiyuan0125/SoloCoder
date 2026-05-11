use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum CompartmentSize {
    Small,
    Medium,
    Large,
}

impl CompartmentSize {
    pub fn can_fit(&self, package_size: &PackageSize) -> bool {
        match (self, package_size) {
            (CompartmentSize::Small, PackageSize::Small) => true,
            (CompartmentSize::Medium, PackageSize::Small | PackageSize::Medium) => true,
            (CompartmentSize::Large, _) => true,
            _ => false,
        }
    }

    pub fn is_better_than(&self, other: &CompartmentSize, required: &PackageSize) -> bool {
        if !self.can_fit(required) || !other.can_fit(required) {
            return false;
        }
        self < other
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum PackageSize {
    Small,
    Medium,
    Large,
}

impl PackageSize {
    pub fn minimal_compartment(&self) -> CompartmentSize {
        match self {
            PackageSize::Small => CompartmentSize::Small,
            PackageSize::Medium => CompartmentSize::Medium,
            PackageSize::Large => CompartmentSize::Large,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Package {
    pub id: String,
    pub size: PackageSize,
    pub phone: String,
    pub courier_name: String,
    pub compartment_id: String,
    pub pickup_code: String,
    pub deposited_at: DateTime<Utc>,
    pub pickup_attempts: u32,
    pub locked_until: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Compartment {
    pub id: String,
    pub size: CompartmentSize,
    pub locker_id: String,
    pub current_package: Option<Package>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Locker {
    pub id: String,
    pub name: String,
    pub compartments: Vec<Compartment>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PickupResult {
    pub package: Package,
    pub fee: f64,
    pub hours: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DepositRequest {
    pub locker_id: String,
    pub package_size: PackageSize,
    pub phone: String,
    pub courier_name: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DepositResponse {
    pub locker_id: String,
    pub compartment_id: String,
    pub pickup_code: String,
    pub deposited_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PickupRequest {
    pub pickup_code: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PickupResponse {
    pub locker_id: String,
    pub compartment_id: String,
    pub package_id: String,
    pub phone: String,
    pub deposited_at: DateTime<Utc>,
    pub picked_at: DateTime<Utc>,
    pub storage_hours: u64,
    pub fee: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LockerInfo {
    pub id: String,
    pub name: String,
    pub total_compartments: usize,
    pub available_compartments: usize,
    pub available_small: usize,
    pub available_medium: usize,
    pub available_large: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct LockerDetails {
    pub id: String,
    pub name: String,
    pub compartments: Vec<CompartmentDetails>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CompartmentDetails {
    pub id: String,
    pub size: CompartmentSize,
    pub occupied: bool,
    pub package: Option<PackageInfo>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PackageInfo {
    pub id: String,
    pub phone: String,
    pub courier_name: String,
    pub pickup_code: String,
    pub deposited_at: DateTime<Utc>,
}
