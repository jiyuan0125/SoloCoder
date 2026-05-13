use ipnet::IpNet;
use serde::{Deserialize, Serialize};
use std::net::IpAddr;
use std::sync::atomic::AtomicU64;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ZoneConfig {
    pub name: String,
    #[serde(rename = "cidrs")]
    pub cidrs: Vec<String>,
    #[serde(rename = "backends")]
    pub backends: Vec<String>,
    #[serde(default = "default_health_interval", rename = "healthIntervalSec")]
    pub health_interval_sec: u64,
}

fn default_health_interval() -> u64 {
    10
}

#[derive(Debug, Clone, Serialize)]
pub struct BackendStatus {
    pub address: String,
    pub healthy: bool,
    pub consecutive_failures: u32,
    pub last_checked: Option<i64>,
}

#[derive(Debug, Clone, Serialize)]
pub struct ZoneStatus {
    pub name: String,
    pub healthy: bool,
    pub degraded: bool,
    pub backends: Vec<BackendStatus>,
    pub request_count: u64,
    pub total_latency_ms: u64,
    pub avg_latency_ms: Option<f64>,
}

#[derive(Debug)]
pub struct BackendState {
    pub address: String,
    pub healthy: bool,
    pub consecutive_failures: u32,
    pub last_checked: Option<std::time::Instant>,
}

#[derive(Debug)]
pub struct ZoneState {
    pub name: String,
    pub cidrs: Vec<IpNet>,
    pub backends: Vec<BackendState>,
    pub health_interval_sec: u64,
    pub request_count: AtomicU64,
    pub total_latency_ms: AtomicU64,
}

impl ZoneState {
    pub fn is_healthy(&self) -> bool {
        self.backends.iter().all(|b| b.healthy)
    }

    pub fn is_degraded(&self) -> bool {
        !self.is_healthy() && self.backends.iter().any(|b| b.healthy)
    }

    pub fn has_any_healthy(&self) -> bool {
        self.backends.iter().any(|b| b.healthy)
    }

    pub fn get_healthy_backends(&self) -> Vec<&BackendState> {
        self.backends.iter().filter(|b| b.healthy).collect()
    }

    pub fn contains_ip(&self, ip: IpAddr) -> bool {
        self.cidrs.iter().any(|c| c.contains(&ip))
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct StatsResponse {
    pub zones: Vec<ZoneStats>,
}

#[derive(Debug, Clone, Serialize)]
pub struct ZoneStats {
    pub name: String,
    pub request_count: u64,
    pub avg_latency_ms: Option<f64>,
    pub healthy: bool,
    pub degraded: bool,
}

#[derive(Debug, Clone, Serialize)]
pub struct ZonesResponse {
    pub zones: Vec<ZoneStatus>,
}
