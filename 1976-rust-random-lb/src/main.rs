use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{delete, get, post, put},
    Json, Router,
};
use chrono::{DateTime, Utc};
use rand::Rng;
use serde::{Deserialize, Serialize};
use std::collections::VecDeque;
use std::sync::Arc;
use tokio::sync::Mutex;
use uuid::Uuid;

const HOTSPOT_WINDOW_SECONDS: i64 = 10;
const PENALTY_FACTOR: f64 = 0.5;
const DEVIATION_THRESHOLD: f64 = 0.10;

#[derive(Debug, Clone)]
struct Backend {
    id: Uuid,
    address: String,
    weight: u32,
    selected_count: u64,
    response_times: VecDeque<u64>,
    max_response_times: usize,
    active_connections: u32,
    recent_selections: VecDeque<DateTime<Utc>>,
    created_at: DateTime<Utc>,
}

impl Backend {
    fn new(address: String, weight: u32) -> Self {
        Self {
            id: Uuid::new_v4(),
            address,
            weight,
            selected_count: 0,
            response_times: VecDeque::new(),
            max_response_times: 100,
            active_connections: 0,
            recent_selections: VecDeque::new(),
            created_at: Utc::now(),
        }
    }

    fn cleanup_expired_selections(&mut self) {
        let now = Utc::now();
        let cutoff = now - chrono::Duration::seconds(HOTSPOT_WINDOW_SECONDS);
        while let Some(&selection) = self.recent_selections.front() {
            if selection < cutoff {
                self.recent_selections.pop_front();
            } else {
                break;
            }
        }
    }

    fn get_adjusted_weight(&self) -> f64 {
        let recent_count = self.get_recent_count() as f64;
        (self.weight as f64) / (1.0 + recent_count * PENALTY_FACTOR)
    }

    fn record_selection(&mut self) {
        self.selected_count += 1;
        self.recent_selections.push_back(Utc::now());
        while self.recent_selections.len() > 1000 {
            self.recent_selections.pop_front();
        }
        self.cleanup_expired_selections();
    }

    fn record_response_time(&mut self, duration_ms: u64) {
        self.response_times.push_back(duration_ms);
        while self.response_times.len() > self.max_response_times {
            self.response_times.pop_front();
        }
    }

    fn get_avg_response_time(&self) -> Option<f64> {
        if self.response_times.is_empty() {
            return None;
        }
        let sum: u64 = self.response_times.iter().sum();
        Some(sum as f64 / self.response_times.len() as f64)
    }

    fn get_recent_count(&self) -> usize {
        let now = Utc::now();
        let cutoff = now - chrono::Duration::seconds(HOTSPOT_WINDOW_SECONDS);
        self.recent_selections
            .iter()
            .filter(|&&t| t >= cutoff)
            .count()
    }
}

struct AppState {
    backends: Mutex<Vec<Backend>>,
    total_selections: Mutex<u64>,
}

#[derive(Debug, Deserialize)]
struct CreateBackendRequest {
    address: String,
    weight: u32,
}

#[derive(Debug, Deserialize)]
struct UpdateWeightRequest {
    weight: u32,
}

#[derive(Debug, Serialize)]
struct BackendResponse {
    id: Uuid,
    address: String,
    weight: u32,
    created_at: DateTime<Utc>,
}

#[derive(Debug, Serialize)]
struct BackendStats {
    id: Uuid,
    address: String,
    config_weight: u32,
    adjusted_weight: f64,
    selected_count: u64,
    actual_traffic_ratio: f64,
    avg_response_time_ms: Option<f64>,
    active_connections: u32,
    recent_selections_count: usize,
    deviation_exceeds_threshold: bool,
    deviation_percentage: f64,
}

#[derive(Debug, Serialize)]
struct SelectResult {
    selected_backend: BackendResponse,
    selection_process: SelectionProcess,
}

#[derive(Debug, Serialize)]
struct SelectionProcess {
    candidates: Vec<CandidateInfo>,
    random_value: f64,
    total_adjusted_weight: f64,
    cumulative_weights: Vec<(Uuid, f64, String)>,
}

#[derive(Debug, Serialize)]
struct CandidateInfo {
    id: Uuid,
    address: String,
    base_weight: u32,
    adjusted_weight: f64,
    recent_count: usize,
}

async fn create_backend(
    State(state): State<Arc<AppState>>,
    Json(payload): Json<CreateBackendRequest>,
) -> impl IntoResponse {
    if payload.weight == 0 {
        return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "Weight must be greater than 0"})),
        );
    }

    let backend = Backend::new(payload.address.clone(), payload.weight);
    let response = BackendResponse {
        id: backend.id,
        address: backend.address.clone(),
        weight: backend.weight,
        created_at: backend.created_at,
    };

    state.backends.lock().await.push(backend);

    (StatusCode::CREATED, Json(serde_json::json!(response)))
}

async fn list_backends(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let backends = state.backends.lock().await;
    let responses: Vec<BackendResponse> = backends
        .iter()
        .map(|b| BackendResponse {
            id: b.id,
            address: b.address.clone(),
            weight: b.weight,
            created_at: b.created_at,
        })
        .collect();

    (StatusCode::OK, Json(serde_json::json!(responses)))
}

async fn update_backend_weight(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(payload): Json<UpdateWeightRequest>,
) -> impl IntoResponse {
    if payload.weight == 0 {
        return (
            StatusCode::BAD_REQUEST,
            Json(serde_json::json!({"error": "Weight must be greater than 0"})),
        );
    }

    let mut backends = state.backends.lock().await;
    if let Some(backend) = backends.iter_mut().find(|b| b.id == id) {
        backend.weight = payload.weight;
        let response = BackendResponse {
            id: backend.id,
            address: backend.address.clone(),
            weight: backend.weight,
            created_at: backend.created_at,
        };
        return (StatusCode::OK, Json(serde_json::json!(response)));
    }

    (
        StatusCode::NOT_FOUND,
        Json(serde_json::json!({"error": "Backend not found"})),
    )
}

async fn delete_backend(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    let mut backends = state.backends.lock().await;
    let initial_len = backends.len();
    backends.retain(|b| b.id != id);

    if backends.len() < initial_len {
        return (StatusCode::NO_CONTENT, Json(serde_json::json!({})));
    }

    (
        StatusCode::NOT_FOUND,
        Json(serde_json::json!({"error": "Backend not found"})),
    )
}

async fn get_backends_stats(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let backends = state.backends.lock().await;
    let total_selections = *state.total_selections.lock().await;

    let total_config_weight: u32 = backends.iter().map(|b| b.weight).sum();

    let stats: Vec<BackendStats> = backends
        .iter()
        .map(|b| {
            let actual_ratio = if total_selections > 0 {
                b.selected_count as f64 / total_selections as f64
            } else {
                0.0
            };

            let expected_ratio = if total_config_weight > 0 {
                b.weight as f64 / total_config_weight as f64
            } else {
                0.0
            };

            let deviation_percentage = if expected_ratio > 0.0 {
                (actual_ratio - expected_ratio) / expected_ratio
            } else {
                0.0
            };

            let deviation_exceeds_threshold =
                total_selections > 0 && deviation_percentage.abs() > DEVIATION_THRESHOLD;

            BackendStats {
                id: b.id,
                address: b.address.clone(),
                config_weight: b.weight,
                adjusted_weight: b.get_adjusted_weight(),
                selected_count: b.selected_count,
                actual_traffic_ratio: actual_ratio,
                avg_response_time_ms: b.get_avg_response_time(),
                active_connections: b.active_connections,
                recent_selections_count: b.get_recent_count(),
                deviation_exceeds_threshold,
                deviation_percentage,
            }
        })
        .collect();

    (
        StatusCode::OK,
        Json(serde_json::json!({
            "total_selections": total_selections,
            "total_config_weight": total_config_weight,
            "backends": stats,
            "deviation_threshold_percentage": DEVIATION_THRESHOLD * 100.0,
            "hotspot_window_seconds": HOTSPOT_WINDOW_SECONDS,
            "penalty_factor": PENALTY_FACTOR,
        })),
    )
}

async fn select_backend(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let mut backends = state.backends.lock().await;
    let mut total_selections = state.total_selections.lock().await;

    if backends.is_empty() {
        return (
            StatusCode::SERVICE_UNAVAILABLE,
            Json(serde_json::json!({"error": "No backends available"})),
        );
    }

    let total_adjusted_weight: f64 = backends.iter().map(|b| b.get_adjusted_weight()).sum();

    if total_adjusted_weight <= 0.0 {
        return (
            StatusCode::SERVICE_UNAVAILABLE,
            Json(serde_json::json!({"error": "No valid backends available"})),
        );
    }

    let candidates: Vec<CandidateInfo> = backends
        .iter()
        .map(|b| CandidateInfo {
            id: b.id,
            address: b.address.clone(),
            base_weight: b.weight,
            adjusted_weight: b.get_adjusted_weight(),
            recent_count: b.get_recent_count(),
        })
        .collect();

    let mut rng = rand::thread_rng();
    let random_value = rng.gen::<f64>() * total_adjusted_weight;

    let mut cumulative = 0.0;
    let mut cumulative_weights = Vec::new();
    let mut selected_id: Option<Uuid> = None;

    for backend in backends.iter() {
        let adjusted = backend.get_adjusted_weight();
        cumulative_weights.push((backend.id, cumulative, backend.address.clone()));
        
        cumulative += adjusted;
        if random_value <= cumulative && selected_id.is_none() {
            selected_id = Some(backend.id);
        }
    }

    let selected_id = selected_id.expect("Should have selected a backend");

    let selected_backend = backends
        .iter_mut()
        .find(|b| b.id == selected_id)
        .expect("Selected backend should exist");

    selected_backend.record_selection();
    *total_selections += 1;

    let response = BackendResponse {
        id: selected_backend.id,
        address: selected_backend.address.clone(),
        weight: selected_backend.weight,
        created_at: selected_backend.created_at,
    };

    let selection_process = SelectionProcess {
        candidates,
        random_value,
        total_adjusted_weight,
        cumulative_weights,
    };

    let result = SelectResult {
        selected_backend: response,
        selection_process,
    };

    (StatusCode::OK, Json(serde_json::json!(result)))
}

#[tokio::main]
async fn main() {
    let state = Arc::new(AppState {
        backends: Mutex::new(Vec::new()),
        total_selections: Mutex::new(0),
    });

    let app = Router::new()
        .route("/backends", post(create_backend).get(list_backends))
        .route("/backends/:id", put(update_backend_weight).delete(delete_backend))
        .route("/backends/stats", get(get_backends_stats))
        .route("/select", get(select_backend))
        .with_state(state);

    let port = std::env::var("PORT")
        .unwrap_or_else(|_| "3000".to_string())
        .parse::<u16>()
        .expect("PORT must be a valid port number");

    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], port));
    println!("Listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
