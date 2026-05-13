use std::collections::HashMap;
use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};

use axum::{
    extract::{Path, State},
    http::{Request, StatusCode},
    middleware::Next,
    response::{IntoResponse, Response},
    routing::{get, put},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use tokio::sync::Mutex;
use tracing_subscriber;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
enum Algorithm {
    FixedWindow,
    SlidingWindow,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Rule {
    algorithm: Algorithm,
    limit: u64,
}

#[derive(Debug, Clone, Serialize)]
struct PathStats {
    path: String,
    algorithm: Algorithm,
    current_qps: f64,
    throttled_count: u64,
}

#[derive(Debug, Clone, Serialize)]
struct StatsResponse {
    path_stats: Vec<PathStats>,
    active_rules_count: usize,
    last_minute_qps: f64,
    global_qps_limit: u64,
}

#[derive(Debug, Clone, Serialize)]
struct RulesResponse {
    rules: HashMap<String, Rule>,
    global_qps_limit: u64,
}

#[derive(Debug, Clone, Deserialize)]
struct RulePayload {
    algorithm: Algorithm,
    limit: u64,
}

#[derive(Debug, Clone, Deserialize)]
struct GlobalQpsPayload {
    limit: u64,
}

#[derive(Clone)]
struct FixedWindowData {
    window_start: u64,
    count: u64,
}

#[derive(Clone)]
struct SlidingWindowData {
    requests: Vec<u64>,
}

#[derive(Clone)]
struct PathTracker {
    rule: Option<Rule>,
    fixed_window: FixedWindowData,
    sliding_window: SlidingWindowData,
    throttled_count: u64,
    recent_requests: Vec<u64>,
}

struct GlobalState {
    rules: Mutex<HashMap<String, Rule>>,
    path_trackers: Mutex<HashMap<String, PathTracker>>,
    global_qps_limit: Mutex<u64>,
    global_recent_requests: Mutex<Vec<u64>>,
}

fn current_seconds() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_secs()
}

fn get_fixed_window_start() -> u64 {
    current_seconds() / 60 * 60
}

fn normalize_path(path: &str) -> String {
    if !path.starts_with('/') {
        format!("/{}", path)
    } else {
        path.to_string()
    }
}

fn check_fixed_window(tracker: &mut PathTracker, limit: u64) -> bool {
    let window_start = get_fixed_window_start();
    
    if tracker.fixed_window.window_start != window_start {
        tracker.fixed_window.window_start = window_start;
        tracker.fixed_window.count = 0;
    }
    
    if tracker.fixed_window.count < limit {
        tracker.fixed_window.count += 1;
        true
    } else {
        false
    }
}

fn check_sliding_window(tracker: &mut PathTracker, limit: u64) -> bool {
    let now = current_seconds();
    let cutoff = now - 60;
    
    tracker.sliding_window.requests.retain(|&t| t > cutoff);
    
    if tracker.sliding_window.requests.len() < limit as usize {
        tracker.sliding_window.requests.push(now);
        true
    } else {
        false
    }
}

fn calculate_qps(requests: &[u64]) -> f64 {
    let now = current_seconds();
    let cutoff = now - 60;
    let count = requests.iter().filter(|&&t| t > cutoff).count();
    count as f64 / 60.0
}

async fn rate_limit_middleware(
    State(state): State<Arc<GlobalState>>,
    req: Request<axum::body::Body>,
    next: Next,
) -> Response {
    let path = normalize_path(req.uri().path());
    
    if path.starts_with("/rules") || path.starts_with("/stats") || path.starts_with("/config") {
        return next.run(req).await;
    }
    
    let rule_opt = {
        let rules = state.rules.lock().await;
        rules.get(&path).cloned()
    };
    
    let global_allowed = {
        let mut global_reqs = state.global_recent_requests.lock().await;
        let limit = *state.global_qps_limit.lock().await;
        let now = current_seconds();
        let cutoff = now - 60;
        
        global_reqs.retain(|&t| t > cutoff);
        
        if global_reqs.len() < limit as usize {
            global_reqs.push(now);
            true
        } else {
            false
        }
    };
    
    if !global_allowed {
        {
            let mut trackers = state.path_trackers.lock().await;
            let tracker = trackers.entry(path.clone()).or_insert_with(|| PathTracker {
                rule: rule_opt.clone(),
                fixed_window: FixedWindowData { window_start: get_fixed_window_start(), count: 0 },
                sliding_window: SlidingWindowData { requests: Vec::new() },
                throttled_count: 0,
                recent_requests: Vec::new(),
            });
            tracker.throttled_count += 1;
        }
        return (StatusCode::TOO_MANY_REQUESTS, "Too Many Requests").into_response();
    }
    
    let rule = match rule_opt {
        Some(r) => r,
        None => return next.run(req).await,
    };
    
    let allowed = {
        let mut trackers = state.path_trackers.lock().await;
        let tracker = trackers.entry(path.clone()).or_insert_with(|| PathTracker {
            rule: Some(rule.clone()),
            fixed_window: FixedWindowData { window_start: get_fixed_window_start(), count: 0 },
            sliding_window: SlidingWindowData { requests: Vec::new() },
            throttled_count: 0,
            recent_requests: Vec::new(),
        });
        
        tracker.rule = Some(rule.clone());
        
        let now = current_seconds();
        let cutoff = now - 60;
        tracker.recent_requests.retain(|&t| t > cutoff);
        tracker.recent_requests.push(now);
        
        match rule.algorithm {
            Algorithm::FixedWindow => check_fixed_window(tracker, rule.limit),
            Algorithm::SlidingWindow => check_sliding_window(tracker, rule.limit),
        }
    };
    
    if allowed {
        next.run(req).await
    } else {
        {
            let mut trackers = state.path_trackers.lock().await;
            let tracker = trackers.entry(path).or_insert_with(|| PathTracker {
                rule: Some(rule.clone()),
                fixed_window: FixedWindowData { window_start: get_fixed_window_start(), count: 0 },
                sliding_window: SlidingWindowData { requests: Vec::new() },
                throttled_count: 0,
                recent_requests: Vec::new(),
            });
            tracker.throttled_count += 1;
        }
        (StatusCode::TOO_MANY_REQUESTS, "Too Many Requests").into_response()
    }
}

async fn put_rule(
    State(state): State<Arc<GlobalState>>,
    Path(path): Path<String>,
    Json(payload): Json<RulePayload>,
) -> impl IntoResponse {
    let rule = Rule {
        algorithm: payload.algorithm,
        limit: payload.limit,
    };
    
    let normalized_path = normalize_path(&path);
    
    {
        let mut rules = state.rules.lock().await;
        rules.insert(normalized_path, rule);
    }
    
    StatusCode::OK
}

async fn get_rules(State(state): State<Arc<GlobalState>>) -> impl IntoResponse {
    let rules = state.rules.lock().await.clone();
    let global_limit = *state.global_qps_limit.lock().await;
    
    Json(RulesResponse {
        rules,
        global_qps_limit: global_limit,
    })
}

async fn put_global_qps(
    State(state): State<Arc<GlobalState>>,
    Json(payload): Json<GlobalQpsPayload>,
) -> impl IntoResponse {
    {
        let mut limit = state.global_qps_limit.lock().await;
        *limit = payload.limit;
    }
    
    StatusCode::OK
}

async fn get_stats(State(state): State<Arc<GlobalState>>) -> impl IntoResponse {
    let rules = state.rules.lock().await.clone();
    let trackers = state.path_trackers.lock().await.clone();
    let global_limit = *state.global_qps_limit.lock().await;
    
    let global_qps = {
        let global_reqs = state.global_recent_requests.lock().await;
        calculate_qps(&global_reqs)
    };
    
    let mut path_stats = Vec::new();
    for (path, rule) in &rules {
        let (qps, throttled, algo) = match trackers.get(path) {
            Some(t) => {
                (
                    calculate_qps(&t.recent_requests),
                    t.throttled_count,
                    t.rule.as_ref().map(|r| r.algorithm).unwrap_or(rule.algorithm),
                )
            }
            None => (0.0, 0, rule.algorithm),
        };
        
        path_stats.push(PathStats {
            path: path.clone(),
            algorithm: algo,
            current_qps: qps,
            throttled_count: throttled,
        });
    }
    
    Json(StatsResponse {
        path_stats,
        active_rules_count: rules.len(),
        last_minute_qps: global_qps,
        global_qps_limit: global_limit,
    })
}

async fn default_handler() -> impl IntoResponse {
    "OK"
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt::init();
    
    let port: u16 = std::env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse()
        .expect("PORT must be a valid number");
    
    let state = Arc::new(GlobalState {
        rules: Mutex::new(HashMap::new()),
        path_trackers: Mutex::new(HashMap::new()),
        global_qps_limit: Mutex::new(u64::MAX),
        global_recent_requests: Mutex::new(Vec::new()),
    });
    
    let app = Router::new()
        .route("/rules/*path", put(put_rule))
        .route("/rules", get(get_rules))
        .route("/stats", get(get_stats))
        .route("/config/global-qps", put(put_global_qps))
        .route("/*path", get(default_handler).post(default_handler).put(default_handler).delete(default_handler))
        .route("/", get(default_handler).post(default_handler).put(default_handler).delete(default_handler))
        .layer(axum::middleware::from_fn_with_state(
            state.clone(),
            rate_limit_middleware,
        ))
        .with_state(state);
    
    let addr = format!("0.0.0.0:{}", port);
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    println!("Server running on {}", addr);
    
    axum::serve(listener, app).await.unwrap();
}
