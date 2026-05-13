use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::Json,
    routing::{get, post, put},
    Router,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;

use crate::topic::TopicManager;

#[derive(Clone)]
pub struct AppState {
    pub topic_manager: Arc<TopicManager>,
}

#[derive(Debug, Deserialize)]
pub struct ProduceRequest {
    pub messages: Vec<ProduceMessage>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct ProduceMessage {
    pub payload: String,
}

#[derive(Debug, Serialize)]
pub struct ProduceResponse {
    pub offsets: Vec<u64>,
}

#[derive(Debug, Deserialize)]
pub struct ConsumeQuery {
    pub group: String,
    #[serde(default = "default_limit")]
    pub limit: usize,
}

fn default_limit() -> usize {
    100
}

#[derive(Debug, Serialize)]
pub struct ConsumeResponse {
    pub messages: Vec<ConsumeMessage>,
}

#[derive(Debug, Serialize)]
pub struct ConsumeMessage {
    pub offset: u64,
    pub payload: String,
    pub timestamp: u64,
}

#[derive(Debug, Deserialize)]
pub struct AckRequest {
    pub group: String,
    pub offset: u64,
}

#[derive(Debug, Deserialize)]
pub struct SetOffsetRequest {
    pub offset: u64,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

pub async fn produce_handler(
    State(state): State<AppState>,
    Path(topic_name): Path<String>,
    Json(request): Json<ProduceRequest>,
) -> Result<Json<ProduceResponse>, (StatusCode, Json<ErrorResponse>)> {
    let mut offsets = Vec::with_capacity(request.messages.len());
    
    for msg in request.messages {
        let payload = msg.payload.as_bytes();
        match state.topic_manager.produce(&topic_name, payload) {
            Ok(offset) => offsets.push(offset),
            Err(e) => {
                return Err((
                    StatusCode::INTERNAL_SERVER_ERROR,
                    Json(ErrorResponse { error: e.to_string() }),
                ));
            }
        }
    }
    
    Ok(Json(ProduceResponse { offsets }))
}

pub async fn consume_handler(
    State(state): State<AppState>,
    Path(topic_name): Path<String>,
    Query(query): Query<ConsumeQuery>,
) -> Result<Json<ConsumeResponse>, (StatusCode, Json<ErrorResponse>)> {
    let limit = if query.limit == 0 || query.limit > 1000 {
        100
    } else {
        query.limit
    };
    
    let messages = match state.topic_manager.consume(&topic_name, &query.group, limit) {
        Ok(msgs) => msgs,
        Err(e) => {
            return Err((
                StatusCode::INTERNAL_SERVER_ERROR,
                Json(ErrorResponse { error: e.to_string() }),
            ));
        }
    };
    
    let consume_messages: Vec<ConsumeMessage> = messages
        .into_iter()
        .map(|msg| ConsumeMessage {
            offset: msg.offset,
            payload: String::from_utf8_lossy(&msg.payload).into_owned(),
            timestamp: msg.timestamp,
        })
        .collect();
    
    Ok(Json(ConsumeResponse { messages: consume_messages }))
}

pub async fn ack_handler(
    State(state): State<AppState>,
    Path(topic_name): Path<String>,
    Json(request): Json<AckRequest>,
) -> Result<(), (StatusCode, Json<ErrorResponse>)> {
    match state.topic_manager.ack(&topic_name, &request.group, request.offset) {
        Ok(_) => Ok(()),
        Err(e) => Err((
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse { error: e.to_string() }),
        )),
    }
}

pub async fn set_consumer_offset_handler(
    State(state): State<AppState>,
    Path(group_name): Path<String>,
    Query(query): Query<SetOffsetQuery>,
    Json(request): Json<SetOffsetRequest>,
) -> Result<(), (StatusCode, Json<ErrorResponse>)> {
    let topic = match query.topic {
        Some(t) => t,
        None => {
            return Err((
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse { error: "Missing 'topic' query parameter".to_string() }),
            ));
        }
    };
    
    match state.topic_manager.set_consumer_offset(&topic, &group_name, request.offset) {
        Ok(_) => Ok(()),
        Err(e) => Err((
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse { error: e.to_string() }),
        )),
    }
}

#[derive(Debug, Deserialize)]
pub struct SetOffsetQuery {
    pub topic: Option<String>,
}

pub fn create_router(state: AppState) -> Router {
    Router::new()
        .route("/topics/:name/produce", post(produce_handler))
        .route("/topics/:name/consume", get(consume_handler))
        .route("/topics/:name/ack", post(ack_handler))
        .route("/consumer-groups/:name/offset", put(set_consumer_offset_handler))
        .with_state(state)
}
