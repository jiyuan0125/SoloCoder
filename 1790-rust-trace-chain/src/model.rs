use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use uuid::Uuid;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Span {
    pub trace_id: String,
    pub span_id: String,
    pub parent_span_id: Option<String>,
    pub start_time: DateTime<Utc>,
    pub end_time: Option<DateTime<Utc>>,
    pub operation_name: String,
    #[serde(default)]
    pub tags: std::collections::HashMap<String, String>,
}

impl Span {
    pub fn new(
        trace_id: Option<String>,
        parent_span_id: Option<String>,
        operation_name: String,
        tags: std::collections::HashMap<String, String>,
    ) -> Self {
        let trace_id = trace_id.unwrap_or_else(|| Uuid::new_v4().to_string());
        let span_id = Uuid::new_v4().to_string();
        
        Self {
            trace_id,
            span_id,
            parent_span_id,
            start_time: Utc::now(),
            end_time: None,
            operation_name,
            tags,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateSpanRequest {
    pub trace_id: Option<String>,
    pub parent_span_id: Option<String>,
    pub operation_name: String,
    #[serde(default)]
    pub tags: std::collections::HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EndSpanRequest {
    pub span_id: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SearchByTagsRequest {
    pub tags: std::collections::HashMap<String, String>,
    #[serde(default = "default_page")]
    pub page: usize,
    #[serde(default = "default_page_size")]
    pub page_size: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PaginatedResponse<T> {
    pub total: usize,
    pub page: usize,
    pub page_size: usize,
    pub items: Vec<T>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CleanupByTimeRequest {
    pub hours: Option<i64>,
}

fn default_page() -> usize { 1 }
fn default_page_size() -> usize { 20 }

#[derive(Debug, Clone, Serialize)]
pub struct TraceSpanDisplay {
    pub span: Span,
    pub indent_level: usize,
}
