use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum AuthPolicy {
    None,
    ApiKey { key: String },
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Route {
    pub id: String,
    pub path: String,
    pub backend: String,
    pub auth_policy: AuthPolicy,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AddRouteRequest {
    pub path: String,
    pub backend: String,
    pub auth_policy: AuthPolicy,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RouteResponse {
    pub id: String,
    pub route: Route,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BackendCall {
    pub url: String,
    #[serde(default = "default_timeout_secs")]
    pub timeout_secs: u64,
    pub field_name: String,
}

fn default_timeout_secs() -> u64 {
    5
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AggregateScene {
    pub name: String,
    pub backends: Vec<BackendCall>,
}
