use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::response::IntoResponse;
use axum::routing::{get, post};
use axum::{Json, Router};
use chrono::DateTime;
use serde::Deserialize;
use std::sync::Arc;
use uuid::Uuid;

use crate::aggregate::AggregateManager;
use crate::models::{AppendEventRequest, AppendEventResponse, ErrorResponse, EventQuery};

#[derive(Clone)]
pub struct AppState {
    pub aggregate_manager: Arc<AggregateManager>,
}

#[derive(Debug, Deserialize)]
pub struct GetEventsQuery {
    pub event_type: Option<String>,
    pub start_time: Option<String>,
    pub end_time: Option<String>,
    pub offset: Option<u64>,
    pub limit: Option<u64>,
}

#[derive(Debug, Deserialize)]
pub struct GetStateAtVersionQuery {
    pub version: u64,
}

pub fn create_router(app_state: AppState) -> Router {
    Router::new()
        .route("/events", post(append_event))
        .route("/aggregates/:aggregate_id/events", get(get_events))
        .route("/aggregates/:aggregate_id/state", get(get_current_state))
        .route("/aggregates/:aggregate_id/state/version", get(get_state_at_version))
        .route("/health", get(health_check))
        .with_state(app_state)
}

async fn health_check() -> impl IntoResponse {
    (StatusCode::OK, Json(serde_json::json!({ "status": "ok" })))
}

async fn append_event(
    State(state): State<AppState>,
    Json(req): Json<AppendEventRequest>,
) -> impl IntoResponse {
    match state
        .aggregate_manager
        .append_event(
            req.aggregate_id,
            req.event_type,
            req.payload,
            req.expected_version,
        )
        .await
    {
        Ok(event) => (
            StatusCode::OK,
            Json(AppendEventResponse {
                event_id: event.event_id,
                aggregate_id: event.aggregate_id,
                version: event.version,
                timestamp: event.timestamp,
            }),
        )
            .into_response(),
        Err(current_version) => (
            StatusCode::CONFLICT,
            Json(ErrorResponse {
                error: "version_conflict".to_string(),
                message: format!(
                    "Expected version {} but current version is {}",
                    req.expected_version, current_version
                ),
            }),
        )
            .into_response(),
    }
}

async fn get_events(
    State(state): State<AppState>,
    Path(aggregate_id): Path<Uuid>,
    Query(query): Query<GetEventsQuery>,
) -> impl IntoResponse {
    let start_time = query
        .start_time
        .and_then(|t| DateTime::parse_from_rfc3339(&t).ok())
        .map(|t| t.with_timezone(&chrono::Utc));

    let end_time = query
        .end_time
        .and_then(|t| DateTime::parse_from_rfc3339(&t).ok())
        .map(|t| t.with_timezone(&chrono::Utc));

    let event_query = EventQuery {
        event_type: query.event_type,
        start_time,
        end_time,
        offset: query.offset,
        limit: query.limit,
    };

    let result = state
        .aggregate_manager
        .event_store()
        .get_events(aggregate_id, event_query)
        .await;

    (StatusCode::OK, Json(result)).into_response()
}

async fn get_current_state(
    State(state): State<AppState>,
    Path(aggregate_id): Path<Uuid>,
) -> impl IntoResponse {
    let current_state = state.aggregate_manager.get_state(aggregate_id).await;
    (StatusCode::OK, Json(current_state)).into_response()
}

async fn get_state_at_version(
    State(state): State<AppState>,
    Path(aggregate_id): Path<Uuid>,
    Query(query): Query<GetStateAtVersionQuery>,
) -> impl IntoResponse {
    let state_at_version = state
        .aggregate_manager
        .get_state_at_version(aggregate_id, query.version)
        .await;

    if state_at_version.version == 0 && query.version > 0 {
        return (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: "version_not_found".to_string(),
                message: format!("Version {} not found", query.version),
            }),
        )
            .into_response();
    }

    (StatusCode::OK, Json(state_at_version)).into_response()
}
