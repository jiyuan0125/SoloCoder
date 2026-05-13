mod models;
mod errors;
mod validators;
mod store;
mod service;
mod routes;

use axum::routing::{get, post};
use axum::Router;
use std::env;

use crate::routes::AppState;
use crate::service::ConfigService;
use crate::store::MemoryStore;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let store = MemoryStore::new();
    let service = ConfigService::new(store);
    let app_state = AppState { service };

    let app = Router::new()
        .route("/health", get(routes::health_check))
        .route("/api/changes", post(routes::submit_change))
        .route("/api/changes/:id", get(routes::get_change_request))
        .route("/api/changes/:id/approve", post(routes::approve_change))
        .route("/api/changes/:id/reject", post(routes::reject_change))
        .route("/api/changes/:id/canary", post(routes::start_canary))
        .route("/api/changes/:id/deploy", post(routes::deploy_full))
        .route("/api/changes/:id/rollback", post(routes::rollback_change))
        .route("/api/configs/:namespace", get(routes::get_configs))
        .route("/api/audit", get(routes::get_audit_logs))
        .with_state(app_state);

    let port: u16 = env::var("PORT")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(3000);

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
