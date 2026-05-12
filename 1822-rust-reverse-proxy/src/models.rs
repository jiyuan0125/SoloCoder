use regex::Regex;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AddBackendRequest {
    pub pattern: String,
    pub targets: Vec<Target>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Target {
    pub address: String,
    pub weight: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeleteBackendRequest {
    pub pattern: String,
    pub address: String,
}

#[derive(Debug, Clone)]
pub struct BackendInner {
    pub address: String,
    pub original_weight: u32,
    pub current_weight: u32,
    pub is_healthy: bool,
    pub consecutive_successes: u32,
    pub consecutive_failures: u32,
    pub total_requests: u64,
    pub pending_requests: u64,
    pub marked_for_deletion: bool,
}

impl BackendInner {
    pub fn new(address: String, weight: u32) -> Self {
        Self {
            address,
            original_weight: weight,
            current_weight: weight,
            is_healthy: true,
            consecutive_successes: 0,
            consecutive_failures: 0,
            total_requests: 0,
            pending_requests: 0,
            marked_for_deletion: false,
        }
    }
}

#[derive(Debug, Clone)]
pub struct Route {
    pub _pattern: String,
    pub regex: Regex,
}

#[derive(Debug, Clone, Serialize)]
pub struct BackendStatus {
    pub address: String,
    pub original_weight: u32,
    pub current_weight: u32,
    pub is_healthy: bool,
    pub total_requests: u64,
    pub pending_requests: u64,
    pub marked_for_deletion: bool,
}

#[derive(Debug, Clone, Serialize)]
pub struct RouteStatus {
    pub pattern: String,
    pub backends: Vec<BackendStatus>,
}
