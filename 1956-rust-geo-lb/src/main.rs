mod app_state;
mod handlers;
mod health;
mod models;
mod proxy;

use crate::app_state::AppState;
use crate::handlers::{handle_get_stats, handle_get_zones, handle_post_zones, handle_proxy};
use axum::routing::{get, post};
use axum::Router;
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Duration;
use tokio::net::TcpListener;
use tracing::{info, Level};
use tracing_subscriber::FmtSubscriber;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let subscriber = FmtSubscriber::builder()
        .with_max_level(Level::INFO)
        .finish();

    tracing::subscriber::set_global_default(subscriber)
        .expect("Failed to set tracing subscriber");

    let port: u16 = std::env::var("PORT")
        .unwrap_or_else(|_| "8604".to_string())
        .parse()
        .expect("PORT must be a valid number");

    let app_state = Arc::new(AppState::new());
    let client = reqwest::Client::builder()
        .timeout(Duration::from_secs(30))
        .build()?;

    tokio::spawn(health::start_health_checker(app_state.clone()));

    let management_routes = Router::new()
        .route("/zones", post(handle_post_zones))
        .route("/zones", get(handle_get_zones))
        .route("/stats", get(handle_get_stats))
        .with_state(app_state.clone());

    let proxy_routes = Router::new()
        .fallback(handle_proxy)
        .with_state((app_state.clone(), client));

    let app = Router::new()
        .merge(management_routes)
        .merge(proxy_routes);

    let addr: SocketAddr = format!("0.0.0.0:{}", port).parse()?;
    let listener = TcpListener::bind(&addr).await?;

    info!("Geo Load Balancer listening on {}", addr);

    axum::serve(
        listener,
        app.into_make_service_with_connect_info::<SocketAddr>(),
    )
    .await?;

    Ok(())
}
