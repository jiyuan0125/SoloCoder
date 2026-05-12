use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize)]
pub struct ApiResponse<T> {
    pub success: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub data: Option<T>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

impl<T> ApiResponse<T> {
    pub fn ok(data: T) -> Self {
        ApiResponse {
            success: true,
            data: Some(data),
            error: None,
        }
    }

    pub fn empty_ok() -> Self {
        ApiResponse {
            success: true,
            data: None,
            error: None,
        }
    }

    pub fn err(message: String) -> Self {
        ApiResponse {
            success: false,
            data: None,
            error: Some(message),
        }
    }
}

#[derive(Debug, Deserialize, Serialize)]
pub struct SetCacheRequest {
    pub value: serde_json::Value,
    #[serde(default, rename = "ttlSeconds")]
    pub ttl_seconds: Option<u64>,
}

#[derive(Debug, Serialize)]
pub struct CacheEntryResponse {
    pub key: String,
    pub value: serde_json::Value,
    #[serde(rename = "remainingTtl", skip_serializing_if = "Option::is_none")]
    pub remaining_ttl: Option<u64>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct BatchGetRequest {
    pub keys: Vec<String>,
}

#[derive(Debug, Serialize)]
pub struct BatchGetResponseItem {
    pub key: String,
    pub found: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub value: Option<serde_json::Value>,
    #[serde(rename = "remainingTtl", skip_serializing_if = "Option::is_none")]
    pub remaining_ttl: Option<u64>,
}

#[derive(Debug, Serialize)]
pub struct BatchGetResponse {
    pub results: Vec<BatchGetResponseItem>,
    #[serde(rename = "totalFound")]
    pub total_found: usize,
    #[serde(rename = "totalKeys")]
    pub total_keys: usize,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct BatchSetItem {
    pub key: String,
    pub value: serde_json::Value,
    #[serde(default, rename = "ttlSeconds")]
    pub ttl_seconds: Option<u64>,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct BatchSetRequest {
    pub items: Vec<BatchSetItem>,
}

#[derive(Debug, Serialize)]
pub struct BatchResultItem {
    pub key: String,
    pub success: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct BatchSetResponse {
    pub results: Vec<BatchResultItem>,
    #[serde(rename = "totalSuccess")]
    pub total_success: usize,
    #[serde(rename = "totalFailed")]
    pub total_failed: usize,
    #[serde(rename = "totalItems")]
    pub total_items: usize,
}

#[derive(Debug, Deserialize, Serialize)]
pub struct BatchDeleteRequest {
    pub keys: Vec<String>,
}

#[derive(Debug, Serialize)]
pub struct BatchDeleteResponse {
    pub results: Vec<BatchResultItem>,
    #[serde(rename = "totalDeleted")]
    pub total_deleted: usize,
    #[serde(rename = "totalKeys")]
    pub total_keys: usize,
}

#[derive(Debug, Serialize)]
pub struct StatsResponse {
    #[serde(rename = "totalRequests")]
    pub total_requests: u64,
    #[serde(rename = "cacheHits")]
    pub cache_hits: u64,
    #[serde(rename = "cacheMisses")]
    pub cache_misses: u64,
    #[serde(rename = "hitRate")]
    pub hit_rate: f64,
    #[serde(rename = "totalSets")]
    pub total_sets: u64,
    #[serde(rename = "totalDeletes")]
    pub total_deletes: u64,
    #[serde(rename = "totalEvictions")]
    pub total_evictions: u64,
    #[serde(rename = "expiredRemoved")]
    pub expired_removed: u64,
    #[serde(rename = "currentSize")]
    pub current_size: usize,
    pub capacity: usize,
}

#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub service: String,
    pub version: String,
}
