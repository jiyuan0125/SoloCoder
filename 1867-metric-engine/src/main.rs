use axum::{
    extract::{Query, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use chrono::{DateTime, Utc};
use dashmap::DashMap;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;
use tokio::time::interval;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
enum MetricType {
    Counter,
    Gauge,
    Histogram,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct MetricMetadata {
    name: String,
    metric_type: MetricType,
    description: String,
    label_keys: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct DataPoint {
    value: f64,
    timestamp: DateTime<Utc>,
}

#[derive(Debug, Clone)]
struct TimeSeries {
    points: Vec<DataPoint>,
}

#[derive(Debug, Clone)]
struct MetricData {
    metadata: MetricMetadata,
    series: DashMap<String, TimeSeries>,
}

#[derive(Debug, Clone)]
struct AppState {
    metrics: DashMap<String, MetricData>,
    retention_period: Duration,
}

#[derive(Debug, Deserialize)]
struct RegisterRequest {
    name: String,
    metric_type: MetricType,
    description: String,
    label_keys: Vec<String>,
}

#[derive(Debug, Deserialize)]
struct ReportRequest {
    name: String,
    labels: HashMap<String, String>,
    value: f64,
}

#[derive(Debug, Deserialize)]
struct QueryParams {
    name: String,
    #[serde(flatten)]
    extra: HashMap<String, String>,
}

#[derive(Debug, Serialize)]
struct QueryResult {
    labels: HashMap<String, String>,
    latest_value: Option<f64>,
    count: usize,
    #[serde(skip_serializing_if = "Option::is_none")]
    percentiles: Option<HashMap<String, f64>>,
}

impl AppState {
    fn new() -> Self {
        Self {
            metrics: DashMap::new(),
            retention_period: Duration::from_secs(24 * 60 * 60),
        }
    }

    fn register(&self, req: RegisterRequest) -> Result<(), (StatusCode, String)> {
        if let Some(existing) = self.metrics.get(&req.name) {
            if existing.metadata.metric_type != req.metric_type {
                return Err((
                    StatusCode::BAD_REQUEST,
                    format!("Metric type mismatch for {}", req.name),
                ));
            }
            return Ok(());
        }

        self.metrics.insert(
            req.name.clone(),
            MetricData {
                metadata: MetricMetadata {
                    name: req.name.clone(),
                    metric_type: req.metric_type,
                    description: req.description,
                    label_keys: req.label_keys,
                },
                series: DashMap::new(),
            },
        );
        Ok(())
    }

    fn report(&self, req: ReportRequest) -> Result<(), (StatusCode, String)> {
        let metric = self
            .metrics
            .get(&req.name)
            .ok_or_else(|| (StatusCode::BAD_REQUEST, "Metric not registered".to_string()))?;

        let metadata = &metric.metadata;
        let mut label_string_parts = Vec::new();
        let mut sorted_labels: Vec<(String, String)> = req.labels.iter()
            .map(|(k, v)| (k.clone(), v.clone()))
            .collect();
        sorted_labels.sort_by(|a, b| a.0.cmp(&b.0));
        
        for (key, value) in &sorted_labels {
            if !metadata.label_keys.contains(key) {
                return Err((
                    StatusCode::BAD_REQUEST,
                    format!("Invalid label key: {}", key),
                ));
            }
            label_string_parts.push(format!("{}={}", key, value));
        }
        let series_key = label_string_parts.join(",");

        if !req.value.is_finite() {
            return Ok(());
        }

        match metadata.metric_type {
            MetricType::Counter => {
                if req.value < 0.0 {
                    return Ok(());
                }
            }
            MetricType::Gauge | MetricType::Histogram => {}
        }

        let point = DataPoint {
            value: req.value,
            timestamp: Utc::now(),
        };

        metric
            .series
            .entry(series_key)
            .and_modify(|s| s.points.push(point.clone()))
            .or_insert_with(|| TimeSeries { points: vec![point] });

        Ok(())
    }

    fn query(&self, params: QueryParams) -> Vec<QueryResult> {
        let metric = match self.metrics.get(&params.name) {
            Some(m) => m,
            None => return Vec::new(),
        };

        let filter_labels: HashMap<String, String> = params
            .extra
            .iter()
            .filter(|(k, _)| *k != "name")
            .map(|(k, v)| (k.clone(), v.clone()))
            .collect();

        let mut results = Vec::new();

        for series_ref in metric.series.iter() {
            let labels = parse_labels(series_ref.key());
            
            let matches = filter_labels
                .iter()
                .all(|(k, v)| labels.get(k) == Some(v));
            
            if !matches {
                continue;
            }

            let points: Vec<DataPoint> = series_ref
                .points
                .iter()
                .filter(|p| {
                    let age = Utc::now().signed_duration_since(p.timestamp);
                    age.num_seconds() <= self.retention_period.as_secs() as i64
                })
                .cloned()
                .collect();

            if points.is_empty() {
                continue;
            }

            let latest_value = points.last().map(|p| p.value);
            let percentiles = match metric.metadata.metric_type {
                MetricType::Histogram => Some(calculate_percentiles(&points)),
                _ => None,
            };

            results.push(QueryResult {
                labels,
                latest_value,
                count: points.len(),
                percentiles,
            });
        }

        results
    }

    fn export(&self) -> String {
        let mut output = String::new();

        for metric_ref in self.metrics.iter() {
            let metadata = &metric_ref.metadata;
            
            if !metadata.description.is_empty() {
                output.push_str(&format!("# HELP {} {}\n", metadata.name, metadata.description));
            }
            output.push_str(&format!("# TYPE {} {}\n", metadata.name, match metadata.metric_type {
                MetricType::Counter => "counter",
                MetricType::Gauge => "gauge",
                MetricType::Histogram => "histogram",
            }));

            for series_ref in metric_ref.series.iter() {
                let labels = parse_labels(series_ref.key());
                let label_str = if labels.is_empty() {
                    String::new()
                } else {
                    let mut parts: Vec<String> = labels
                        .iter()
                        .map(|(k, v)| format!("{}=\"{}\"", k, v))
                        .collect();
                    parts.sort();
                    format!("{{{}}}", parts.join(","))
                };

                let points: Vec<&DataPoint> = series_ref
                    .points
                    .iter()
                    .filter(|p| {
                        let age = Utc::now().signed_duration_since(p.timestamp);
                        age.num_seconds() <= self.retention_period.as_secs() as i64
                    })
                    .collect();

                if let Some(latest) = points.last() {
                    output.push_str(&format!(
                        "{}{} {} {}\n",
                        metadata.name,
                        label_str,
                        latest.value,
                        latest.timestamp.timestamp_millis()
                    ));
                }
            }
        }

        output
    }

    fn cleanup(&self) {
        let cutoff = Utc::now()
            .checked_sub_signed(chrono::Duration::from_std(self.retention_period).unwrap())
            .unwrap();

        for metric_ref in self.metrics.iter() {
            for mut series_ref in metric_ref.series.iter_mut() {
                series_ref
                    .points
                    .retain(|p| p.timestamp >= cutoff);
            }

            metric_ref
                .series
                .retain(|_, s| !s.points.is_empty());
        }
    }
}

fn parse_labels(key: &str) -> HashMap<String, String> {
    if key.is_empty() {
        return HashMap::new();
    }
    
    key.split(',')
        .filter_map(|part| {
            let mut iter = part.splitn(2, '=');
            match (iter.next(), iter.next()) {
                (Some(k), Some(v)) => Some((k.to_string(), v.to_string())),
                _ => None,
            }
        })
        .collect()
}

fn calculate_percentiles(points: &[DataPoint]) -> HashMap<String, f64> {
    let mut values: Vec<f64> = points.iter().map(|p| p.value).collect();
    values.sort_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal));

    let n = values.len();
    if n == 0 {
        return HashMap::new();
    }

    let mut result = HashMap::new();
    result.insert("p50".to_string(), percentile(&values, 50.0));
    result.insert("p90".to_string(), percentile(&values, 90.0));
    result.insert("p99".to_string(), percentile(&values, 99.0));
    result
}

fn percentile(sorted: &[f64], p: f64) -> f64 {
    let n = sorted.len();
    if n == 0 {
        return 0.0;
    }
    let index = ((p / 100.0) * (n - 1) as f64).round() as usize;
    sorted[index]
}

async fn register_handler(
    State(state): State<Arc<AppState>>,
    Json(req): Json<RegisterRequest>,
) -> impl IntoResponse {
    match state.register(req) {
        Ok(()) => (StatusCode::OK, Json(serde_json::json!({"status": "ok"}))),
        Err((status, msg)) => (status, Json(serde_json::json!({"error": msg}))),
    }
}

async fn report_handler(
    State(state): State<Arc<AppState>>,
    Json(req): Json<ReportRequest>,
) -> impl IntoResponse {
    match state.report(req) {
        Ok(()) => (StatusCode::OK, Json(serde_json::json!({"status": "ok"}))),
        Err((status, msg)) => (status, Json(serde_json::json!({"error": msg}))),
    }
}

async fn query_handler(
    State(state): State<Arc<AppState>>,
    Query(params): Query<QueryParams>,
) -> impl IntoResponse {
    let results = state.query(params);
    (StatusCode::OK, Json(results))
}

async fn export_handler(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let content = state.export();
    (
        StatusCode::OK,
        [(axum::http::header::CONTENT_TYPE, "text/plain; version=0.0.4")],
        content,
    )
}

fn app(state: Arc<AppState>) -> Router {
    Router::new()
        .route("/metrics/register", post(register_handler))
        .route("/metrics/report", post(report_handler))
        .route("/metrics/query", get(query_handler))
        .route("/metrics/export", get(export_handler))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let state = Arc::new(AppState::new());
    
    let cleanup_state = Arc::clone(&state);
    tokio::spawn(async move {
        let mut interval = interval(Duration::from_secs(60));
        loop {
            interval.tick().await;
            cleanup_state.cleanup();
            tracing::debug!("Cleanup executed");
        }
    });

    let port: u16 = std::env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse()
        .expect("PORT must be a valid number");

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    tracing::info!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app(state)).await.unwrap();
}
