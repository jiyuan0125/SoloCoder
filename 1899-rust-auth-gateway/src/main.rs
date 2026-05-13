mod config;
mod store;
mod auth;
mod proxy;

use axum::body::Body;
use axum::extract::{Query, State};
use axum::http::{StatusCode, HeaderValue, HeaderName, Request};
use axum::response::{IntoResponse, Json, Response};
use axum::routing::{get, post, delete};
use axum::Router;
use rand::Rng;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};
use tracing::{info, error, warn};
use tracing_subscriber::layer::SubscriberExt;
use tracing_subscriber::util::SubscriberInitExt;

use crate::auth::{authenticate, build_www_authenticate, AuthResult, AuthenticatedUser};
use crate::config::{AppConfig, AuthMethod};
use crate::store::{ApiKeyMetadata, AppState, RouteEntry, RouteStore};

#[derive(Debug, Deserialize, Serialize)]
struct CreateApiKeyRequest {
    service_name: String,
    permissions: Vec<String>,
}

#[derive(Debug, Serialize)]
struct CreateApiKeyResponse {
    key: String,
    service_name: String,
    permissions: Vec<String>,
    created_at: u64,
}

#[derive(Debug, Serialize)]
struct ListApiKeyResponse {
    keys: Vec<ListApiKeyEntry>,
}

#[derive(Debug, Serialize)]
struct ListApiKeyEntry {
    key_prefix: String,
    service_name: String,
    permissions: Vec<String>,
    created_at: u64,
}

#[derive(Debug, Deserialize, Serialize)]
struct RotateJwtSecretRequest {
    new_secret: Option<String>,
}

#[derive(Debug, Serialize)]
struct HealthResponse {
    status: &'static str,
    timestamp: u64,
}

#[derive(Debug, Deserialize)]
struct DeleteApiKeyQuery {
    key: String,
}

fn generate_api_key() -> String {
    const CHARSET: &[u8] = b"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    let mut rng = rand::thread_rng();
    let bytes: Vec<u8> = (0..32).map(|_| CHARSET[rng.gen_range(0..CHARSET.len())]).collect();
    String::from_utf8_lossy(&bytes).to_string()
}

fn now_secs() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d: std::time::Duration| d.as_secs())
        .unwrap_or(0)
}

fn mask_key(key: &str) -> String {
    if key.len() <= 8 {
        "********".to_string()
    } else {
        format!("{}****{}", &key[..4], &key[key.len() - 4..])
    }
}

async fn require_admin_impl(
    state: &AppState,
    headers: &axum::http::HeaderMap,
) -> Result<AuthenticatedUser, Response> {
    let result = authenticate(state, headers, &[AuthMethod::Jwt]).await;
    match result {
        AuthResult::Success(user) => {
            if user.is_admin(&state.admin_user) {
                Ok(user)
            } else {
                warn!("Admin access denied for user: {}", user.identifier);
                Err((
                    StatusCode::FORBIDDEN,
                    Json(serde_json::json!({"error": "Forbidden", "message": "Admin role required"})),
                ).into_response())
            }
        }
        AuthResult::NoCredentials => {
            warn!("Admin endpoint accessed without credentials");
            let mut resp = (
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({"error": "Unauthorized", "message": "No credentials provided"})),
            ).into_response();
            resp.headers_mut().insert(
                HeaderName::from_static("www-authenticate"),
                HeaderValue::from_str("Bearer").unwrap(),
            );
            Err(resp)
        }
        AuthResult::Expired => {
            warn!("Admin endpoint accessed with expired token");
            Err((
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({"error": "Unauthorized", "message": "Token expired"})),
            ).into_response())
        }
        _ => {
            warn!("Admin endpoint authentication failed");
            Err((
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({"error": "Unauthorized", "message": "Invalid credentials"})),
            ).into_response())
        }
    }
}

#[axum::debug_handler]
async fn create_api_key(
    State(state): State<Arc<AppState>>,
    mut req: Request<Body>,
) -> Response {
    let headers = req.headers().clone();
    if let Err(resp) = require_admin_impl(&state, &headers).await {
        return resp;
    }

    let body_bytes = match axum::body::to_bytes(req.into_body(), usize::MAX).await {
        Ok(b) => b,
        Err(e) => return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "Bad Request", "message": format!("Failed to read body: {}", e)})),
        ).into_response(),
    };

    let request: CreateApiKeyRequest = match serde_json::from_slice(&body_bytes) {
        Ok(r) => r,
        Err(e) => return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "Bad Request", "message": format!("Invalid JSON: {}", e)})),
        ).into_response(),
    };

    let key = generate_api_key();
    let metadata = ApiKeyMetadata {
        service_name: request.service_name.clone(),
        permissions: request.permissions.clone(),
        created_at: now_secs(),
    };

    state.api_keys.add(key.clone(), metadata.clone()).await;
    info!("API Key created for service: {}", request.service_name);

    (
        StatusCode::CREATED,
        Json(CreateApiKeyResponse {
            key,
            service_name: metadata.service_name,
            permissions: metadata.permissions,
            created_at: metadata.created_at,
        }),
    ).into_response()
}

#[axum::debug_handler]
async fn list_api_keys(
    State(state): State<Arc<AppState>>,
    req: Request<Body>,
) -> Response {
    let headers = req.headers().clone();
    if let Err(resp) = require_admin_impl(&state, &headers).await {
        return resp;
    }

    let keys = state.api_keys.list().await;
    let entries: Vec<ListApiKeyEntry> = keys
        .into_iter()
        .map(|(key, meta)| ListApiKeyEntry {
            key_prefix: mask_key(&key),
            service_name: meta.service_name,
            permissions: meta.permissions,
            created_at: meta.created_at,
        })
        .collect();

    (StatusCode::OK, Json(ListApiKeyResponse { keys: entries })).into_response()
}

#[axum::debug_handler]
async fn delete_api_key(
    State(state): State<Arc<AppState>>,
    Query(query): Query<DeleteApiKeyQuery>,
    req: Request<Body>,
) -> Response {
    let headers = req.headers().clone();
    if let Err(resp) = require_admin_impl(&state, &headers).await {
        return resp;
    }

    let key = query.key;
    let deleted = state.api_keys.remove(&key).await;
    if deleted {
        info!("API Key deleted");
        (
            StatusCode::OK,
            Json(serde_json::json!({"status": "deleted", "message": "API key removed"})),
        ).into_response()
    } else {
        warn!("Attempted to delete non-existent API key");
        (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": "Not Found", "message": "API key not found"})),
        ).into_response()
    }
}

#[axum::debug_handler]
async fn rotate_jwt_secret(
    State(state): State<Arc<AppState>>,
    mut req: Request<Body>,
) -> Response {
    let headers = req.headers().clone();
    if let Err(resp) = require_admin_impl(&state, &headers).await {
        return resp;
    }

    let body_bytes = match axum::body::to_bytes(req.into_body(), usize::MAX).await {
        Ok(b) => b,
        Err(e) => return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "Bad Request", "message": format!("Failed to read body: {}", e)})),
        ).into_response(),
    };

    let request: RotateJwtSecretRequest = if body_bytes.is_empty() {
        RotateJwtSecretRequest { new_secret: None }
    } else {
        match serde_json::from_slice(&body_bytes) {
            Ok(r) => r,
            Err(e) => return (
                StatusCode::BAD_REQUEST,
                Json(serde_json::json!({"error": "Bad Request", "message": format!("Invalid JSON: {}", e)})),
            ).into_response(),
        }
    };

    let new_secret = request.new_secret.unwrap_or_else(generate_api_key);
    state.jwt_keys.rotate_secret(new_secret.clone()).await;

    info!("JWT secret rotated");

    (
        StatusCode::OK,
        Json(serde_json::json!({
            "status": "rotated",
            "new_secret": new_secret,
            "old_ttl_seconds": state.jwt_keys.old_ttl().as_secs(),
        })),
    ).into_response()
}

async fn health_handler() -> Response {
    (
        StatusCode::OK,
        Json(HealthResponse {
            status: "ok",
            timestamp: now_secs(),
        }),
    ).into_response()
}

#[axum::debug_handler]
async fn gateway_handler(
    State(state): State<Arc<AppState>>,
    mut req: Request<Body>,
) -> Response {
    let path = req.uri().path().to_string();
    let route = match state.routes.match_route(&path).await {
        Some(r) => r,
        None => {
            info!("Route not found: {}", path);
            return (
                StatusCode::NOT_FOUND,
                Json(serde_json::json!({"error": "Not Found", "message": "Resource not found"})),
            ).into_response();
        }
    };

    let auth_methods = route.auth_methods.clone();
    let backend_url: String = match state.routes.get_backend_url(&route.backend).await {
        Some(u) => u,
        None => {
            error!("Backend not configured: {}", route.backend);
            return (
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(serde_json::json!({"error": "Internal Server Error", "message": "Backend misconfiguration"})),
            ).into_response();
        }
    };

    if auth_methods.is_empty() {
        info!("No auth required for {}, proxying directly", path);
        let method = req.method().clone();
        let uri = req.uri().clone();
        let headers = req.headers().clone();
        let body_bytes = match axum::body::to_bytes(req.into_body(), usize::MAX).await {
            Ok(b) => Some(b.to_vec()),
            Err(e) => {
                error!("Failed to read request body: {}", e);
                return (
                    StatusCode::BAD_REQUEST,
                    Json(serde_json::json!({"error": "Bad Request", "message": format!("Failed to read body: {}", e)})),
                ).into_response();
            }
        };

        return match crate::proxy::proxy_request(method, &uri, headers, body_bytes, &backend_url, None).await {
            Ok(resp) => resp,
            Err(e) => {
                error!("Proxy failed: {}", e);
                (
                    StatusCode::BAD_GATEWAY,
                    Json(serde_json::json!({"error": "Bad Gateway", "message": e})),
                ).into_response()
            }
        };
    }

    let auth_result = authenticate(&state, req.headers(), &auth_methods).await;
    let user = match auth_result {
        AuthResult::Success(u) => {
            info!("Authentication successful for {}", path);
            u
        }
        AuthResult::NoCredentials => {
            warn!("No credentials provided for {}", path);
            let mut resp = (
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({"error": "Unauthorized", "message": "Authentication required"})),
            ).into_response();
            let www_auth = build_www_authenticate(&auth_methods);
            if let Ok(v) = HeaderValue::from_str(&www_auth) {
                resp.headers_mut().insert(
                    HeaderName::from_static("www-authenticate"),
                    v,
                );
            }
            return resp;
        }
        AuthResult::InvalidCredentials => {
            warn!("Invalid credentials for {}", path);
            let mut resp = (
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({"error": "Unauthorized", "message": "Invalid credentials"})),
            ).into_response();
            let www_auth = build_www_authenticate(&auth_methods);
            if let Ok(v) = HeaderValue::from_str(&www_auth) {
                resp.headers_mut().insert(
                    HeaderName::from_static("www-authenticate"),
                    v,
                );
            }
            return resp;
        }
        AuthResult::Expired => {
            warn!("Expired credentials for {}", path);
            let mut resp = (
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({"error": "Unauthorized", "message": "Credentials expired"})),
            ).into_response();
            let www_auth = build_www_authenticate(&auth_methods);
            if let Ok(v) = HeaderValue::from_str(&www_auth) {
                resp.headers_mut().insert(
                    HeaderName::from_static("www-authenticate"),
                    v,
                );
            }
            return resp;
        }
        AuthResult::MethodNotAllowed => {
            warn!("Auth method not allowed for {}", path);
            let mut resp = (
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({"error": "Unauthorized", "message": "Authentication method not allowed"})),
            ).into_response();
            let www_auth = build_www_authenticate(&auth_methods);
            if let Ok(v) = HeaderValue::from_str(&www_auth) {
                resp.headers_mut().insert(
                    HeaderName::from_static("www-authenticate"),
                    v,
                );
            }
            return resp;
        }
    };

    let method = req.method().clone();
    let uri = req.uri().clone();
    let headers = req.headers().clone();
    let body_bytes = match axum::body::to_bytes(req.into_body(), usize::MAX).await {
        Ok(b) => Some(b.to_vec()),
        Err(e) => {
            error!("Failed to read request body: {}", e);
            return (
                StatusCode::BAD_REQUEST,
                Json(serde_json::json!({"error": "Bad Request", "message": format!("Failed to read body: {}", e)})),
            ).into_response();
        }
    };

    info!("Proxying {} to {}", path, backend_url);
    match crate::proxy::proxy_request(method, &uri, headers, body_bytes, &backend_url, Some(&user)).await {
        Ok(resp) => resp,
        Err(e) => {
            error!("Proxy failed: {}", e);
            (
                StatusCode::BAD_GATEWAY,
                Json(serde_json::json!({"error": "Bad Gateway", "message": e})),
            ).into_response()
        }
    }
}

fn build_route_entries(config: &AppConfig) -> Vec<RouteEntry> {
    config
        .routes
        .iter()
        .map(|(pattern, rc)| RouteEntry {
            path_pattern: pattern.clone(),
            backend: rc.backend.clone(),
            auth_methods: rc.auth_methods.clone(),
        })
        .collect()
}

fn build_backend_map(config: &AppConfig) -> HashMap<String, String> {
    config
        .backends
        .iter()
        .map(|(name, bc)| (name.clone(), bc.url.clone()))
        .collect()
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "auth_gateway=info,tower_http=info".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let config = AppConfig::load().expect("Failed to load config");

    let api_keys = Arc::new(store::ApiKeyStore::from_initial(&config.initial_api_keys).await);
    let jwt_keys = Arc::new(store::JwtKeyStore::new(&config.jwt));
    let routes = Arc::new(RouteStore::new(
        build_route_entries(&config),
        build_backend_map(&config),
    ));
    let admin_user: Arc<str> = config.jwt.admin_user.clone().unwrap_or_else(|| "admin".to_string()).into();

    let state = Arc::new(AppState {
        api_keys,
        jwt_keys,
        routes,
        admin_user,
    });

    let app = Router::new()
        .route("/health", get(health_handler))
        .route("/admin/api-keys", post(create_api_key).get(list_api_keys))
        .route("/admin/api-keys", delete(delete_api_key))
        .route("/admin/jwt/rotate", post(rotate_jwt_secret))
        .fallback(gateway_handler)
        .with_state(state);

    let port = std::env::var("PORT")
        .ok()
        .and_then(|s| s.parse::<u16>().ok())
        .unwrap_or(8080);

    let addr: SocketAddr = SocketAddr::from(([0, 0, 0, 0], port));
    info!("Auth Gateway listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.expect("Failed to bind");
    axum::serve(listener, app).await.expect("Server failed");
}
