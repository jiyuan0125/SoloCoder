use serde::{Deserialize, Serialize};
use chrono::{DateTime, Utc};
use std::collections::HashMap;

#[derive(Debug, Clone, Deserialize, Serialize)]
pub struct Span {
    pub trace_id: String,
    pub span_id: String,
    pub parent_span_id: Option<String>,
    pub service_name: String,
    pub operation_name: String,
    pub start_time: DateTime<Utc>,
    pub duration: u64,
    pub tags: HashMap<String, String>,
}

#[derive(Debug, Clone, Serialize)]
pub struct SpanWithSlow {
    #[serde(flatten)]
    pub span: Span,
    pub slow: bool,
}

#[derive(Debug, Clone)]
pub struct Trace {
    pub trace_id: String,
    pub spans: Vec<Span>,
    pub earliest_start: DateTime<Utc>,
    pub total_duration: u64,
}

#[derive(Debug, Clone, Serialize)]
pub struct TraceResponse {
    pub trace_id: String,
    pub spans: Vec<SpanWithSlow>,
    pub total_duration: u64,
}

#[derive(Debug, Clone, Serialize)]
pub struct TraceListItem {
    pub trace_id: String,
    pub services: Vec<String>,
    pub total_duration: u64,
    pub has_slow_spans: bool,
    pub start_time: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize)]
pub struct DependencyGraph {
    pub nodes: Vec<String>,
    pub edges: Vec<DependencyEdge>,
    pub cycles: Vec<Vec<String>>,
}

#[derive(Debug, Clone, Serialize)]
pub struct DependencyEdge {
    pub from: String,
    pub to: String,
    pub call_count: u64,
    pub avg_latency_ms: f64,
}

#[derive(Debug, Clone, Serialize)]
pub struct Stats {
    pub total_traces: u64,
    pub avg_latency_ms: f64,
    pub p50_ms: f64,
    pub p95_ms: f64,
    pub p99_ms: f64,
    pub slow_span_ratio: f64,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TraceQuery {
    pub service_name: Option<String>,
    pub min_duration: Option<u64>,
    pub start_time: Option<DateTime<Utc>>,
    pub end_time: Option<DateTime<Utc>>,
}
