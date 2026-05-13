mod algorithms;
mod handlers;
mod middleware;
mod store;
mod types;

use crate::handlers::AppState;
use crate::middleware::make_rate_limit_layer;
use crate::store::RateLimiterEngine;
use axum::{routing, Router};
use std::env;
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::net::TcpListener;
use tower::ServiceBuilder;

#[tokio::main]
async fn main() {
    env_logger::init();

    let port: u16 = env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(8080);

    let max_counters: usize = env::var("MAX_COUNTERS")
        .ok()
        .and_then(|m| m.parse().ok())
        .unwrap_or(10000);

    log::info!("Starting rate limiter service on port {}", port);
    log::info!("Maximum counters: {}", max_counters);

    let engine = Arc::new(RateLimiterEngine::new(max_counters));
    let app_state = AppState {
        engine: engine.clone(),
    };

    let rate_limit_layer = make_rate_limit_layer(engine.clone());

    let api_routes = Router::new()
        .route("/rules", routing::post(handlers::create_rule))
        .route("/rules", routing::get(handlers::get_all_rules))
        .route("/rules/:id", routing::get(handlers::get_rule))
        .route("/rules/:id", routing::put(handlers::update_rule))
        .route("/rules/:id", routing::delete(handlers::delete_rule))
        .route("/stats", routing::get(handlers::get_path_statistics))
        .route("/throttled", routing::get(handlers::get_all_throttled_ips))
        .route("/throttled/:ip", routing::get(handlers::get_ip_throttled_state))
        .with_state(app_state.clone());

    let health_route = Router::new()
        .route("/health", routing::get(handlers::health_check));

    let app = Router::new()
        .route("/", routing::get(handlers::default_handler))
        .nest("/api", api_routes)
        .merge(health_route)
        .layer(
            ServiceBuilder::new()
                .layer(rate_limit_layer),
        );

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    let listener = TcpListener::bind(addr).await.unwrap();

    log::info!("Server listening on {}", addr);

    axum::serve(listener, app).await.unwrap();
}
