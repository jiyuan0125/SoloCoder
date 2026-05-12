use std::collections::HashMap;
use std::sync::Arc;
use std::time::{Duration, SystemTime, UNIX_EPOCH};

use axum::{
    extract::{Query, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use tokio::time::interval;

const RETENTION_PERIOD: Duration = Duration::from_secs(24 * 60 * 60);

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
enum MetricType {
    Counter,
    Gauge,
    Histogram,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct MetricDef {
    name: String,
    #[serde(rename = "type")]
    metric_type: MetricType,
    description: String,
    labels: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct ReportRequest {
    name: String,
    labels: HashMap<String, String>,
    value: f64,
}

#[derive(Debug, Clone)]
struct DataPoint {
    timestamp: u64,
    value: f64,
}

type LabelsKey = Vec<(String, String)>;

fn labels_to_key(labels: &HashMap<String, String>) -> LabelsKey {
    let mut v: Vec<(String, String)> = labels.iter().map(|(k, v)| (k.clone(), v.clone())).collect();
    v.sort();
    v
}

fn labels_match_filter(labels: &[(String, String)], filter: &HashMap<String, String>) -> bool {
    for (k, v) in filter {
        let mut found = false;
        for (lk, lv) in labels {
            if lk == k && lv == v {
                found = true;
                break;
            }
        }
        if !found {
            return false;
        }
    }
    true
}

#[derive(Debug, Clone)]
struct TimeSeries {
    def: MetricDef,
    series: HashMap<LabelsKey, Vec<DataPoint>>,
}

#[derive(Debug, Clone, Default)]
struct MetricsStore {
    metrics: HashMap<String, TimeSeries>,
}

impl MetricsStore {
    fn register(&mut self, def: MetricDef) -> Result<(), String> {
        if let Some(existing) = self.metrics.get(&def.name) {
            if existing.def.metric_type != def.metric_type {
                return Err(format!(
                    "metric type mismatch: expected {:?}, got {:?}",
                    existing.def.metric_type, def.metric_type
                ));
            }
            return Ok(());
        }
        self.metrics.insert(
            def.name.clone(),
            TimeSeries {
                def,
                series: HashMap::new(),
            },
        );
        Ok(())
    }

    fn report(&mut self, req: ReportRequest) -> Result<(), String> {
        let now = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs();

        let ts = match self.metrics.get_mut(&req.name) {
            Some(ts) => ts,
            None => return Err(format!("metric '{}' not registered", req.name)),
        };

        match ts.def.metric_type {
            MetricType::Counter => {
                if req.value < 0.0 {
                    return Err("counter value cannot be negative".to_string());
                }
            }
            MetricType::Gauge => {}
            MetricType::Histogram => {}
        }

        let key = labels_to_key(&req.labels);
        let points = ts.series.entry(key).or_default();
        points.push(DataPoint {
            timestamp: now,
            value: req.value,
        });

        Ok(())
    }

    fn cleanup(&mut self) {
        let cutoff = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
            .saturating_sub(RETENTION_PERIOD.as_secs() as u64);

        for ts in self.metrics.values_mut() {
            ts.series.retain(|_, points| {
                points.retain(|p| p.timestamp >= cutoff);
                !points.is_empty()
            });
        }
    }

    fn query(&self, name: &str, filter: &HashMap<String, String>) -> Vec<QueryResult> {
        let ts = match self.metrics.get(name) {
            Some(ts) => ts,
            None => return vec![],
        };

        let cutoff = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .unwrap_or_default()
            .as_secs()
            .saturating_sub(RETENTION_PERIOD.as_secs() as u64);

        let mut results = Vec::new();

        for (key, points) in &ts.series {
            if !labels_match_filter(key, filter) {
                continue;
            }

            let labels: HashMap<String, String> = key.iter().cloned().collect();

            match ts.def.metric_type {
                MetricType::Counter => {
                    let mut sum = 0.0;
                    for p in points {
                        if p.timestamp >= cutoff {
                            sum += p.value;
                        }
                    }
                    results.push(QueryResult {
                        name: name.to_string(),
                        metric_type: MetricType::Counter,
                        labels,
                        value: sum,
                        p50: None,
                        p90: None,
                        p99: None,
                    });
                }
                MetricType::Gauge => {
                    let filtered: Vec<f64> = points
                        .iter()
                        .filter(|p| p.timestamp >= cutoff)
                        .map(|p| p.value)
                        .collect();
                    let value = filtered.last().copied().unwrap_or(0.0);
                    results.push(QueryResult {
                        name: name.to_string(),
                        metric_type: MetricType::Gauge,
                        labels,
                        value,
                        p50: None,
                        p90: None,
                        p99: None,
                    });
                }
                MetricType::Histogram => {
                        let mut values: Vec<f64> = points
                            .iter()
                            .filter(|p| p.timestamp >= cutoff)
                            .map(|p| p.value)
                            .collect();
                        values.sort_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal));

                        let p50 = percentile(&values, 0.5);
                        let p90 = percentile(&values, 0.9);
                        let p99 = percentile(&values, 0.99);

                        let sum: f64 = values.iter().sum();

                        results.push(QueryResult {
                            name: name.to_string(),
                            metric_type: MetricType::Histogram,
                            labels,
                            value: sum,
                            p50,
                            p90,
                            p99,
                        });
                    }
            }
        }

        results
    }

    fn export(&self) -> String {
        let mut output = String::new();

        for (name, ts) in &self.metrics {
            let type_str = match ts.def.metric_type {
                MetricType::Counter => "counter",
                MetricType::Gauge => "gauge",
                MetricType::Histogram => "histogram",
            };

            output.push_str(&format!("# HELP {} {}\n", name, ts.def.description));
            output.push_str(&format!("# TYPE {} {}\n", name, type_str));

            let cutoff = SystemTime::now()
                .duration_since(UNIX_EPOCH)
                .unwrap_or_default()
                .as_secs()
                .saturating_sub(RETENTION_PERIOD.as_secs() as u64);

            for (key, points) in &ts.series {
                let filtered: Vec<&DataPoint> = points
                    .iter()
                    .filter(|p| p.timestamp >= cutoff)
                    .collect();

                if filtered.is_empty() {
                    continue;
                }

                let labels_str = if key.is_empty() {
                    String::new()
                } else {
                    let parts: Vec<String> = key
                        .iter()
                        .map(|(k, v)| format!("{}=\"{}\"", k, escape_label_value(v)))
                        .collect();
                    format!("{{{}}}", parts.join(","))
                };

                match ts.def.metric_type {
                    MetricType::Counter => {
                        let sum: f64 = filtered.iter().map(|p| p.value).sum();
                        output.push_str(&format!("{}{} {}\n", name, labels_str, sum));
                    }
                    MetricType::Gauge => {
                        if let Some(last) = filtered.last() {
                            output.push_str(&format!(
                                "{}{} {}\n",
                                name, labels_str, last.value
                            ));
                        }
                    }
                    MetricType::Histogram => {
                        let mut values: Vec<f64> =
                            filtered.iter().map(|p| p.value).collect();
                        values
                            .sort_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal));

                        let count = values.len() as f64;
                        let sum: f64 = values.iter().sum();

                        let labels_open = if labels_str.is_empty() {
                            "{".to_string()
                        } else {
                            format!("{},", labels_str.trim_end_matches('}'))
                        };

                        let buckets = [0.0, 1.0, 5.0, 10.0, 25.0, 50.0, 100.0, 250.0, 500.0];
                        for &b in &buckets {
                            let le_count = values.iter().filter(|&&v| v <= b).count();
                            output.push_str(&format!(
                                "{}_bucket{}le=\"{}\"}} {}\n",
                                name, labels_open, b, le_count
                            ));
                        }
                        output.push_str(&format!(
                            "{}_bucket{}le=\"+Inf\"}} {}\n",
                            name, labels_open, values.len()
                        ));
                        output.push_str(&format!("{}_sum{} {}\n", name, labels_str, sum));
                        output.push_str(&format!("{}_count{} {}\n", name, labels_str, count));
                    }
                }
            }
        }

        output
    }
}

fn escape_label_value(v: &str) -> String {
    v.replace('\\', "\\\\")
        .replace('"', "\\\"")
        .replace('\n', "\\n")
}

fn percentile(values: &[f64], p: f64) -> Option<f64> {
    if values.is_empty() {
        return None;
    }
    if p <= 0.0 {
        return Some(values[0]);
    }
    if p >= 1.0 {
        return Some(values[values.len() - 1]);
    }

    let n = values.len() as f64;
    let rank = p * (n - 1.0);
    let lower = rank.floor() as usize;
    let upper = rank.ceil() as usize;
    let weight = rank - lower as f64;

    if lower == upper {
        Some(values[lower])
    } else {
        Some(values[lower] * (1.0 - weight) + values[upper] * weight)
    }
}

#[derive(Debug, Clone, Serialize)]
struct QueryResult {
    name: String,
    #[serde(rename = "type")]
    metric_type: MetricType,
    labels: HashMap<String, String>,
    value: f64,
    #[serde(skip_serializing_if = "Option::is_none")]
    p50: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    p90: Option<f64>,
    #[serde(skip_serializing_if = "Option::is_none")]
    p99: Option<f64>,
}

type AppState = Arc<RwLock<MetricsStore>>;

async fn register(
    State(state): State<AppState>,
    Json(def): Json<MetricDef>,
) -> Response {
    let mut store = state.write().await;
    match store.register(def) {
        Ok(_) => (StatusCode::OK, Json(serde_json::json!({"status": "ok"}))).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e}))).into_response(),
    }
}

async fn report(
    State(state): State<AppState>,
    Json(req): Json<ReportRequest>,
) -> Response {
    let mut store = state.write().await;
    match store.report(req) {
        Ok(_) => (StatusCode::OK, Json(serde_json::json!({"status": "ok"}))).into_response(),
        Err(e) => (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": e}))).into_response(),
    }
}

async fn query(
    State(state): State<AppState>,
    Query(params): Query<HashMap<String, String>>,
) -> Response {
    let name = match params.get("name") {
        Some(n) => n.clone(),
        None => return (StatusCode::BAD_REQUEST, Json(serde_json::json!({"error": "name parameter required"}))).into_response(),
    };

    let mut filter = params.clone();
    filter.remove("name");

    let store = state.read().await;
    let results = store.query(&name, &filter);
    (StatusCode::OK, Json(results)).into_response()
}

async fn export(State(state): State<AppState>) -> Response {
    let store = state.read().await;
    let output = store.export();
    (
        StatusCode::OK,
        [(axum::http::header::CONTENT_TYPE, "text/plain; version=0.0.4")],
        output,
    )
        .into_response()
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .init();

    let store: AppState = Arc::new(RwLock::new(MetricsStore::default()));

    {
        let store = Arc::clone(&store);
        tokio::spawn(async move {
            let mut ticker = interval(Duration::from_secs(60));
            loop {
                ticker.tick().await;
                let mut s = store.write().await;
                s.cleanup();
            }
        });
    }

    let app = Router::new()
        .route("/metrics/register", post(register))
        .route("/metrics/report", post(report))
        .route("/metrics/query", get(query))
        .route("/metrics/export", get(export))
        .with_state(store);

    let port = std::env::var("PORT").unwrap_or_else(|_| "3000".to_string());
    let addr = format!("0.0.0.0:{}", port).parse().unwrap();

    tracing::info!("listening on {}", addr);
    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}
