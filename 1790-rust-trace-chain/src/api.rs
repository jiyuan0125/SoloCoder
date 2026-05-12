use std::sync::Arc;
use tokio::sync::Mutex;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    Json,
    Router,
    routing::{get, post, delete},
};
use crate::model::{Span, CreateSpanRequest, EndSpanRequest, SearchByTagsRequest, CleanupByTimeRequest};
use crate::storage::TraceStorage;

type AppState = Arc<Mutex<TraceStorage>>;

pub fn create_router(state: AppState) -> Router {
    Router::new()
        .route("/spans", post(create_span))
        .route("/spans/end", post(end_span))
        .route("/traces/:trace_id", get(get_trace))
        .route("/search", post(search_by_tags))
        .route("/traces/:trace_id", delete(cleanup_trace))
        .route("/cleanup", post(cleanup_old_spans))
        .with_state(state)
}

async fn create_span(
    State(state): State<AppState>,
    Json(req): Json<CreateSpanRequest>,
) -> impl IntoResponse {
    let span = Span::new(
        req.trace_id,
        req.parent_span_id,
        req.operation_name,
        req.tags,
    );
    
    let span_clone = span.clone();
    let mut storage = state.lock().await;
    storage.add_span(span);
    
    (StatusCode::CREATED, Json(span_clone))
}

async fn end_span(
    State(state): State<AppState>,
    Json(req): Json<EndSpanRequest>,
) -> impl IntoResponse {
    let mut storage = state.lock().await;
    match storage.end_span(&req.span_id) {
        Some(span) => (StatusCode::OK, Json(serde_json::json!({ "span": span }))).into_response(),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({ "error": "Span not found" }))).into_response(),
    }
}

async fn get_trace(
    State(state): State<AppState>,
    Path(trace_id): Path<String>,
) -> impl IntoResponse {
    let storage = state.lock().await;
    match storage.get_trace(&trace_id) {
        Some(spans) => (StatusCode::OK, Json(serde_json::json!({
            "trace_id": trace_id,
            "spans": spans,
        }))).into_response(),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({ "error": "Trace not found" }))).into_response(),
    }
}

async fn search_by_tags(
    State(state): State<AppState>,
    Json(req): Json<SearchByTagsRequest>,
) -> impl IntoResponse {
    let storage = state.lock().await;
    let page = if req.page < 1 { 1 } else { req.page };
    let page_size = if req.page_size < 1 { 20 } else { req.page_size };
    
    let result = storage.search_by_tags(&req.tags, page, page_size);
    (StatusCode::OK, Json(result)).into_response()
}

async fn cleanup_trace(
    State(state): State<AppState>,
    Path(trace_id): Path<String>,
) -> impl IntoResponse {
    let mut storage = state.lock().await;
    let removed = storage.cleanup_trace(&trace_id);
    
    if removed > 0 {
        (StatusCode::OK, Json(serde_json::json!({
            "trace_id": trace_id,
            "spans_removed": removed,
        }))).into_response()
    } else {
        (StatusCode::NOT_FOUND, Json(serde_json::json!({ "error": "Trace not found" }))).into_response()
    }
}

async fn cleanup_old_spans(
    State(state): State<AppState>,
    Json(req): Json<CleanupByTimeRequest>,
) -> impl IntoResponse {
    let mut storage = state.lock().await;
    let removed = storage.cleanup_old_spans(req.hours);
    
    (StatusCode::OK, Json(serde_json::json!({
        "spans_removed": removed,
        "hours": req.hours.unwrap_or(24),
    }))).into_response()
}
