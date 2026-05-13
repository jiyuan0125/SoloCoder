use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "lowercase")]
pub enum ProbeParams {
    Http {
        url: String,
        expected_status: Option<u16>,
        expected_body: Option<String>,
        timeout_secs: Option<u64>,
    },
    Tcp {
        host: String,
        port: u16,
        timeout_secs: Option<u64>,
    },
    Dns {
        domain: String,
        timeout_secs: Option<u64>,
    },
    Script {
        command: String,
        timeout_secs: Option<u64>,
    },
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum ScheduleStrategy {
    FixedInterval {
        interval_secs: u64,
    },
    ExponentialBackoff {
        initial_interval_secs: u64,
        max_interval_secs: u64,
    },
    Cron {
        expression: String,
    },
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Probe {
    #[serde(default = "Uuid::new_v4")]
    pub id: Uuid,
    pub name: Option<String>,
    pub params: ProbeParams,
    pub schedule: ScheduleStrategy,
    #[serde(default = "default_weight")]
    pub weight: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateProbe {
    pub name: Option<String>,
    pub params: ProbeParams,
    pub schedule: ScheduleStrategy,
    #[serde(default = "default_weight")]
    pub weight: u32,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum HealthStatus {
    Healthy,
    Unhealthy,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProbeResult {
    pub probe_id: Uuid,
    pub status: HealthStatus,
    pub timestamp: i64,
    pub message: Option<String>,
    pub duration_ms: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AggregatedHealth {
    pub overall_status: OverallStatus,
    pub health_score: f32,
    pub total_weight: u32,
    pub healthy_weight: u32,
    #[serde(default = "default_threshold")]
    pub threshold: f32,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum OverallStatus {
    Healthy,
    Degraded,
    Unhealthy,
}

fn default_weight() -> u32 {
    100
}

fn default_threshold() -> f32 {
    0.8
}

impl Probe {
    pub fn from_create(create: CreateProbe) -> Self {
        Probe {
            id: Uuid::new_v4(),
            name: create.name,
            params: create.params,
            schedule: create.schedule,
            weight: create.weight,
        }
    }
}
