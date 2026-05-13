use std::sync::Arc;
use tokio::sync::RwLock;

mod cache;
mod api;

use cache::CacheEngine;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(8906);

    let default_ttl: u64 = std::env::var("DEFAULT_TTL")
        .ok()
        .and_then(|t| t.parse().ok())
        .unwrap_or(300);

    let capacity: usize = std::env::var("CAPACITY")
        .ok()
        .and_then(|c| c.parse().ok())
        .unwrap_or(10000);

    let cache = Arc::new(RwLock::new(CacheEngine::new(capacity, default_ttl)));

    let app = api::create_router(cache);

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("Cache middleware listening on {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
