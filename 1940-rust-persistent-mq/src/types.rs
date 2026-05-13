use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Message {
    pub offset: u64,
    pub key: Option<String>,
    pub payload: String,
    pub timestamp: u64,
    pub partition: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PendingMessage {
    pub message: Message,
    pub expire_at: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TopicConfig {
    pub name: String,
    pub partitions: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProduceRequest {
    pub key: Option<String>,
    pub payload: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ProduceResponse {
    pub offset: u64,
    pub partition: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AckRequest {
    pub offset: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GroupMetadata {
    pub last_ack_offset: u64,
    pub pending_acks: HashMap<u64, u64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PartitionMetadata {
    pub topic: String,
    pub partition: u32,
    pub next_offset: u64,
}
