use std::collections::HashMap;
use std::time::Duration;

use axum::{
    extract::State,
    http::StatusCode,
    response::IntoResponse,
    Json,
};
use serde::{Deserialize, Serialize};

use crate::pool::{ConnectionPool, PoolStats, PoolConfig};

#[derive(Clone)]
pub struct AppState {
    pub pool: ConnectionPool,
}

#[derive(Serialize)]
pub struct StatsResponse {
    pub total_active: usize,
    pub total_idle: usize,
    pub total_borrowed: usize,
    pub total_borrow_count: u64,
    pub total_create_failures: u64,
    pub per_backend: HashMap<String, BackendStatsResponse>,
}

#[derive(Serialize)]
pub struct BackendStatsResponse {
    pub active: usize,
    pub idle: usize,
    pub borrowed: usize,
    pub borrow_count: u64,
    pub create_failures: u64,
}

#[derive(Serialize)]
pub struct ConfigResponse {
    pub max_connections: usize,
    pub idle_timeout_secs: u64,
    pub health_check_interval_secs: u64,
    pub graceful_shutdown_timeout_secs: u64,
    pub max_health_failures: u32,
}

#[derive(Deserialize)]
pub struct UpdateConfigRequest {
    pub max_connections: Option<usize>,
    pub idle_timeout_secs: Option<u64>,
    pub health_check_interval_secs: Option<u64>,
    pub graceful_shutdown_timeout_secs: Option<u64>,
    pub max_health_failures: Option<u32>,
}

#[derive(Serialize)]
pub struct HealthResponse {
    pub status: &'static str,
}

impl From<PoolStats> for StatsResponse {
    fn from(stats: PoolStats) -> Self {
        let per_backend = stats
            .per_backend
            .into_iter()
            .map(|(k, v)| (k, BackendStatsResponse {
                active: v.active,
                idle: v.idle,
                borrowed: v.borrowed,
                borrow_count: v.borrow_count,
                create_failures: v.create_failures,
            }))
            .collect();

        Self {
            total_active: stats.total_active,
            total_idle: stats.total_idle,
            total_borrowed: stats.total_borrowed,
            total_borrow_count: stats.total_borrow_count,
            total_create_failures: stats.total_create_failures,
            per_backend,
        }
    }
}

impl From<PoolConfig> for ConfigResponse {
    fn from(config: PoolConfig) -> Self {
        Self {
            max_connections: config.max_connections,
            idle_timeout_secs: config.idle_timeout.as_secs(),
            health_check_interval_secs: config.health_check_interval.as_secs(),
            graceful_shutdown_timeout_secs: config.graceful_shutdown_timeout.as_secs(),
            max_health_failures: config.max_health_failures,
        }
    }
}

pub async fn health_handler(
    State(_state): State<AppState>,
) -> impl IntoResponse {
    Json(HealthResponse { status: "ok" })
}

pub async fn stats_handler(
    State(state): State<AppState>,
) -> impl IntoResponse {
    let stats = state.pool.get_stats();
    Json(StatsResponse::from(stats))
}

pub async fn config_handler(
    State(state): State<AppState>,
    Json(payload): Json<UpdateConfigRequest>,
) -> impl IntoResponse {
    let current = state.pool.get_config();

    let new_config = PoolConfig {
        max_connections: payload.max_connections.unwrap_or(current.max_connections),
        idle_timeout: payload
            .idle_timeout_secs
            .map(Duration::from_secs)
            .unwrap_or(current.idle_timeout),
        health_check_interval: payload
            .health_check_interval_secs
            .map(Duration::from_secs)
            .unwrap_or(current.health_check_interval),
        graceful_shutdown_timeout: payload
            .graceful_shutdown_timeout_secs
            .map(Duration::from_secs)
            .unwrap_or(current.graceful_shutdown_timeout),
        max_health_failures: payload.max_health_failures.unwrap_or(current.max_health_failures),
    };

    state.pool.update_config(new_config.clone());

    (StatusCode::OK, Json(ConfigResponse::from(new_config)))
}
