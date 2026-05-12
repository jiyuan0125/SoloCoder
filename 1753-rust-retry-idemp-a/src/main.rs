use std::net::SocketAddr;
use axum::{
    routing::{get, post},
    Router,
};
use tracing_subscriber;
use retry_idempotent::idempotency::new_idempotency_store;
use retry_idempotent::retry::RetryPolicy;
use retry_idempotent::downstream::new_downstream_service;
use retry_idempotent::service::new_service;
use retry_idempotent::handlers;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();
    
    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(3000);
    
    let idempotency_store = new_idempotency_store();
    let retry_policy = RetryPolicy::default_policy();
    let downstream = new_downstream_service();
    let service = new_service(idempotency_store, retry_policy, downstream);
    
    let app = Router::new()
        .route("/health", get(handlers::health))
        .route("/execute", post(handlers::execute_request))
        .route("/query/:key", get(handlers::query_request))
        .route("/statistics", get(handlers::get_statistics))
        .with_state(service);
    
    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("Server starting on {}", addr);
    
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
