use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum ConfigType {
    String,
    Number,
    Boolean,
    Array,
    Object,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Schema {
    #[serde(rename = "type")]
    pub config_type: ConfigType,
    pub required: Option<bool>,
    #[serde(alias = "min")]
    pub minimum: Option<f64>,
    #[serde(alias = "max")]
    pub maximum: Option<f64>,
    #[serde(alias = "minLength")]
    pub min_length: Option<usize>,
    #[serde(alias = "maxLength")]
    pub max_length: Option<usize>,
    pub pattern: Option<String>,
    pub items: Option<Box<Schema>>,
    pub properties: Option<std::collections::HashMap<String, Schema>>,
    #[serde(alias = "enum")]
    pub enum_values: Option<Vec<serde_json::Value>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConfigChangeRequest {
    pub key: String,
    pub value: serde_json::Value,
    pub schema: Schema,
    pub description: Option<String>,
    pub author: Option<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ValidationError {
    pub field: String,
    pub expected: String,
    pub actual: String,
    pub message: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ValidationResult {
    pub valid: bool,
    pub errors: Vec<ValidationError>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ConfigVersion {
    pub version: Uuid,
    pub key: String,
    pub value: serde_json::Value,
    pub schema: Schema,
    pub description: Option<String>,
    pub author: Option<String>,
    pub created_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum DeployStrategy {
    Canary,
    Full,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
#[serde(rename_all = "lowercase")]
pub enum DeployStatus {
    Pending,
    InProgress,
    Success,
    Failed,
    RolledBack,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeployRecord {
    pub id: Uuid,
    pub key: String,
    pub version: Uuid,
    pub strategy: DeployStrategy,
    pub percentage: Option<u32>,
    pub status: DeployStatus,
    pub created_at: DateTime<Utc>,
    pub completed_at: Option<DateTime<Utc>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ImpactResponse {
    pub key: String,
    pub services: Vec<ServiceInfo>,
    pub total_instances: u32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServiceInfo {
    pub name: String,
    pub instance_count: u32,
    pub environment: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ServiceMapping {
    pub key: String,
    pub services: Vec<ServiceInfo>,
}
