use chrono::{DateTime, Utc};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub enum RatePeriod {
    Peak,
    Normal,
    Valley,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ChargerStatus {
    Idle,
    Charging,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Charger {
    pub id: Uuid,
    pub name: String,
    pub power_kw: u32,
    pub status: ChargerStatus,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum OrderStatus {
    Active,
    Completed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Order {
    pub id: Uuid,
    pub charger_id: Uuid,
    pub user_id: Uuid,
    pub start_time: DateTime<Utc>,
    pub end_time: Option<DateTime<Utc>>,
    pub duration_minutes: Option<u64>,
    pub energy_kwh: Option<Decimal>,
    pub total_cost: Option<Decimal>,
    pub status: OrderStatus,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
pub struct RateConfig {
    pub peak_rate: Decimal,
    pub normal_rate: Decimal,
    pub valley_rate: Decimal,
}

impl Default for RateConfig {
    fn default() -> Self {
        RateConfig {
            peak_rate: Decimal::new(15, 1),
            normal_rate: Decimal::new(10, 1),
            valley_rate: Decimal::new(5, 1),
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SystemState {
    pub chargers: HashMap<Uuid, Charger>,
    pub orders: HashMap<Uuid, Order>,
    pub rate_config: RateConfig,
}

impl SystemState {
    pub fn new() -> Self {
        SystemState {
            chargers: HashMap::new(),
            orders: HashMap::new(),
            rate_config: RateConfig::default(),
        }
    }
}

impl Default for SystemState {
    fn default() -> Self {
        Self::new()
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderRequest {
    pub charger_id: Uuid,
    pub user_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateOrderResponse {
    pub order: Order,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EndOrderRequest {
    pub order_id: Uuid,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EndOrderResponse {
    pub order: Order,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateChargerRequest {
    pub name: String,
    pub power_kw: u32,
}
