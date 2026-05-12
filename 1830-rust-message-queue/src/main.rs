use axum::{
    extract::{Path, State},
    http::{HeaderMap, StatusCode},
    response::{IntoResponse, Sse},
    routing::{get, post},
    Json, Router,
};
use futures::{stream, StreamExt};
use serde::{Deserialize, Serialize};
use std::{
    collections::HashMap,
    sync::Arc,
};
use tokio::sync::{broadcast, RwLock};
use tokio_stream::wrappers::BroadcastStream;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

const REPLAY_COUNT: usize = 100;

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Message {
    payload: serde_json::Value,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct CreateChannelRequest {
    name: String,
    buffer_size: usize,
}

struct Channel {
    sender: broadcast::Sender<Message>,
    buffer: Arc<RwLock<VecDeque<Message>>>,
    buffer_size: usize,
}

struct AppState {
    channels: RwLock<HashMap<String, Arc<Channel>>>,
}

use std::collections::VecDeque;

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "rust_message_queue=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let app_state = Arc::new(AppState {
        channels: RwLock::new(HashMap::new()),
    });

    let app = Router::new()
        .route("/channels", post(create_channel))
        .route("/channels/:name/messages", post(publish_message))
        .route("/channels/:name/subscribe", get(subscribe))
        .with_state(app_state);

    let port = std::env::var("PORT").unwrap_or_else(|_| "3000".to_string());
    let addr = format!("0.0.0.0:{}", port);
    tracing::info!("listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn create_channel(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateChannelRequest>,
) -> impl IntoResponse {
    let mut channels = state.channels.write().await;

    if channels.contains_key(&req.name) {
        return (
            StatusCode::CONFLICT,
            Json(serde_json::json!({"error": "channel already exists"})),
        );
    }

    let (sender, _) = broadcast::channel(req.buffer_size);
    let channel = Arc::new(Channel {
        sender,
        buffer: Arc::new(RwLock::new(VecDeque::with_capacity(req.buffer_size))),
        buffer_size: req.buffer_size,
    });

    channels.insert(req.name.clone(), channel);

    tracing::debug!("channel {} created with buffer size {}", req.name, req.buffer_size);

    (
        StatusCode::CREATED,
        Json(serde_json::json!({
            "name": req.name,
            "buffer_size": req.buffer_size
        })),
    )
}

async fn publish_message(
    State(state): State<Arc<AppState>>,
    Path(name): Path<String>,
    Json(msg): Json<Message>,
) -> impl IntoResponse {
    let channels = state.channels.read().await;

    let channel = match channels.get(&name) {
        Some(c) => c.clone(),
        None => {
            return (
                StatusCode::NOT_FOUND,
                Json(serde_json::json!({"error": "channel not found"})),
            );
        }
    };

    drop(channels);

    let mut buffer = channel.buffer.write().await;
    if buffer.len() >= channel.buffer_size {
        buffer.pop_front();
    }
    buffer.push_back(msg.clone());
    drop(buffer);

    let _ = channel.sender.send(msg);

    tracing::debug!("message published to channel {}", name);

    (
        StatusCode::OK,
        Json(serde_json::json!({"status": "ok"})),
    )
}

async fn subscribe(
    State(state): State<Arc<AppState>>,
    Path(name): Path<String>,
) -> impl IntoResponse {
    let channels = state.channels.read().await;

    let channel = match channels.get(&name) {
        Some(c) => c.clone(),
        None => {
            return Err((
                StatusCode::NOT_FOUND,
                Json(serde_json::json!({"error": "channel not found"})),
            ));
        }
    };

    drop(channels);

    let buffer = channel.buffer.read().await;
    let messages_to_replay: Vec<Message> = buffer
        .iter()
        .rev()
        .take(REPLAY_COUNT)
        .rev()
        .cloned()
        .collect();
    drop(buffer);

    let receiver = channel.sender.subscribe();

    let stream = stream::iter(messages_to_replay)
        .map(|m| -> Result<_, std::convert::Infallible> {
            Ok(axum::response::sse::Event::default().json_data(m).unwrap())
        })
        .chain(
            BroadcastStream::new(receiver)
                .filter_map(|result| async move {
                    match result {
                        Ok(msg) => Some(Ok::<_, std::convert::Infallible>(
                            axum::response::sse::Event::default().json_data(msg).unwrap()
                        )),
                        Err(_) => None,
                    }
                }),
        );

    let mut headers = HeaderMap::new();
    headers.insert("Content-Type", "text/event-stream".parse().unwrap());
    headers.insert("Cache-Control", "no-cache".parse().unwrap());
    headers.insert("Connection", "keep-alive".parse().unwrap());

    Ok((headers, Sse::new(stream).keep_alive(axum::response::sse::KeepAlive::default())))
}
