mod cache;
mod stats;

use std::sync::Arc;
use std::sync::RwLock;
use once_cell::sync::Lazy;

use axum::body::Body;
use axum::extract::Path;
use axum::http::StatusCode;
use axum::response::Response;
use axum::routing::{delete, get, put};
use axum::{Json, Router, Server};

use cache::TwoLevelCache;
use serde::Deserialize;
use stats::StatsResponse;

static CACHE: Lazy<Arc<RwLock<TwoLevelCache>>> = Lazy::new(|| {
    let l1_capacity: usize = std::env::var("L1_CAPACITY")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(1000);

    let l2_capacity: usize = std::env::var("L2_CAPACITY")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(10000);

    Arc::new(RwLock::new(TwoLevelCache::new(l1_capacity, l2_capacity)))
});

#[derive(Deserialize)]
pub struct PutRequest {
    pub value: String,
}

async fn get_handler(
    Path(key): Path<String>,
) -> Response<Body> {
    let cache = CACHE.clone();
    match tokio::task::spawn_blocking(move || {
        let cache = cache.read().unwrap();
        cache.get(&key)
    }).await {
        Ok(Some(value)) => Response::builder()
            .status(StatusCode::OK)
            .body(Body::from(value))
            .unwrap(),
        Ok(None) => Response::builder()
            .status(StatusCode::NOT_FOUND)
            .body(Body::empty())
            .unwrap(),
        Err(_) => Response::builder()
            .status(StatusCode::INTERNAL_SERVER_ERROR)
            .body(Body::empty())
            .unwrap(),
    }
}

async fn put_handler(
    Path(key): Path<String>,
    Json(body): Json<PutRequest>,
) -> Response<Body> {
    let cache = CACHE.clone();
    let _ = tokio::task::spawn_blocking(move || {
        let cache = cache.write().unwrap();
        cache.put(key, body.value);
    }).await;
    Response::builder()
        .status(StatusCode::NO_CONTENT)
        .body(Body::empty())
        .unwrap()
}

async fn delete_handler(
    Path(key): Path<String>,
) -> Response<Body> {
    let cache = CACHE.clone();
    let _ = tokio::task::spawn_blocking(move || {
        let cache = cache.write().unwrap();
        cache.delete(&key);
    }).await;
    Response::builder()
        .status(StatusCode::NO_CONTENT)
        .body(Body::empty())
        .unwrap()
}

async fn stats_handler(
) -> Json<StatsResponse> {
    let cache = CACHE.clone();
    let (stats, l1_size, l2_size) = tokio::task::spawn_blocking(move || {
        let cache = cache.read().unwrap();
        cache.get_stats()
    }).await.unwrap();
    Json(StatsResponse::new(stats, l1_size, l2_size))
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let port: u16 = std::env::var("PORT")
        .ok()
        .and_then(|s| s.parse().ok())
        .unwrap_or(3000);

    let _ = Lazy::force(&CACHE);

    let app = Router::new()
        .route("/stats", get(stats_handler))
        .route("/cache/:key", get(get_handler))
        .route("/cache/:key", put(put_handler))
        .route("/cache/:key", delete(delete_handler));

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("Listening on {}", addr);

    Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
