mod discovery;
mod heartbeat;
mod models;
mod notify;
mod routes;
mod store;

use crate::discovery::discover_service;
use crate::notify::{list_subscriptions, subscribe, unsubscribe};
use crate::routes::{deregister_instance, handle_heartbeat, patch_metadata, register_instance};
use crate::store::AppState;
use axum::{routing, Router};
use std::net::SocketAddr;
use std::time::Duration;
use tower_http::trace::TraceLayer;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let state = AppState::new();

    let app = Router::new()
        .route(
            "/services/:name/instances",
            routing::post(register_instance),
        )
        .route(
            "/services/:name/instances/:iid",
            routing::delete(deregister_instance),
        )
        .route(
            "/services/:name/instances/:iid/heartbeat",
            routing::put(handle_heartbeat),
        )
        .route(
            "/services/:name/instances/:iid/metadata",
            routing::patch(patch_metadata),
        )
        .route(
            "/services/:name/instances",
            routing::get(discover_service),
        )
        .route("/subscriptions", routing::post(subscribe).get(list_subscriptions))
        .route("/subscriptions/:id", routing::delete(unsubscribe))
        .with_state(state.clone())
        .layer(TraceLayer::new_for_http());

    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(3000);

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("Server listening on {}", addr);

    let state_for_heartbeat = state.clone();
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(Duration::from_secs(5));
        loop {
            interval.tick().await;
            let result = heartbeat::check_heartbeat_timeouts(state_for_heartbeat.clone()).await;

            if !result.became_unhealthy.is_empty() {
                tracing::info!(
                    "{} instances became unhealthy",
                    result.became_unhealthy.len()
                );
                notify::notify_batch(
                    state_for_heartbeat.clone(),
                    "unhealthy",
                    result.became_unhealthy,
                )
                .await;
            }

            if !result.cleaned_up.is_empty() {
                tracing::info!("{} instances cleaned up", result.cleaned_up.len());
                notify::notify_batch(
                    state_for_heartbeat.clone(),
                    "deregistered",
                    result.cleaned_up,
                )
                .await;
            }

            if !result.became_healthy.is_empty() {
                tracing::info!("{} instances became healthy", result.became_healthy.len());
                notify::notify_batch(
                    state_for_heartbeat.clone(),
                    "healthy",
                    result.became_healthy,
                )
                .await;
            }
        }
    });

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
