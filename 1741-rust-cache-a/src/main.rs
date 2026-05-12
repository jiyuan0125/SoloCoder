mod cache;
mod api;

use std::env;
use std::sync::Arc;
use tokio::sync::RwLock;
use tracing_subscriber::EnvFilter;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();

    let max_size: usize = env::var("CACHE_MAX_SIZE")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(10000);

    let cache = Arc::new(RwLock::new(cache::Cache::new(max_size)));

    let app = api::create_router(cache);

    let port = env::var("PORT").unwrap_or_else(|_| "8422".to_string());
    let addr = format!("0.0.0.0:{}", port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();

    tracing::info!("listening on {}", addr);
    axum::serve(listener, app).await.unwrap();
}
