use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum InstanceStatus {
    Registered,
    Healthy,
    Unhealthy,
}

impl ToString for InstanceStatus {
    fn to_string(&self) -> String {
        match self {
            InstanceStatus::Registered => "registered",
            InstanceStatus::Healthy => "healthy",
            InstanceStatus::Unhealthy => "unhealthy",
        }
        .to_string()
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServiceInstance {
    pub instance_id: Uuid,
    pub service_name: String,
    pub host: String,
    pub port: u16,
    pub metadata: HashMap<String, String>,
    pub status: InstanceStatus,
    pub last_heartbeat: Option<std::time::SystemTime>,
    pub registered_at: std::time::SystemTime,
    pub unhealthy_since: Option<std::time::SystemTime>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegisterInstanceRequest {
    pub host: String,
    pub port: u16,
    #[serde(default)]
    pub metadata: HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PatchMetadataRequest {
    #[serde(flatten)]
    pub updates: HashMap<String, Option<String>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegisterInstanceResponse {
    pub instance_id: Uuid,
    pub service_name: String,
    pub host: String,
    pub port: u16,
    pub metadata: HashMap<String, String>,
    pub status: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceInfo {
    pub instance_id: Uuid,
    pub service_name: String,
    pub host: String,
    pub port: u16,
    pub metadata: HashMap<String, String>,
    pub status: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PaginatedResponse<T> {
    pub items: Vec<T>,
    pub total: usize,
    pub page: usize,
    pub page_size: usize,
    pub total_pages: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SubscriptionRequest {
    pub service_name: Option<String>,
    pub callback_url: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Subscription {
    pub subscription_id: Uuid,
    pub service_name: Option<String>,
    pub callback_url: String,
    pub created_at: std::time::SystemTime,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct InstanceChangeNotification {
    pub event_type: String,
    pub service_name: String,
    pub instance: InstanceInfo,
    pub timestamp: std::time::SystemTime,
}
