mod pool;
mod api;

use std::env;
use std::net::SocketAddr;
use std::time::Duration;

use axum::{routing::get, routing::post, Router};
use tokio::signal;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use crate::pool::{PoolConfig, ConnectionPool};
use crate::api::{AppState, stats_handler, config_handler, health_handler};

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "conn_pool=info,tower_http=debug,axum::rejection=trace".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let port: u16 = env::var("PORT")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(8106);

    let config = PoolConfig {
        max_connections: 10,
        idle_timeout: Duration::from_secs(60),
        health_check_interval: Duration::from_secs(30),
        graceful_shutdown_timeout: Duration::from_secs(10),
        max_health_failures: 2,
    };

    let pool = ConnectionPool::new(config.clone());
    let pool_clone = pool.clone();

    let pool_for_cleanup = pool.clone();
    tokio::spawn(async move {
        pool_for_cleanup.start_background_tasks().await;
    });

    let state = AppState {
        pool: pool_clone.clone(),
    };

    let app = Router::new()
        .route("/health", get(health_handler))
        .route("/stats", get(stats_handler))
        .route("/config", post(config_handler))
        .with_state(state);

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();

    let server = axum::serve(listener, app);

    let shutdown_signal = async {
        let ctrl_c = async {
            signal::ctrl_c()
                .await
                .expect("failed to install Ctrl+C handler");
        };

        #[cfg(unix)]
        let terminate = async {
            signal::unix::signal(signal::unix::SignalKind::terminate())
                .expect("failed to install signal handler")
                .recv()
                .await;
        };

        #[cfg(not(unix))]
        let terminate = std::future::pending::<()>();

        tokio::select! {
            _ = ctrl_c => {},
            _ = terminate => {},
        }

        tracing::info!("shutdown signal received, initiating graceful shutdown");
        pool_clone.shutdown().await;
    };

    tokio::select! {
        _ = server => {},
        _ = shutdown_signal => {},
    }

    tracing::info!("server shutdown complete");
}
