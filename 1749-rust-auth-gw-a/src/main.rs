mod jwt;
mod blacklist;
mod access_log;
mod middleware;

use std::env;
use axum::{
    extract::State,
    http::StatusCode,
    middleware as axum_middleware,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use serde::{Deserialize, Serialize};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};
use crate::blacklist::start_cleanup_task;
use crate::middleware::{auth_middleware, AuthState};

#[derive(Debug, Serialize, Deserialize)]
struct HealthResponse {
    status: String,
}

#[derive(Debug, Serialize, Deserialize)]
struct EchoResponse {
    message: String,
    method: String,
    user_id: Option<String>,
    headers: std::collections::HashMap<String, String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct TokenRequest {
    user_id: String,
    username: String,
    role: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
struct TokenResponse {
    token: String,
}

async fn health_handler() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "ok".to_string(),
    })
}

async fn token_handler(
    State(state): State<AuthState>,
    Json(payload): Json<TokenRequest>,
) -> impl IntoResponse {
    let role = payload.role.unwrap_or_else(|| "user".to_string());
    match state.jwt_service.generate_token(&payload.user_id, &payload.username, &role) {
        Ok(token) => (
            StatusCode::OK,
            Json(TokenResponse { token }),
        ).into_response(),
        Err(_) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(serde_json::json!({ "error": "无法生成 Token" })),
        ).into_response(),
    }
}

async fn echo_handler(
    req: axum::extract::Request,
) -> Json<EchoResponse> {
    let headers: std::collections::HashMap<String, String> = req
        .headers()
        .iter()
        .filter_map(|(k, v)| {
            v.to_str()
                .ok()
                .map(|val| (k.as_str().to_lowercase(), val.to_string()))
        })
        .collect();

    let user_id = headers.get("x-auth-user-id").cloned();

    Json(EchoResponse {
        message: "请求已通过认证网关".to_string(),
        method: req.method().to_string(),
        user_id,
        headers,
    })
}

async fn logout_handler(
    State(state): State<AuthState>,
    headers: axum::http::HeaderMap,
) -> impl IntoResponse {
    let auth_header = match headers.get("authorization") {
        Some(h) => h,
        None => {
            return (
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({ "error": "缺少 Token" })),
            ).into_response();
        }
    };

    let auth_str = match auth_header.to_str() {
        Ok(s) => s,
        Err(_) => {
            return (
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({ "error": "无效的 Authorization 头" })),
            ).into_response();
        }
    };

    let token = auth_str.strip_prefix("Bearer ").unwrap_or("");
    let token_data = match state.jwt_service.validate_token(token) {
        Ok(d) => d,
        Err(_) => {
            return (
                StatusCode::UNAUTHORIZED,
                Json(serde_json::json!({ "error": "Token 无效" })),
            ).into_response();
        }
    };

    let jti = state.jwt_service.get_jti(&token_data.claims);
    let exp = state.jwt_service.get_expiry(&token_data.claims);
    state.blacklist.add(jti, exp);

    (
        StatusCode::OK,
        Json(serde_json::json!({ "message": "Token 已注销" })),
    ).into_response()
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(tracing_subscriber::EnvFilter::try_from_default_env()
            .unwrap_or_else(|_| "auth_gateway=info,tower_http=info".into()))
        .with(tracing_subscriber::fmt::layer())
        .init();

    let port: u16 = env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse()
        .unwrap_or(8080);

    let jwt_secret = env::var("JWT_SECRET")
        .unwrap_or_else(|_| "default-jwt-secret-change-in-production-please".to_string());

    let jwt_service = crate::jwt::JwtService::new(&jwt_secret);
    let blacklist = crate::blacklist::TokenBlacklist::new();
    let access_logger = crate::access_log::AccessLogger::new();

    start_cleanup_task(blacklist.clone(), 60);

    let auth_state = AuthState {
        jwt_service: jwt_service.clone(),
        blacklist: blacklist.clone(),
        access_logger: access_logger.clone(),
    };

    let public_routes = Router::new()
        .route("/health", get(health_handler))
        .route("/token", post(token_handler))
        .route("/logout", post(logout_handler))
        .with_state(auth_state.clone());

    let protected_routes = Router::new()
        .route("/", get(echo_handler))
        .route("/echo", get(echo_handler).post(echo_handler))
        .route("/*path", get(echo_handler).post(echo_handler))
        .layer(axum_middleware::from_fn_with_state(
            auth_state.clone(),
            auth_middleware,
        ))
        .with_state(auth_state);

    let app = Router::new()
        .merge(public_routes)
        .merge(protected_routes);

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("认证网关服务启动中，监听端口: {}", port);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
