mod cache;
mod handlers;
mod models;

use std::env;
use std::sync::Arc;
use std::time::Duration;

use axum::routing::{delete, get, post, put};
use axum::Router;
use tokio::signal;

use cache::{CacheManager, DEFAULT_CAPACITY, DEFAULT_TTL_SECS};

const CLEANUP_INTERVAL_SECS: u64 = 60;

fn build_app(cache: CacheManager) -> Router {
    Router::new()
        .route("/health", get(handlers::health))
        .route("/cache/:key", get(handlers::get_cache))
        .route("/cache/:key", put(handlers::set_cache))
        .route("/cache/:key", delete(handlers::delete_cache))
        .route("/batch/get", post(handlers::batch_get))
        .route("/batch/set", post(handlers::batch_set))
        .route("/batch/delete", post(handlers::batch_delete))
        .route("/stats", get(handlers::get_stats))
        .route("/clear", post(handlers::clear_cache))
        .with_state(Arc::new(cache))
}

fn parse_env_config() -> (u16, usize, u64) {
    let port = env::var("PORT")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(8080);

    let capacity = env::var("CACHE_CAPACITY")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(DEFAULT_CAPACITY);

    let default_ttl = env::var("DEFAULT_TTL_SECS")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(DEFAULT_TTL_SECS);

    (port, capacity, default_ttl)
}

async fn cleanup_task(cache: CacheManager) {
    let mut interval = tokio::time::interval(Duration::from_secs(CLEANUP_INTERVAL_SECS));
    loop {
        interval.tick().await;
        let removed = cache.cleanup_expired();
        if removed > 0 {
            eprintln!("Cleaned up {} expired entries", removed);
        }
    }
}

async fn shutdown_signal() {
    let ctrl_c = async {
        signal::ctrl_c()
            .await
            .expect("failed to install Ctrl+C handler");
    };

    let terminate = async {
        signal::unix::signal(signal::unix::SignalKind::terminate())
            .expect("failed to install signal handler")
            .recv()
            .await;
    };

    tokio::select! {
        _ = ctrl_c => {},
        _ = terminate => {},
    }

    println!("Shutting down...");
}

#[tokio::main]
async fn main() {
    let (port, capacity, default_ttl) = parse_env_config();

    println!("Starting Rust Cache Service");
    println!("Port: {}", port);
    println!("Capacity: {}", capacity);
    println!("Default TTL: {}s", default_ttl);

    let cache = CacheManager::with_capacity(capacity, default_ttl);
    let cleanup_cache = cache.clone();

    tokio::spawn(cleanup_task(cleanup_cache));

    let app = build_app(cache);
    let addr = format!("0.0.0.0:{}", port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();

    println!("Server listening on: {}", addr);

    axum::serve(listener, app)
        .with_graceful_shutdown(shutdown_signal())
        .await
        .unwrap();
}
