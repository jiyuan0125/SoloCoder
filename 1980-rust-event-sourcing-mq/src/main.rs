mod aggregate;
mod config;
mod event_store;
mod models;
mod routes;
mod snapshot_manager;

use std::sync::Arc;

use aggregate::AggregateManager;
use config::AppConfig;
use event_store::EventStore;
use routes::{create_router, AppState};
use snapshot_manager::SnapshotManager;
use tracing::info;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let config = AppConfig::from_env();
    info!("Starting event sourcing service...");
    info!("Port: {}", config.port);
    info!("Snapshot directory: {}", config.snapshot_dir);
    info!(
        "Snapshot config: events_threshold={}, time_interval_secs={}",
        config.snapshot_config.events_threshold, config.snapshot_config.time_interval_secs
    );

    let event_store = EventStore::new();
    let snapshot_manager = SnapshotManager::new(&config.snapshot_dir);
    let aggregate_manager = Arc::new(AggregateManager::new(
        event_store.clone(),
        snapshot_manager.clone(),
        &config,
    ));

    let loaded_snapshots = snapshot_manager.load_all_snapshots().await;
    info!(
        "Loaded {} snapshots from disk",
        loaded_snapshots.len()
    );

    for snapshot in loaded_snapshots {
        info!(
            "Restored snapshot for aggregate {} at version {}",
            snapshot.aggregate_id, snapshot.version
        );
    }

    let app_state = AppState { aggregate_manager };

    let router = create_router(app_state);

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], config.port));
    info!("Listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::Server::from_tcp(listener.into_std().unwrap())
        .unwrap()
        .serve(router.into_make_service())
        .await
        .unwrap();
}
