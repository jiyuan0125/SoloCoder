mod auth;
mod health;
mod router;
mod trace;

use std::sync::Arc;
use std::time::Instant;

use auth::ApiKeyRegistry;
use axum::{
    body::Body,
    body::to_bytes,
    extract::{Path, State},
    http::{HeaderMap, HeaderValue, Request, StatusCode},
    response::{IntoResponse, Response},
    routing::{delete, get, post},
    Json, Router,
};
use chrono::Utc;
use health::HealthChecker;
use router::{AuthStrategy, CreateRouteRequest, MatchType, RouteStore};
use serde::Serialize;
use tokio::net::TcpListener;
use trace::{TraceRecord, TraceStore};
use uuid::Uuid;

#[derive(Debug, Clone)]
struct AppState {
    routes: RouteStore,
    api_keys: ApiKeyRegistry,
    traces: TraceStore,
    health: HealthChecker,
}

#[derive(Debug, Serialize)]
struct CreateRouteResponse {
    id: Uuid,
    path: String,
    match_type: MatchType,
    backend_url: String,
    auth_strategy: AuthStrategy,
}

#[derive(Debug, Serialize)]
struct CreateKeyResponse {
    api_key: String,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

async fn create_route(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateRouteRequest>,
) -> impl IntoResponse {
    let route = state.routes.add(req).await;
    state.health.register_backend(&route.backend_url).await;

    (
        StatusCode::CREATED,
        Json(CreateRouteResponse {
            id: route.id,
            path: route.path,
            match_type: route.match_type,
            backend_url: route.backend_url,
            auth_strategy: route.auth_strategy,
        }),
    )
}

async fn delete_route(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    if state.routes.remove(id).await {
        StatusCode::NO_CONTENT.into_response()
    } else {
        (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: "Route not found".to_string(),
            }),
        )
            .into_response()
    }
}

async fn create_api_key(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let key = state.api_keys.create_key().await;
    Json(CreateKeyResponse { api_key: key })
}

async fn get_trace_by_id(
    State(state): State<Arc<AppState>>,
    Path(trace_id): Path<Uuid>,
) -> impl IntoResponse {
    match state.traces.get_by_trace_id(trace_id).await {
        Some(record) => Json(record).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: "Trace not found".to_string(),
            }),
        )
            .into_response(),
    }
}

async fn get_traces_by_service(
    State(state): State<Arc<AppState>>,
    Path(service_name): Path<String>,
) -> impl IntoResponse {
    let traces = state.traces.get_by_service(&service_name).await;
    Json(traces)
}

fn extract_service_name(backend_url: &str) -> String {
    backend_url
        .replace("http://", "")
        .replace("https://", "")
        .split(':')
        .next()
        .unwrap_or(backend_url)
        .to_string()
}

async fn proxy_request(
    State(state): State<Arc<AppState>>,
    headers: HeaderMap,
    req: Request<Body>,
) -> impl IntoResponse {
    let trace_id = Uuid::new_v4();
    let start_time = Instant::now();
    let path = req.uri().path().to_string();

    let matched_route = match state.routes.match_route(&path).await {
        Some(route) => route,
        None => {
            return (
                StatusCode::NOT_FOUND,
                Json(ErrorResponse {
                    error: "Route not found".to_string(),
                }),
            )
                .into_response();
        }
    };

    if !state.health.is_healthy(&matched_route.backend_url).await {
        let duration_ms = start_time.elapsed().as_millis() as u64;
        let record = TraceRecord {
            trace_id,
            service_name: extract_service_name(&matched_route.backend_url),
            status_code: 502,
            duration_ms,
            timestamp: Utc::now().timestamp(),
        };
        state.traces.record(record).await;

        return (
            StatusCode::BAD_GATEWAY,
            Json(ErrorResponse {
                error: "Backend service unavailable".to_string(),
            }),
        )
            .into_response();
    }

    if matched_route.auth_strategy == AuthStrategy::ApiKey {
        let api_key = headers
            .get("x-api-key")
            .and_then(|h| h.to_str().ok())
            .map(|s| s.to_string());

        match api_key {
            Some(key) => {
                if !state.api_keys.validate_key(&key).await {
                    return (
                        StatusCode::UNAUTHORIZED,
                        Json(ErrorResponse {
                            error: "Invalid API key".to_string(),
                        }),
                    )
                        .into_response();
                }
            }
            None => {
                return (
                    StatusCode::UNAUTHORIZED,
                    Json(ErrorResponse {
                        error: "Missing X-API-Key header".to_string(),
                    }),
                )
                    .into_response();
            }
        }
    }

    let client = reqwest::Client::new();
    let method = req.method().clone();
    let uri = req.uri().clone();
    let query = uri.query().map(|q| format!("?{}", q)).unwrap_or_default();

    let backend_url = matched_route.backend_url.trim_end_matches('/');
    let proxy_path = if matched_route.match_type == MatchType::Prefix {
        let prefix = matched_route.path.trim_end_matches('*');
        let stripped = path.strip_prefix(prefix).unwrap_or(&path);
        format!("{}{}{}", backend_url, stripped, query)
    } else {
        format!("{}{}{}", backend_url, path, query)
    };

    let mut proxy_req = client.request(method.clone(), &proxy_path);

    for (name, value) in headers.iter() {
        if name != "host" {
            if let Ok(v) = value.to_str() {
                proxy_req = proxy_req.header(name.as_str(), v);
            }
        }
    }

    let body_bytes = match to_bytes(req.into_body(), 1024 * 1024 * 10).await {
        Ok(bytes) => bytes,
        Err(_) => {
            let duration_ms = start_time.elapsed().as_millis() as u64;
            let record = TraceRecord {
                trace_id,
                service_name: extract_service_name(&matched_route.backend_url),
                status_code: 500,
                duration_ms,
                timestamp: Utc::now().timestamp(),
            };
            state.traces.record(record).await;

            return (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse {
                    error: "Failed to read request body".to_string(),
                }),
            )
                .into_response();
        }
    };

    if !body_bytes.is_empty() {
        proxy_req = proxy_req.body(body_bytes);
    }

    let response = match proxy_req.send().await {
        Ok(resp) => resp,
        Err(_) => {
            let duration_ms = start_time.elapsed().as_millis() as u64;
            let record = TraceRecord {
                trace_id,
                service_name: extract_service_name(&matched_route.backend_url),
                status_code: 502,
                duration_ms,
                timestamp: Utc::now().timestamp(),
            };
            state.traces.record(record).await;

            return (
                StatusCode::BAD_GATEWAY,
                Json(ErrorResponse {
                    error: "Failed to connect to backend".to_string(),
                }),
            )
                .into_response();
        }
    };

    let status = response.status();
    let duration_ms = start_time.elapsed().as_millis() as u64;

    let record = TraceRecord {
        trace_id,
        service_name: extract_service_name(&matched_route.backend_url),
        status_code: status.as_u16(),
        duration_ms,
        timestamp: Utc::now().timestamp(),
    };
    state.traces.record(record).await;

    let mut builder = Response::builder().status(status);

    for (name, value) in response.headers() {
        if let Ok(v) = value.to_str() {
            builder = builder.header(name.as_str(), v);
        }
    }

    let body = match response.bytes().await {
        Ok(bytes) => Body::from(bytes),
        Err(_) => Body::empty(),
    };

    let trace_id_header = HeaderValue::from_str(&trace_id.to_string()).unwrap();
    builder = builder.header("X-Trace-ID", trace_id_header);

    builder.body(body).unwrap_or_else(|_| Response::new(Body::empty()))
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let routes = RouteStore::new();
    let api_keys = ApiKeyRegistry::new();
    let traces = TraceStore::new();
    let health = HealthChecker::new();

    health.start().await;

    let app_state = Arc::new(AppState {
        routes: routes.clone(),
        api_keys: api_keys.clone(),
        traces: traces.clone(),
        health: health.clone(),
    });

    let admin_routes = Router::new()
        .route("/routes", post(create_route))
        .route("/routes/:id", delete(delete_route))
        .route("/keys", post(create_api_key))
        .route("/traces/:trace_id", get(get_trace_by_id))
        .route("/traces/service/:service_name", get(get_traces_by_service))
        .with_state(app_state.clone());

    let proxy_route = Router::new()
        .fallback(proxy_request)
        .with_state(app_state.clone());

    let app = Router::new().merge(admin_routes).merge(proxy_route);

    let port = std::env::var("PORT").unwrap_or_else(|_| "3000".to_string());
    let addr = format!("0.0.0.0:{}", port);

    let listener = TcpListener::bind(&addr).await.expect("Failed to bind");
    println!("API Gateway listening on {}", addr);

    axum::serve(listener, app).await.expect("Failed to serve");
}
