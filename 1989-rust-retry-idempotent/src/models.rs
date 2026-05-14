use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Utc};
use std::fmt;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum TaskStatus {
    Pending,
    Running,
    Retrying,
    Succeeded,
    Failed,
}

impl fmt::Display for TaskStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let s = match self {
            TaskStatus::Pending => "pending",
            TaskStatus::Running => "running",
            TaskStatus::Retrying => "retrying",
            TaskStatus::Succeeded => "succeeded",
            TaskStatus::Failed => "failed",
        };
        write!(f, "{}", s)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum RetryStrategy {
    Fixed,
    ExponentialBackoff,
}

impl Default for RetryStrategy {
    fn default() -> Self {
        RetryStrategy::ExponentialBackoff
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskRequest {
    pub url: String,
    #[serde(default)]
    pub body: Option<serde_json::Value>,
    #[serde(default = "default_max_retries")]
    pub max_retries: u32,
    #[serde(default)]
    pub retry_strategy: RetryStrategy,
    #[serde(default = "default_base_interval")]
    pub base_interval_ms: u64,
    #[serde(default)]
    pub jitter_max_ms: u64,
    #[serde(default = "default_http_timeout")]
    pub http_timeout_seconds: u64,
    pub idempotency_key: Option<String>,
}

fn default_max_retries() -> u32 { 3 }
fn default_base_interval() -> u64 { 1000 }
fn default_http_timeout() -> u64 { 30 }

#[derive(Debug, Clone, Serialize)]
pub struct TaskResponse {
    pub task_id: Uuid,
    pub idempotency_key: String,
    pub status: TaskStatus,
}

#[derive(Debug, Clone, Serialize)]
pub struct TaskDetail {
    pub task_id: Uuid,
    pub idempotency_key: String,
    pub url: String,
    pub status: TaskStatus,
    pub current_attempt: u32,
    pub max_retries: u32,
    pub created_at: DateTime<Utc>,
    pub started_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
    pub last_error: Option<String>,
    pub response_status: Option<u16>,
    pub response_body: Option<serde_json::Value>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Task {
    pub task_id: Uuid,
    pub idempotency_key: String,
    pub url: String,
    pub body: Option<serde_json::Value>,
    pub status: TaskStatus,
    pub current_attempt: u32,
    pub max_retries: u32,
    pub retry_strategy: RetryStrategy,
    pub base_interval_ms: u64,
    pub jitter_max_ms: u64,
    pub http_timeout_seconds: u64,
    pub created_at: DateTime<Utc>,
    pub started_at: Option<DateTime<Utc>>,
    pub completed_at: Option<DateTime<Utc>>,
    pub last_error: Option<String>,
    pub response_status: Option<u16>,
    pub response_body: Option<serde_json::Value>,
    pub execution_started_at: Option<DateTime<Utc>>,
}

impl Task {
    pub fn new(req: TaskRequest, idempotency_key: String) -> Self {
        Self {
            task_id: Uuid::new_v4(),
            idempotency_key,
            url: req.url,
            body: req.body,
            status: TaskStatus::Pending,
            current_attempt: 0,
            max_retries: req.max_retries,
            retry_strategy: req.retry_strategy,
            base_interval_ms: req.base_interval_ms,
            jitter_max_ms: req.jitter_max_ms,
            http_timeout_seconds: req.http_timeout_seconds,
            created_at: Utc::now(),
            started_at: None,
            completed_at: None,
            last_error: None,
            response_status: None,
            response_body: None,
            execution_started_at: None,
        }
    }

    pub fn to_detail(&self) -> TaskDetail {
        TaskDetail {
            task_id: self.task_id,
            idempotency_key: self.idempotency_key.clone(),
            url: self.url.clone(),
            status: self.status,
            current_attempt: self.current_attempt,
            max_retries: self.max_retries,
            created_at: self.created_at,
            started_at: self.started_at,
            completed_at: self.completed_at,
            last_error: self.last_error.clone(),
            response_status: self.response_status,
            response_body: self.response_body.clone(),
        }
    }

    pub fn is_terminal(&self) -> bool {
        matches!(self.status, TaskStatus::Succeeded | TaskStatus::Failed)
    }

    pub fn get_retry_delay_ms(&self) -> u64 {
        use rand::Rng;
        let attempt = self.current_attempt.saturating_sub(1);
        match self.retry_strategy {
            RetryStrategy::Fixed => self.base_interval_ms,
            RetryStrategy::ExponentialBackoff => {
                let base = self.base_interval_ms;
                let exp = 2u64.saturating_pow(attempt);
                let delay = base.saturating_mul(exp);
                let jitter = if self.jitter_max_ms > 0 {
                    rand::thread_rng().gen_range(0..=self.jitter_max_ms)
                } else {
                    0
                };
                delay.saturating_add(jitter)
            }
        }
    }
}

#[derive(Debug, Clone, Serialize)]
pub struct StatsResponse {
    pub total_tasks: u64,
    pub by_status: ByStatusStats,
    pub avg_execution_time_ms: Option<f64>,
    pub retry_distribution: RetryDistribution,
}

#[derive(Debug, Clone, Serialize, Default)]
pub struct ByStatusStats {
    pub pending: u64,
    pub running: u64,
    pub retrying: u64,
    pub succeeded: u64,
    pub failed: u64,
}

#[derive(Debug, Clone, Serialize, Default)]
pub struct RetryDistribution {
    pub zero: u64,
    pub one: u64,
    pub two: u64,
    pub three: u64,
    pub four_plus: u64,
}
