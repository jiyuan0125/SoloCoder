use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum CircuitState {
    Closed,
    Open,
    HalfOpen,
}

impl CircuitState {
    pub fn as_str(&self) -> &'static str {
        match self {
            CircuitState::Closed => "closed",
            CircuitState::Open => "open",
            CircuitState::HalfOpen => "half_open",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CircuitConfig {
    #[serde(default = "default_failure_threshold")]
    pub failure_threshold: f64,
    #[serde(default = "default_window_size")]
    pub window_size: usize,
    #[serde(default = "default_cooling_period")]
    pub cooling_period_secs: u64,
    #[serde(default)]
    pub notify_url: Option<String>,
    #[serde(default)]
    pub target_url: Option<String>,
}

fn default_failure_threshold() -> f64 {
    0.5
}

fn default_window_size() -> usize {
    20
}

fn default_cooling_period() -> u64 {
    30
}

impl Default for CircuitConfig {
    fn default() -> Self {
        CircuitConfig {
            failure_threshold: default_failure_threshold(),
            window_size: default_window_size(),
            cooling_period_secs: default_cooling_period(),
            notify_url: None,
            target_url: None,
        }
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct RequestRecord {
    pub timestamp: DateTime<Utc>,
    pub success: bool,
}

#[derive(Debug, Clone, Serialize)]
pub struct CircuitStatus {
    pub service_name: String,
    pub state: CircuitState,
    pub config: CircuitConfig,
    pub current_window: WindowStats,
    pub recent_requests: Vec<RequestRecord>,
    pub opened_at: Option<DateTime<Utc>>,
    pub last_state_change: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Default)]
pub struct WindowStats {
    pub successes: usize,
    pub failures: usize,
    pub total: usize,
    pub failure_rate: Option<f64>,
}

#[derive(Debug, Clone, Serialize)]
pub struct StateChangeNotification {
    pub service_name: String,
    pub old_state: CircuitState,
    pub new_state: CircuitState,
    pub reason: String,
    pub timestamp: DateTime<Utc>,
}

#[derive(Debug, Deserialize)]
pub struct CreateBreakerRequest {
    pub service_name: String,
    #[serde(default)]
    pub config: CircuitConfig,
}

#[derive(Debug, Deserialize)]
pub struct UpdateConfigRequest {
    #[serde(default)]
    pub failure_threshold: Option<f64>,
    #[serde(default)]
    pub window_size: Option<usize>,
    #[serde(default)]
    pub cooling_period_secs: Option<u64>,
    #[serde(default)]
    pub notify_url: Option<Option<String>>,
    #[serde(default)]
    pub target_url: Option<Option<String>>,
}

#[derive(Debug, Deserialize)]
pub struct ProxyRequest {
    #[serde(default)]
    pub body: Option<serde_json::Value>,
}
