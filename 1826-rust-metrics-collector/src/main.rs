mod api;
mod store;
mod summary;
mod types;

use api::{ingest_metrics, query_metrics};
use axum::Router;
use store::MetricsStore;
use std::env;
use std::net::SocketAddr;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();

    let store = MetricsStore::new();

    let app = Router::new()
        .route(
            "/metrics",
            axum::routing::get(query_metrics).post(ingest_metrics),
        )
        .with_state(store);

    let port_str = env::var("PORT").unwrap_or_else(|_| "3000".to_string());
    let port: u16 = port_str.parse().expect("PORT must be a valid port number");

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
