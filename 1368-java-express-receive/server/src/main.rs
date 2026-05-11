use axum::{
    extract::{State, Path, Query},
    http::StatusCode,
    response::IntoResponse,
    Json,
    Router,
    routing::{get, post},
};
use clap::Parser;
use serde::Deserialize;
use std::net::SocketAddr;
use std::sync::Arc;
use tower_http::cors::CorsLayer;
use uuid::Uuid;

use express_core::{ExpressService, InboundRequest, PickupRequest, ManualVerifyRequest, ProxyRequest, ProxyFeedbackRequest};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value = "3000")]
    port: u16,

    #[arg(short, long, env = "HOST", default_value = "0.0.0.0")]
    host: String,
}

struct AppState {
    service: ExpressService,
}

#[derive(Debug, Deserialize)]
struct PhoneQuery {
    phone: String,
}

#[derive(Debug, Deserialize)]
struct CodeQuery {
    code: String,
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let service = ExpressService::new();
    let app_state = Arc::new(AppState { service });

    let app = Router::new()
        .route("/api/health", get(health_check))
        .route("/api/packages", get(list_packages).post(create_package))
        .route("/api/packages/:id", get(get_package))
        .route("/api/packages/by-phone", get(get_packages_by_phone))
        .route("/api/packages/by-code", get(get_package_by_code))
        .route("/api/pickup", post(pickup_package))
        .route("/api/manual-verify", post(manual_verify_pickup))
        .route("/api/proxy", post(register_proxy))
        .route("/api/proxy/feedback", post(feedback_proxy))
        .route("/api/proxy-records", get(list_proxy_records))
        .route("/api/notifications", get(list_notifications))
        .route("/api/check-expired-proxies", post(check_expired_proxies))
        .route("/api/check-stranded", post(check_stranded_packages))
        .layer(CorsLayer::permissive())
        .with_state(app_state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse().unwrap();
    println!("快递驿站管理系统服务端启动");
    println!("监听地址: {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}

async fn health_check() -> StatusCode {
    StatusCode::OK
}

async fn create_package(
    State(state): State<Arc<AppState>>,
    Json(req): Json<InboundRequest>,
) -> impl IntoResponse {
    match state.service.inbound_package(req).await {
        Ok(result) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": result
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn list_packages(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let packages = state.service.get_all_packages().await;
    (StatusCode::OK, Json(serde_json::json!({
        "success": true,
        "data": packages
    })))
}

async fn get_package(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_package(id).await {
        Some(pkg) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": pkg
        }))),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({
            "success": false,
            "error": "包裹不存在"
        }))),
    }
}

async fn get_packages_by_phone(
    State(state): State<Arc<AppState>>,
    Query(query): Query<PhoneQuery>,
) -> impl IntoResponse {
    let packages = state.service.get_packages_by_phone(&query.phone).await;
    (StatusCode::OK, Json(serde_json::json!({
        "success": true,
        "data": packages
    })))
}

async fn get_package_by_code(
    State(state): State<Arc<AppState>>,
    Query(query): Query<CodeQuery>,
) -> impl IntoResponse {
    match state.service.get_package_by_code(&query.code).await {
        Some(pkg) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": pkg
        }))),
        None => (StatusCode::NOT_FOUND, Json(serde_json::json!({
            "success": false,
            "error": "包裹不存在"
        }))),
    }
}

async fn pickup_package(
    State(state): State<Arc<AppState>>,
    Json(req): Json<PickupRequest>,
) -> impl IntoResponse {
    match state.service.pickup_package(req).await {
        Ok(result) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": result
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn manual_verify_pickup(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ManualVerifyRequest>,
) -> impl IntoResponse {
    match state.service.manual_verify_pickup(req).await {
        Ok(result) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": result
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn register_proxy(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ProxyRequest>,
) -> impl IntoResponse {
    match state.service.register_proxy(req).await {
        Ok((record, notification)) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": {
                "proxy_record": record,
                "notification": notification
            }
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn feedback_proxy(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ProxyFeedbackRequest>,
) -> impl IntoResponse {
    match state.service.feedback_proxy(req).await {
        Ok((record, notification)) => (StatusCode::OK, Json(serde_json::json!({
            "success": true,
            "data": {
                "proxy_record": record,
                "notification": notification
            }
        }))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({
            "success": false,
            "error": e.to_string()
        }))),
    }
}

async fn list_proxy_records(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let records = state.service.get_proxy_records().await;
    (StatusCode::OK, Json(serde_json::json!({
        "success": true,
        "data": records
    })))
}

async fn list_notifications(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let notifications = state.service.get_notifications().await;
    (StatusCode::OK, Json(serde_json::json!({
        "success": true,
        "data": notifications
    })))
}

async fn check_expired_proxies(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let packages = state.service.check_expired_proxies().await;
    (StatusCode::OK, Json(serde_json::json!({
        "success": true,
        "data": {
            "abnormal_count": packages.len(),
            "packages": packages
        }
    })))
}

async fn check_stranded_packages(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let (stranded, to_return) = state.service.check_stranded_packages().await;
    (StatusCode::OK, Json(serde_json::json!({
        "success": true,
        "data": {
            "stranded_count": stranded.len(),
            "stranded": stranded,
            "return_count": to_return.len(),
            "to_return": to_return
        }
    })))
}
