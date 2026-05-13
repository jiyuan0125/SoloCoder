use std::collections::HashMap;
use std::sync::Arc;

use axum::body::Bytes;
use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::response::{sse::Event, IntoResponse, Response, Sse};
use axum::routing::{get, post};
use axum::{Json, Router};
use chrono::{DateTime, Utc};
use futures::stream;
use futures::StreamExt;
use serde::{Deserialize, Serialize};
use tokio::sync::{broadcast, RwLock};
use tokio_stream::wrappers::BroadcastStream;
use uuid::Uuid;

const MAX_MESSAGE_SIZE: usize = 1024 * 1024;
const DEFAULT_BUFFER_SIZE: usize = 256;
const HISTORY_SIZE: usize = 100;

#[derive(Debug, Clone)]
struct Message {
    id: Uuid,
    timestamp: DateTime<Utc>,
    content: Vec<u8>,
}

#[derive(Debug, Clone)]
struct Channel {
    name: String,
    buffer_size: usize,
    sender: broadcast::Sender<Message>,
    history: Arc<RwLock<Vec<Message>>>,
    total_messages: Arc<RwLock<u64>>,
}

impl Channel {
    fn new(name: String, buffer_size: usize) -> Self {
        let (sender, _) = broadcast::channel(buffer_size);
        Channel {
            name,
            buffer_size,
            sender,
            history: Arc::new(RwLock::new(Vec::with_capacity(HISTORY_SIZE))),
            total_messages: Arc::new(RwLock::new(0)),
        }
    }

    fn subscriber_count(&self) -> usize {
        self.sender.receiver_count()
    }

    async fn publish(&self, content: Vec<u8>) -> (Uuid, DateTime<Utc>) {
        let message = Message {
            id: Uuid::new_v4(),
            timestamp: Utc::now(),
            content,
        };

        {
            let mut total = self.total_messages.write().await;
            *total += 1;
        }

        {
            let mut history = self.history.write().await;
            if history.len() >= HISTORY_SIZE {
                history.remove(0);
            }
            history.push(message.clone());
        }

        if self.sender.receiver_count() > 0 {
            if self.sender.send(message.clone()).is_err() {
                tracing::warn!(
                    channel = self.name,
                    "All receivers dropped, message not delivered"
                );
            }
        }

        (message.id, message.timestamp)
    }

    async fn get_history(&self) -> Vec<Message> {
        self.history.read().await.clone()
    }

    fn subscribe(&self) -> broadcast::Receiver<Message> {
        self.sender.subscribe()
    }
}

#[derive(Debug, Clone)]
struct AppState {
    channels: Arc<RwLock<HashMap<String, Channel>>>,
}

#[derive(Debug, Deserialize)]
struct CreateChannelRequest {
    name: String,
    #[serde(default = "default_buffer_size")]
    buffer_size: usize,
}

fn default_buffer_size() -> usize {
    DEFAULT_BUFFER_SIZE
}

#[derive(Debug, Serialize)]
struct ChannelMetadata {
    name: String,
    buffer_size: usize,
    subscriber_count: usize,
    total_messages: u64,
}

#[derive(Debug, Serialize)]
struct PublishResponse {
    message_id: Uuid,
    timestamp: DateTime<Utc>,
}

#[derive(Debug, Deserialize)]
struct ListQuery {
    #[serde(default)]
    q: Option<String>,
    #[serde(default)]
    search: Option<String>,
}

async fn create_channel(
    State(state): State<AppState>,
    Json(req): Json<CreateChannelRequest>,
) -> impl IntoResponse {
    let mut channels = state.channels.write().await;

    if channels.contains_key(&req.name) {
        return (StatusCode::CONFLICT, ());
    }

    let buffer_size = if req.buffer_size == 0 {
        DEFAULT_BUFFER_SIZE
    } else {
        req.buffer_size
    };

    let channel = Channel::new(req.name.clone(), buffer_size);
    channels.insert(req.name, channel);

    (StatusCode::CREATED, ())
}

async fn publish_message(
    State(state): State<AppState>,
    Path(name): Path<String>,
    body: Bytes,
) -> impl IntoResponse {
    let content_length = body.len();

    if content_length == 0 {
        return (
            StatusCode::BAD_REQUEST,
            Json(None::<PublishResponse>),
        );
    }

    if content_length > MAX_MESSAGE_SIZE {
        return (
            StatusCode::PAYLOAD_TOO_LARGE,
            Json(None::<PublishResponse>),
        );
    }

    let channels = state.channels.read().await;

    let channel = match channels.get(&name) {
        Some(c) => c.clone(),
        None => {
            return (
                StatusCode::NOT_FOUND,
                Json(None::<PublishResponse>),
            );
        }
    };

    drop(channels);

    let (message_id, timestamp) = channel.publish(body.to_vec()).await;

    (
        StatusCode::OK,
        Json(Some(PublishResponse { message_id, timestamp })),
    )
}

fn message_to_event(msg: Message) -> Result<Event, axum::Error> {
    let content = String::from_utf8_lossy(&msg.content);

    let data = serde_json::json!({
        "id": msg.id,
        "timestamp": msg.timestamp,
        "content": content,
    })
    .to_string();

    Ok(Event::default()
        .event("message")
        .id(msg.id.to_string())
        .data(data))
}

async fn subscribe_channel(
    State(state): State<AppState>,
    Path(name): Path<String>,
) -> Response {
    let channels = state.channels.read().await;

    let channel = match channels.get(&name) {
        Some(c) => c.clone(),
        None => {
            return (StatusCode::NOT_FOUND, "Channel not found").into_response();
        }
    };

    drop(channels);

    let channel_name = name;
    let history = channel.get_history().await;
    let receiver = channel.subscribe();

    let history_stream =
        stream::iter(history.into_iter().map(|m| Ok(m)));

    let broadcast_stream = BroadcastStream::new(receiver).map(move |result| match result {
        Ok(msg) => Ok(msg),
        Err(tokio_stream::wrappers::errors::BroadcastStreamRecvError::Lagged(lost)) => {
            tracing::warn!(
                channel = channel_name,
                lost_messages = lost,
                "Subscriber lagged behind, dropping oldest messages"
            );
            Err(axum::Error::new(format!(
                "Lost {} messages due to slow consumer",
                lost
            )))
        }
    });

    let combined_stream = history_stream.chain(broadcast_stream).map(|result| match result {
        Ok(msg) => message_to_event(msg),
        Err(e) => Err(e),
    });

    let sse: Sse<_> = Sse::new(combined_stream);
    sse.into_response()
}

async fn list_channels(
    State(state): State<AppState>,
    Query(query): Query<ListQuery>,
) -> impl IntoResponse {
    let channels = state.channels.read().await;

    let search_term = query.search.or(query.q);

    let mut result = Vec::new();

    for (name, channel) in channels.iter() {
        if let Some(q) = &search_term {
            if !name.to_lowercase().contains(&q.to_lowercase()) {
                continue;
            }
        }

        let total_messages = *channel.total_messages.read().await;

        result.push(ChannelMetadata {
            name: name.clone(),
            buffer_size: channel.buffer_size,
            subscriber_count: channel.subscriber_count(),
            total_messages,
        });
    }

    Json(result)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "info".into()),
        )
        .init();

    let state = AppState {
        channels: Arc::new(RwLock::new(HashMap::new())),
    };

    let app = Router::new()
        .route("/channels", post(create_channel).get(list_channels))
        .route("/channels/:name/messages", post(publish_message))
        .route("/channels/:name/subscribe", get(subscribe_channel))
        .with_state(state);

    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse::<u16>()
        .expect("PORT must be a valid port number");

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));

    tracing::info!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
