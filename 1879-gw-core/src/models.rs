use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum BackendStatus {
    Pending,
    Online,
    Degraded,
    Offline,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BackendService {
    pub id: Uuid,
    pub name: String,
    pub url: String,
    pub status: BackendStatus,
    pub consecutive_healthy_checks: u32,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum MatchType {
    Exact,
    Prefix,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum AuthPolicy {
    Public,
    ApiKey,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Route {
    pub id: Uuid,
    pub path: String,
    pub match_type: MatchType,
    pub backend_name: String,
    pub auth_policy: AuthPolicy,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateRouteRequest {
    pub path: String,
    pub match_type: MatchType,
    pub backend_name: String,
    pub auth_policy: AuthPolicy,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiKey {
    pub key: String,
    pub created_at: u64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateApiKeyRequest {
    pub key: Option<String>,
}

#[derive(Debug, Clone, Default, Serialize)]
pub struct BackendStats {
    pub total_requests: u64,
    pub error_requests: u64,
    pub recent_requests: Vec<bool>,
}

#[derive(Debug, Clone, Serialize)]
pub struct ServiceStatusResponse {
    pub name: String,
    pub url: String,
    pub status: BackendStatus,
}

#[derive(Debug, Clone, Serialize)]
pub struct StatsResponse {
    pub backend_name: String,
    pub total_requests: u64,
    pub error_requests: u64,
    pub error_rate: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RegisterBackendRequest {
    pub name: String,
    pub url: String,
}
