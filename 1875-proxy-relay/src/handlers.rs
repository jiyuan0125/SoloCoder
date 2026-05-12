use axum::{
    extract::{Path, State},
    http::{header, HeaderMap, HeaderValue, Request, StatusCode},
    response::Response,
    Json,
};
use http_body_util::BodyExt;
use reqwest::header::{HeaderMap as ReqwestHeaderMap, HeaderName as ReqwestHeaderName, HeaderValue as ReqwestHeaderValue};
use serde::{Deserialize, Serialize};
use tokio::time::timeout;
use tracing::{debug, error, info, warn};
use uuid::Uuid;

use crate::state::{AppState, Backend, CreateBackendRequest, Stats};

#[derive(Debug, Deserialize)]
pub struct UpdateConfigRequest {
    pub timeout_seconds: u64,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

pub async fn create_backend(
    State(state): State<AppState>,
    Json(req): Json<CreateBackendRequest>,
) -> Result<Json<Backend>, (StatusCode, Json<ErrorResponse>)> {
    if req.path_prefix.is_empty() {
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "path_prefix cannot be empty".to_string(),
            }),
        ));
    }
    if req.target_url.is_empty() {
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "target_url cannot be empty".to_string(),
            }),
        ));
    }

    let backend = state.add_backend(req).await;
    info!("Created backend: {} -> {}", backend.path_prefix, backend.target_url);
    Ok(Json(backend))
}

pub async fn get_backends(State(state): State<AppState>) -> Json<Vec<Backend>> {
    let backends = state.get_backends().await;
    Json(backends)
}

pub async fn delete_backend(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<StatusCode, (StatusCode, Json<ErrorResponse>)> {
    if state.remove_backend(id).await {
        info!("Removed backend: {}", id);
        Ok(StatusCode::NO_CONTENT)
    } else {
        Err((
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: format!("Backend with id {} not found", id),
            }),
        ))
    }
}

pub async fn update_config(
    State(state): State<AppState>,
    Json(req): Json<UpdateConfigRequest>,
) -> Result<StatusCode, (StatusCode, Json<ErrorResponse>)> {
    if req.timeout_seconds == 0 {
        return Err((
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "timeout_seconds must be greater than 0".to_string(),
            }),
        ));
    }
    state.update_timeout(req.timeout_seconds).await;
    info!("Updated timeout to {} seconds", req.timeout_seconds);
    Ok(StatusCode::NO_CONTENT)
}

pub async fn get_stats(State(state): State<AppState>) -> Json<Vec<Stats>> {
    let stats = state.get_stats().await;
    Json(stats)
}

pub async fn proxy(
    State(state): State<AppState>,
    headers: HeaderMap,
    mut req: Request<axum::body::Body>,
) -> Response<axum::body::Body> {
    let request_id = Uuid::new_v4();
    let path = req.uri().path().to_string();
    let method = req.method().clone();

    info!("[{}] {} {}", request_id, method, path);

    let backend = match state.find_backend_for_path(&path).await {
        Some(b) => b,
        None => {
            warn!("[{}] No backend found for path: {}", request_id, path);
            return build_error_response(
                StatusCode::NOT_FOUND,
                &format!("No backend configured for path: {}", path),
            );
        }
    };

    let target_path = path.strip_prefix(&backend.path_prefix).unwrap_or(&path);
    let target_url = format!(
        "{}/{}",
        backend.target_url.trim_end_matches('/'),
        target_path.trim_start_matches('/')
    );

    if let Some(query) = req.uri().query() {
        debug!("[{}] Including query: {}", request_id, query);
    }

    let client_timeout = state.get_timeout().await;
    state.increment_request_count(backend.id).await;

    let request_future = async move {
        let mut builder = state.http_client.request(method.clone(), &target_url);

        let x_forwarded_for = build_x_forwarded_for(&headers);
        let mut forward_headers = convert_headers(&headers);
        forward_headers.insert(
            ReqwestHeaderName::from_static("x-request-id"),
            ReqwestHeaderValue::from_str(&request_id.to_string()).unwrap(),
        );
        if let Some(xff) = x_forwarded_for {
            forward_headers.insert(
                ReqwestHeaderName::from_static("x-forwarded-for"),
                ReqwestHeaderValue::from_str(&xff).unwrap(),
            );
        }
        forward_headers.remove(ReqwestHeaderName::from_static("content-length"));
        builder = builder.headers(forward_headers);

        let request_body = req.into_body();
        let body_bytes = match request_body.collect().await {
            Ok(collected) => collected.to_bytes(),
            Err(e) => {
                error!("[{}] Failed to read request body: {}", request_id, e);
                return build_error_response(
                    StatusCode::BAD_REQUEST,
                    "Failed to read request body",
                );
            }
        };
        builder = builder.body(body_bytes);

        let response = match builder.send().await {
            Ok(resp) => resp,
            Err(e) => {
                error!("[{}] Error forwarding request: {}", request_id, e);
                return build_error_response(
                    StatusCode::BAD_GATEWAY,
                    &format!("Backend error: {}", e),
                );
            }
        };

        let status = response.status();
        let response_headers = response.headers().clone();
        let response_body = response.bytes_stream();

        info!("[{}] Backend returned status: {}", request_id, status);

        let mut builder = Response::builder().status(status.as_u16());
        for (name, value) in response_headers.iter() {
            if name.as_str().to_lowercase() == "content-length" {
                continue;
            }
            if let Ok(val) = HeaderValue::from_bytes(value.as_bytes()) {
                builder = builder.header(name.as_str(), val);
            }
        }

        let body = axum::body::Body::from_stream(response_body);
        match builder.body(body) {
            Ok(resp) => resp,
            Err(e) => {
                error!("[{}] Failed to build response: {}", request_id, e);
                build_error_response(StatusCode::INTERNAL_SERVER_ERROR, "Failed to build response")
            }
        }
    };

    match timeout(client_timeout, request_future).await {
        Ok(response) => response,
        Err(_) => {
            error!("[{}] Request timed out after {} seconds", request_id, client_timeout.as_secs());
            build_error_response(StatusCode::GATEWAY_TIMEOUT, "Request timed out")
        }
    }
}

fn build_x_forwarded_for(headers: &HeaderMap) -> Option<String> {
    let client_ip = headers
        .get(header::FORWARDED)
        .and_then(|v| v.to_str().ok())
        .or_else(|| {
            headers
                .get("x-forwarded-for")
                .and_then(|v| v.to_str().ok())
        })
        .and_then(|s| s.split(',').next().map(|s| s.trim().to_string()))
        .unwrap_or_else(|| "unknown".to_string());

    Some(client_ip)
}

fn convert_headers(headers: &HeaderMap) -> ReqwestHeaderMap {
    let mut result = ReqwestHeaderMap::new();
    for (name, value) in headers.iter() {
        if let Ok(req_name) = ReqwestHeaderName::from_bytes(name.as_str().as_bytes()) {
            if let Ok(req_value) = ReqwestHeaderValue::from_bytes(value.as_bytes()) {
                result.insert(req_name, req_value);
            }
        }
    }
    result
}

fn build_error_response(status: StatusCode, message: &str) -> Response<axum::body::Body> {
    let body = serde_json::json!({ "error": message });
    Response::builder()
        .status(status)
        .header(header::CONTENT_TYPE, "application/json")
        .body(axum::body::Body::from(serde_json::to_vec(&body).unwrap()))
        .unwrap()
}
