use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum DestinationType {
    SameCity,
    SameProvince,
    NeighboringProvince,
    Remote,
}

impl DestinationType {
    pub fn price_per_kg(&self) -> f64 {
        match self {
            DestinationType::SameCity => 2.0,
            DestinationType::SameProvince => 4.0,
            DestinationType::NeighboringProvince => 6.0,
            DestinationType::Remote => 8.0,
        }
    }

    pub fn name(&self) -> &'static str {
        match self {
            DestinationType::SameCity => "同城",
            DestinationType::SameProvince => "省内",
            DestinationType::NeighboringProvince => "邻省",
            DestinationType::Remote => "偏远地区",
        }
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum CargoType {
    Normal,
    Fragile,
    Liquid,
}

impl CargoType {
    pub fn name(&self) -> &'static str {
        match self {
            CargoType::Normal => "普通货物",
            CargoType::Fragile => "易碎品",
            CargoType::Liquid => "液体货物",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Package {
    pub id: String,
    pub length_cm: f64,
    pub width_cm: f64,
    pub height_cm: f64,
    pub actual_weight_kg: f64,
    pub destination: DestinationType,
    pub cargo_type: CargoType,
    pub declared_value: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct FreightBreakdown {
    pub volumetric_weight: f64,
    pub chargeable_weight: f64,
    pub base_freight: f64,
    pub remote_surcharge: f64,
    pub insurance_fee: f64,
    pub total_freight: f64,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderStatus {
    Created,
    Cancelled,
    InTransit,
    Delivered,
}

impl OrderStatus {
    pub fn name(&self) -> &'static str {
        match self {
            OrderStatus::Created => "已创建",
            OrderStatus::Cancelled => "已取消",
            OrderStatus::InTransit => "运输中",
            OrderStatus::Delivered => "已送达",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderItem {
    pub package: Package,
    pub breakdown: FreightBreakdown,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: Uuid,
    pub items: Vec<OrderItem>,
    pub total_freight: f64,
    pub status: OrderStatus,
    pub created_at: DateTime<Utc>,
    pub cancelled_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreatePackageRequest {
    pub length_cm: f64,
    pub width_cm: f64,
    pub height_cm: f64,
    pub actual_weight_kg: f64,
    pub destination: DestinationType,
    pub cargo_type: CargoType,
    pub declared_value: Option<f64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderRequest {
    pub packages: Vec<CreatePackageRequest>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderResponse {
    pub id: Uuid,
    pub items: Vec<OrderItemResponse>,
    pub total_freight: f64,
    pub status: OrderStatus,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct OrderItemResponse {
    pub package_id: String,
    pub destination: DestinationType,
    pub cargo_type: CargoType,
    pub breakdown: FreightBreakdown,
}
