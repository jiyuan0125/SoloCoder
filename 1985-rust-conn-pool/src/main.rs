use std::sync::Arc;
use std::env;

mod types;
mod rate_limiter;
mod connection;
mod pool_manager;
mod api;

use pool_manager::PoolManager;
use api::create_router;
use tracing_subscriber::{EnvFilter, fmt};
use tracing::{info, error};

#[tokio::main]
async fn main() {
    let filter = EnvFilter::try_from_default_env()
        .unwrap_or_else(|_| EnvFilter::new("info"));
    
    fmt()
        .with_env_filter(filter)
        .init();
    
    let pool_manager = Arc::new(PoolManager::new());
    pool_manager.start_background_tasks().await;
    
    let router = create_router(pool_manager);
    
    let port = env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse::<u16>()
        .unwrap_or(3000);
    
    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    
    info!("Starting connection pool server on {}", addr);
    
    let listener = match tokio::net::TcpListener::bind(addr).await {
        Ok(l) => l,
        Err(e) => {
            error!("Failed to bind to {}: {}", addr, e);
            std::process::exit(1);
        }
    };
    
    info!("Server listening on {}", addr);
    
    if let Err(e) = axum::serve(listener, router).await {
        error!("Server error: {}", e);
        std::process::exit(1);
    }
}
