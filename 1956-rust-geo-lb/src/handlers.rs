use crate::app_state::AppState;
use crate::models::{
    BackendStatus, StatsResponse, ZoneConfig, ZoneStats, ZoneStatus, ZonesResponse,
};
use axum::extract::{Request, State};
use axum::http::StatusCode;
use axum::response::{IntoResponse, Json, Response};
use std::sync::Arc;

pub async fn handle_post_zones(
    State(state): State<Arc<AppState>>,
    Json(config): Json<ZoneConfig>,
) -> Response {
    if config.name.is_empty() {
        return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "Zone name cannot be empty"})),
        )
            .into_response();
    }

    if config.cidrs.is_empty() {
        return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "At least one CIDR is required"})),
        )
            .into_response();
    }

    if config.backends.is_empty() {
        return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "At least one backend is required"})),
        )
            .into_response();
    }

    match state.add_zone(config).await {
        Ok(_) => (
            StatusCode::CREATED,
            Json(serde_json::json!({"status": "ok"})),
        )
            .into_response(),
        Err(e) => (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": e})),
        )
            .into_response(),
    }
}

pub async fn handle_get_zones(State(state): State<Arc<AppState>>) -> Response {
    let zones = state.get_all_zones().await;

    let mut zone_statuses = Vec::new();

    for zone in zones {
        let request_count = zone.request_count.load(std::sync::atomic::Ordering::SeqCst);
        let total_latency = zone
            .total_latency_ms
            .load(std::sync::atomic::Ordering::SeqCst);

        let avg_latency = if request_count > 0 {
            Some(total_latency as f64 / request_count as f64)
        } else {
            None
        };

        let backends: Vec<BackendStatus> = zone
            .backends
            .iter()
            .map(|b| BackendStatus {
                address: b.address.clone(),
                healthy: b.healthy,
                consecutive_failures: b.consecutive_failures,
                last_checked: b.last_checked.map(|t| {
                    t.duration_since(std::time::Instant::now()).as_secs() as i64 * -1
                }),
            })
            .collect();

        zone_statuses.push(ZoneStatus {
            name: zone.name.clone(),
            healthy: zone.is_healthy(),
            degraded: zone.is_degraded(),
            backends,
            request_count,
            total_latency_ms: total_latency,
            avg_latency_ms: avg_latency,
        });
    }

    (StatusCode::OK, Json(ZonesResponse { zones: zone_statuses })).into_response()
}

pub async fn handle_get_stats(State(state): State<Arc<AppState>>) -> Response {
    let zones = state.get_all_zones().await;

    let mut stats = Vec::new();

    for zone in zones {
        let request_count = zone.request_count.load(std::sync::atomic::Ordering::SeqCst);
        let total_latency = zone
            .total_latency_ms
            .load(std::sync::atomic::Ordering::SeqCst);

        let avg_latency = if request_count > 0 {
            Some(total_latency as f64 / request_count as f64)
        } else {
            None
        };

        stats.push(ZoneStats {
            name: zone.name.clone(),
            request_count,
            avg_latency_ms: avg_latency,
            healthy: zone.is_healthy(),
            degraded: zone.is_degraded(),
        });
    }

    (StatusCode::OK, Json(StatsResponse { zones: stats })).into_response()
}

pub async fn handle_proxy(
    State((state, client)): State<(Arc<AppState>, reqwest::Client)>,
    request: Request,
) -> Response {
    if request.uri().path() == "/health" {
        return (StatusCode::OK, Json(serde_json::json!({"status": "ok"}))).into_response();
    }

    crate::proxy::proxy_request(state, request, &client).await
}
