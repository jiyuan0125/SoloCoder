use std::sync::Arc;

use axum::{routing::{get, post}, Router};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

mod models;
mod store;
mod scheduler;
mod handlers;

use handlers::AppState;

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "retry_scheduler=info,tower_http=debug,axum=trace".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let store = store::TaskStore::new();
    let scheduler = Arc::new(scheduler::Scheduler::new(store.clone(), Some(100)));

    scheduler.clone().run().await;

    let app_state = AppState {
        store: store.clone(),
    };

    let app = Router::new()
        .route("/tasks", post(handlers::submit_task))
        .route("/tasks", get(handlers::list_tasks))
        .route("/tasks/:task_id", get(handlers::get_task))
        .route("/stats", get(handlers::get_stats))
        .with_state(app_state);

    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse::<u16>()
        .unwrap_or(3000);

    let listener = tokio::net::TcpListener::bind(("0.0.0.0", port))
        .await
        .unwrap();

    tracing::info!("Server running on port {}", port);

    axum::serve(listener, app).await.unwrap();
}
