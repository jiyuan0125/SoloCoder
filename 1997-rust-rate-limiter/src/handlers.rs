use crate::store::RateLimiterEngine;
use crate::types::{
    IpThrottledState, PathPatternType, PathStatistics, PathThrottledInfo, RateLimitRule,
    RuleConfig,
};
use axum::extract::{Path, State};
use axum::http::StatusCode;
use axum::response::IntoResponse;
use axum::Json;
use serde::{Deserialize, Serialize};
use std::net::IpAddr;
use std::sync::Arc;
use uuid::Uuid;

#[derive(Clone)]
pub struct AppState {
    pub engine: Arc<RateLimiterEngine>,
}

#[derive(Serialize, Deserialize)]
pub struct RuleResponse {
    pub id: Uuid,
    pub algorithm: String,
    pub path_pattern: String,
    pub path_type: String,
    pub per_ip: bool,
    pub limit: u64,
    pub window_secs: u64,
    pub capacity: u64,
    pub refill_rate: u64,
    pub refill_interval_ms: u64,
}

impl From<RateLimitRule> for RuleResponse {
    fn from(rule: RateLimitRule) -> Self {
        RuleResponse {
            id: rule.id,
            algorithm: match rule.algorithm {
                crate::types::AlgorithmType::FixedWindow => "fixed_window".to_string(),
                crate::types::AlgorithmType::SlidingWindow => "sliding_window".to_string(),
                crate::types::AlgorithmType::TokenBucket => "token_bucket".to_string(),
            },
            path_pattern: rule.path_pattern,
            path_type: match rule.path_type {
                PathPatternType::Exact => "exact".to_string(),
                PathPatternType::Prefix => "prefix".to_string(),
            },
            per_ip: rule.per_ip,
            limit: rule.limit,
            window_secs: rule.window_secs,
            capacity: rule.capacity,
            refill_rate: rule.refill_rate,
            refill_interval_ms: rule.refill_interval.as_millis() as u64,
        }
    }
}

#[derive(Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

pub async fn create_rule(
    State(state): State<AppState>,
    Json(config): Json<RuleConfig>,
) -> impl IntoResponse {
    let rule: RateLimitRule = config.into();
    let rule_id = rule.id;
    state.engine.add_rule(rule);

    match state.engine.get_rule(&rule_id) {
        Some(rule) => (StatusCode::CREATED, Json(RuleResponse::from(rule))).into_response(),
        None => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse {
                error: "Failed to create rule".to_string(),
            }),
        )
            .into_response(),
    }
}

pub async fn get_all_rules(State(state): State<AppState>) -> impl IntoResponse {
    let rules: Vec<RuleResponse> = state
        .engine
        .get_all_rules()
        .into_iter()
        .map(RuleResponse::from)
        .collect();
    Json(rules)
}

pub async fn get_rule(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.engine.get_rule(&id) {
        Some(rule) => (StatusCode::OK, Json(RuleResponse::from(rule))).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: "Rule not found".to_string(),
            }),
        )
            .into_response(),
    }
}

pub async fn update_rule(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(config): Json<RuleConfig>,
) -> impl IntoResponse {
    let mut rule: RateLimitRule = config.into();
    rule.id = id;

    if state.engine.update_rule(rule.clone()) {
        (StatusCode::OK, Json(RuleResponse::from(rule))).into_response()
    } else {
        (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: "Rule not found".to_string(),
            }),
        )
            .into_response()
    }
}

pub async fn delete_rule(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    if state.engine.remove_rule(&id) {
        StatusCode::NO_CONTENT.into_response()
    } else {
        (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: "Rule not found".to_string(),
            }),
        )
            .into_response()
    }
}

pub async fn get_path_statistics(State(state): State<AppState>) -> impl IntoResponse {
    let stats: Vec<PathStatistics> = state.engine.get_stats_store().get_path_stats();
    Json(stats)
}

pub async fn get_ip_throttled_state(
    State(state): State<AppState>,
    Path(ip): Path<IpAddr>,
) -> impl IntoResponse {
    let ip_states = state.engine.get_stats_store().get_ip_throttled_state(ip);
    let paths: Vec<PathThrottledInfo> = ip_states
        .into_iter()
        .map(|(path, is_throttled, remaining, retry_after_secs)| PathThrottledInfo {
            path_pattern: path,
            is_throttled,
            remaining,
            retry_after_secs,
        })
        .collect();

    Json(IpThrottledState { ip, paths })
}

pub async fn get_all_throttled_ips(State(state): State<AppState>) -> impl IntoResponse {
    let ips: Vec<IpAddr> = state.engine.get_stats_store().get_all_throttled_ips();
    Json(ips)
}

pub async fn health_check() -> impl IntoResponse {
    Json(serde_json::json!({ "status": "ok" }))
}

pub async fn default_handler() -> impl IntoResponse {
    Json(serde_json::json!({
        "message": "Rate Limiter Service",
        "version": "0.1.0"
    }))
}
