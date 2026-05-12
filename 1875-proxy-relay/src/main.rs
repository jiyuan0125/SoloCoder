mod handlers;
mod state;

use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Duration;

use axum::routing::{delete, get, post, put};
use axum::Router;
use tokio::time::interval;
use tracing_subscriber::EnvFilter;

use crate::state::AppStateInner;
use crate::handlers::{create_backend, delete_backend, get_backends, get_stats, proxy, update_config};

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info")))
        .init();

    let state = AppStateInner::new();
    let state_clone = Arc::clone(&state);
    tokio::spawn(async move {
        let mut ticker = interval(Duration::from_secs(15));
        loop {
            ticker.tick().await;
            state_clone.check_health().await;
        }
    });

    let app = Router::new()
        .route("/backends", post(create_backend).get(get_backends))
        .route("/backends/:id", delete(delete_backend))
        .route("/config", put(update_config))
        .route("/stats", get(get_stats))
        .fallback(proxy)
        .with_state(state);

    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(3000);

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("Server listening on http://{}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
