mod handlers;
mod models;
mod storage;

use axum::{routing, Router};
use std::net::SocketAddr;
use std::path::PathBuf;
use std::sync::Arc;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

use crate::handlers::AppState;
use crate::storage::ConfigStore;

fn get_or_generate_encryption_key() -> [u8; 32] {
    let key_file = PathBuf::from("data/encryption.key");
    
    if key_file.exists() {
        if let Ok(contents) = std::fs::read(&key_file) {
            if contents.len() == 32 {
                let mut key = [0u8; 32];
                key.copy_from_slice(&contents);
                return key;
            }
        }
    }

    let mut key = [0u8; 32];
    rand::RngCore::fill_bytes(&mut rand::thread_rng(), &mut key);
    
    if let Some(parent) = key_file.parent() {
        std::fs::create_dir_all(parent).ok();
    }
    std::fs::write(&key_file, &key).ok();
    
    key
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "atomic_config=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let encryption_key = get_or_generate_encryption_key();
    let data_dir = PathBuf::from("data");
    let store = ConfigStore::new(data_dir, encryption_key);
    let state: AppState = Arc::new(store);

    let app = Router::new()
        .route("/configs", routing::get(handlers::list_configs))
        .route("/configs", routing::post(handlers::batch_update_config))
        .route("/configs/:key", routing::get(handlers::get_config))
        .route("/configs/:key", routing::put(handlers::update_config))
        .route("/configs/:key", routing::delete(handlers::delete_config))
        .route("/audit/logs", routing::get(handlers::get_audit_logs))
        .with_state(state);

    let port = std::env::var("PORT")
        .ok()
        .and_then(|s| s.parse::<u16>().ok())
        .unwrap_or(3000);

    let addr = SocketAddr::from(([127, 0, 0, 1], port));
    tracing::debug!("监听地址: {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
