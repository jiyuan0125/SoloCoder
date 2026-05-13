use serde::{Serialize, Deserialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PoolConfig {
    pub address: String,
    pub port: u16,
    pub max_connections: u32,
    pub min_idle: u32,
    pub connection_timeout_ms: u64,
    pub acquire_timeout_ms: u64,
    pub idle_timeout_ms: u64,
    pub max_lifetime_ms: u64,
    pub health_check_interval_ms: u64,
    pub max_create_per_second: u32,
    pub heartbeat_failed_threshold: u32,
}

impl Default for PoolConfig {
    fn default() -> Self {
        Self {
            address: String::from("127.0.0.1"),
            port: 8080,
            max_connections: 10,
            min_idle: 2,
            connection_timeout_ms: 5000,
            acquire_timeout_ms: 30000,
            idle_timeout_ms: 300000,
            max_lifetime_ms: 3600000,
            health_check_interval_ms: 30000,
            max_create_per_second: 10,
            heartbeat_failed_threshold: 2,
        }
    }
}

impl PoolConfig {
    pub fn new(address: String, port: u16) -> Self {
        Self {
            address,
            port,
            ..Default::default()
        }
    }
    
    pub fn with_max_connections(mut self, max: u32) -> Self {
        self.max_connections = max;
        self
    }
    
    pub fn with_min_idle(mut self, min: u32) -> Self {
        self.min_idle = min;
        self
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct PoolStats {
    pub name: String,
    pub active_connections: u32,
    pub idle_connections: u32,
    pub waiting_requests: u32,
    pub total_created: u64,
    pub total_destroyed: u64,
    pub state: PoolState,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum PoolState {
    Active,
    ShuttingDown,
    Closed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegisterPoolRequest {
    pub name: String,
    pub address: String,
    pub port: u16,
    #[serde(default)]
    pub max_connections: Option<u32>,
    #[serde(default)]
    pub min_idle: Option<u32>,
    #[serde(default)]
    pub connection_timeout_ms: Option<u64>,
    #[serde(default)]
    pub acquire_timeout_ms: Option<u64>,
    #[serde(default)]
    pub idle_timeout_ms: Option<u64>,
    #[serde(default)]
    pub max_lifetime_ms: Option<u64>,
}

impl RegisterPoolRequest {
    pub fn into_config(self) -> PoolConfig {
        let mut config = PoolConfig::new(self.address, self.port);
        if let Some(v) = self.max_connections { config.max_connections = v; }
        if let Some(v) = self.min_idle { config.min_idle = v; }
        if let Some(v) = self.connection_timeout_ms { config.connection_timeout_ms = v; }
        if let Some(v) = self.acquire_timeout_ms { config.acquire_timeout_ms = v; }
        if let Some(v) = self.idle_timeout_ms { config.idle_timeout_ms = v; }
        if let Some(v) = self.max_lifetime_ms { config.max_lifetime_ms = v; }
        config
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct AcquireResponse {
    pub connection_id: String,
    pub pool_name: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ConnectionHealth {
    Healthy,
    Unhealthy,
    Unknown,
}

impl Default for ConnectionHealth {
    fn default() -> Self {
        ConnectionHealth::Unknown
    }
}
