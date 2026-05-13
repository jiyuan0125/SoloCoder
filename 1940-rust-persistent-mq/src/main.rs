mod handlers;
mod queue;
mod storage;
mod types;

use std::sync::Arc;

use axum::routing::{get, post};
use axum::Router;
use tower_http::trace::TraceLayer;
use tracing_subscriber::EnvFilter;

use crate::handlers::{ack, consume, create_topic, produce};
use crate::queue::MessageQueue;
use crate::storage::Storage;

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::from_default_env())
        .init();

    let port: u16 = std::env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse()
        .expect("PORT must be a number");

    let data_dir = std::env::var("DATA_DIR")
        .unwrap_or_else(|_| "./data".to_string());

    let storage = Storage::new(&data_dir);
    let queue = Arc::new(MessageQueue::new(storage));

    let app = Router::new()
        .route("/topics", post(create_topic))
        .route("/topics/:name/produce", post(produce))
        .route("/topics/:name/consume", get(consume))
        .route("/topics/:name/ack", post(ack))
        .layer(TraceLayer::new_for_http())
        .with_state(queue);

    let listener = tokio::net::TcpListener::bind(("0.0.0.0", port))
        .await
        .expect("Failed to bind");

    tracing::info!("Server listening on 0.0.0.0:{}", port);
    axum::serve(listener, app)
        .await
        .expect("Server failed");
}
