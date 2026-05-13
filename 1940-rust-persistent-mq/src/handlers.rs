use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::response::IntoResponse;
use axum::Json;
use serde::Deserialize;
use std::sync::Arc;

use crate::queue::MessageQueue;
use crate::types::{AckRequest, ProduceRequest, TopicConfig};

#[derive(Debug, Deserialize)]
pub struct ConsumeQuery {
    pub group: String,
}

pub async fn create_topic(
    State(queue): State<Arc<MessageQueue>>,
    Json(config): Json<TopicConfig>,
) -> impl IntoResponse {
    queue.create_topic(&config);
    (StatusCode::CREATED, Json(serde_json::json!({
        "status": "created",
        "name": config.name,
        "partitions": config.partitions
    })))
}

pub async fn produce(
    State(queue): State<Arc<MessageQueue>>,
    Path(topic_name): Path<String>,
    Json(req): Json<ProduceRequest>,
) -> impl IntoResponse {
    if queue.get_topic_config(&topic_name).is_none() {
        return (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": "Topic not found"})),
        );
    }

    let (offset, partition) = queue.produce(&topic_name, req.key.as_deref(), &req.payload);

    (
        StatusCode::OK,
        Json(serde_json::json!({"offset": offset, "partition": partition})),
    )
}

pub async fn consume(
    State(queue): State<Arc<MessageQueue>>,
    Path(topic_name): Path<String>,
    Query(query): Query<ConsumeQuery>,
) -> impl IntoResponse {
    if queue.get_topic_config(&topic_name).is_none() {
        return (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": "Topic not found"})),
        );
    }

    match queue.consume(&topic_name, &query.group) {
        Some(msg) => (StatusCode::OK, Json(serde_json::to_value(msg).unwrap())),
        None => (
            StatusCode::NO_CONTENT,
            Json(serde_json::json!({"status": "no messages"})),
        ),
    }
}

pub async fn ack(
    State(queue): State<Arc<MessageQueue>>,
    Path(topic_name): Path<String>,
    Query(query): Query<ConsumeQuery>,
    Json(req): Json<AckRequest>,
) -> impl IntoResponse {
    if queue.get_topic_config(&topic_name).is_none() {
        return (
            StatusCode::NOT_FOUND,
            Json(serde_json::json!({"error": "Topic not found"})),
        );
    }

    let partition = queue.find_partition_for_offset(&topic_name, req.offset);

    if let Some(p) = partition {
        if queue.ack(&topic_name, &query.group, req.offset, p) {
            return (
                StatusCode::OK,
                Json(serde_json::json!({"status": "acknowledged"})),
            );
        }
    }

    (
        StatusCode::BAD_REQUEST,
        Json(serde_json::json!({"error": "Invalid offset or not pending"})),
    )
}
