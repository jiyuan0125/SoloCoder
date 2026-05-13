mod types;
mod auth;
mod routes;
mod aggregate;
mod stats;

use std::sync::Arc;
use std::time::Duration;

use axum::{
    routing::{get, post, delete},
    Router,
    extract::{State, Path, Json, Request},
    http::{HeaderMap, StatusCode},
    response::{IntoResponse, Response},
    body::Body,
};
use tokio::sync::RwLock;
use tracing_subscriber;

use crate::types::*;
use crate::auth::AuthRegistry;
use crate::routes::RouteRegistry;
use crate::aggregate::AggregateSceneRegistry;
use crate::stats::StatsCollector;

#[derive(Clone)]
struct AppState {
    route_registry: Arc<RwLock<RouteRegistry>>,
    auth_registry: Arc<RwLock<AuthRegistry>>,
    aggregate_scenes: Arc<RwLock<AggregateSceneRegistry>>,
    stats: Arc<RwLock<StatsCollector>>,
    http_client: reqwest::Client,
    default_timeout: Duration,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let port = std::env::var("PORT").unwrap_or_else(|_| "8306".to_string());
    let default_timeout_secs: u64 = std::env::var("DEFAULT_TIMEOUT")
        .ok()
        .and_then(|v| v.parse().ok())
        .unwrap_or(5);

    let default_timeout = Duration::from_secs(default_timeout_secs);

    let state = AppState {
        route_registry: Arc::new(RwLock::new(RouteRegistry::new())),
        auth_registry: Arc::new(RwLock::new(AuthRegistry::new())),
        aggregate_scenes: Arc::new(RwLock::new(AggregateSceneRegistry::new())),
        stats: Arc::new(RwLock::new(StatsCollector::new())),
        http_client: reqwest::Client::builder()
            .timeout(default_timeout)
            .build()
            .unwrap(),
        default_timeout,
    };

    let app = Router::new()
        .route("/routes", post(add_route).get(list_routes))
        .route("/routes/:id", delete(delete_route))
        .route("/aggregates", post(add_aggregate_scene).get(list_aggregate_scenes))
        .route("/aggregate/:scene_name", get(handle_aggregate))
        .fallback(handle_proxy)
        .with_state(state);

    let listener = tokio::net::TcpListener::bind(format!("0.0.0.0:{}", port)).await.unwrap();
    tracing::info!("API Gateway listening on {}", listener.local_addr().unwrap());
    axum::serve(listener, app).await.unwrap();
}

async fn add_route(
    State(state): State<AppState>,
    Json(payload): Json<AddRouteRequest>,
) -> Result<Json<RouteResponse>, StatusCode> {
    let id = uuid::Uuid::new_v4().to_string();
    let route = Route {
        id: id.clone(),
        path: payload.path,
        backend: payload.backend,
        auth_policy: payload.auth_policy,
    };

    state.route_registry.write().await.add(route.clone());
    state.stats.write().await.record_route(&route.path);

    Ok(Json(RouteResponse { id, route }))
}

async fn list_routes(State(state): State<AppState>) -> Json<Vec<Route>> {
    Json(state.route_registry.read().await.list())
}

async fn delete_route(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> StatusCode {
    if state.route_registry.write().await.remove(&id) {
        StatusCode::OK
    } else {
        StatusCode::NOT_FOUND
    }
}

async fn add_aggregate_scene(
    State(state): State<AppState>,
    Json(payload): Json<AggregateScene>,
) -> Result<Json<AggregateScene>, StatusCode> {
    state.aggregate_scenes.write().await.add(payload.clone());
    Ok(Json(payload))
}

async fn list_aggregate_scenes(State(state): State<AppState>) -> Json<Vec<AggregateScene>> {
    Json(state.aggregate_scenes.read().await.list())
}

async fn handle_aggregate(
    State(state): State<AppState>,
    Path(scene_name): Path<String>,
    headers: HeaderMap,
) -> Response {
    let scene = state.aggregate_scenes.read().await.get(&scene_name).cloned();
    let scene = match scene {
        Some(s) => s,
        None => {
            state.stats.write().await.record_aggregate_error(&scene_name);
            return (StatusCode::NOT_FOUND, Json(serde_json::json!({
                "error": "scene not found"
            }))).into_response();
        }
    };

    state.stats.write().await.record_aggregate_start(&scene_name);

    let result = aggregate::execute_aggregate(
        scene,
        &state.http_client,
        state.default_timeout,
        &state.auth_registry,
        &headers,
    ).await;

    let _total_duration = state.stats.write().await.record_aggregate_complete(&scene_name);
    
    let response = match result {
        Ok(data) => {
            let mut json = serde_json::Map::new();
            for (field, res) in data {
                match res {
                    Ok(value) => {
                        json.insert(field, value);
                    }
                    Err(err) => {
                        json.insert(field, serde_json::json!({
                            "error": err
                        }));
                    }
                }
            }
            (StatusCode::OK, Json(json)).into_response()
        }
        Err(e) => {
            (StatusCode::INTERNAL_SERVER_ERROR, Json(serde_json::json!({
                "error": e
            }))).into_response()
        }
    };

    response
}

async fn handle_proxy(
    State(state): State<AppState>,
    headers: HeaderMap,
    req: Request<Body>,
) -> Response {
    let path = req.uri().path().to_string();
    let method = req.method().clone();
    
    let route = state.route_registry.read().await.get_by_path(&path).cloned();
    
    let route = match route {
        Some(r) => r,
        None => {
            return (StatusCode::NOT_FOUND, Json(serde_json::json!({
                "error": "route not found"
            }))).into_response();
        }
    };

    let auth_registry = state.auth_registry.read().await;
    if !auth_registry.validate(&route.auth_policy, &headers) {
        return (StatusCode::UNAUTHORIZED, Json(serde_json::json!({
            "error": "unauthorized"
        }))).into_response();
    }

    state.stats.write().await.record_route(&route.path);

    let backend_url = format!("{}{}", route.backend, req.uri().path_and_query().map(|pq| pq.as_str()).unwrap_or(""));
    
    let mut proxy_req_builder = state.http_client
        .request(method, &backend_url);
    
    for (name, value) in headers.iter() {
        if name != "host" {
            proxy_req_builder = proxy_req_builder.header(name, value);
        }
    }

    let body_bytes = match axum::body::to_bytes(req.into_body(), usize::MAX).await {
        Ok(b) => b,
        Err(e) => {
            return (StatusCode::BAD_REQUEST, Json(serde_json::json!({
                "error": format!("failed to read body: {}", e)
            }))).into_response();
        }
    };

    let proxy_req = proxy_req_builder.body(body_bytes).build().unwrap();

    let response = match state.http_client.execute(proxy_req).await {
        Ok(resp) => {
            let status = resp.status();
            let headers = resp.headers().clone();
            let body = resp.bytes().await.unwrap_or_default();
            
            let mut builder = Response::builder().status(status);
            for (name, value) in headers.iter() {
                builder = builder.header(name, value);
            }
            builder.body(Body::from(body)).unwrap()
        }
        Err(e) => {
            (StatusCode::BAD_GATEWAY, Json(serde_json::json!({
                "error": format!("backend error: {}", e)
            }))).into_response()
        }
    };

    response
}
