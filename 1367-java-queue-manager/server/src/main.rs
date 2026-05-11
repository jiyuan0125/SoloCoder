use std::sync::Arc;
use tokio::sync::RwLock;
use axum::{
    Router,
    routing::{get, post},
    extract::{State, Path, Json},
    http::StatusCode,
    response::{IntoResponse, Response},
};
use clap::Parser;
use tower_http::cors::{Any, CorsLayer};
use serde::Serialize;

use queue_core::QueueManager;
use queue_core::models::{ServiceType, TicketDto, WindowDto};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
}

type AppState = Arc<RwLock<QueueManager>>;

fn json_response<T: Serialize>(status: StatusCode, data: &T) -> Response {
    (status, Json(data)).into_response()
}

fn error_response(status: StatusCode, message: &str) -> Response {
    let body = serde_json::json!({"error": message});
    (status, Json(body)).into_response()
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let mut queue_manager = QueueManager::new();
    
    let window1_types = std::collections::HashSet::from([ServiceType::A]);
    let window2_types = std::collections::HashSet::from([ServiceType::A, ServiceType::B]);
    let window3_types = std::collections::HashSet::from([ServiceType::B, ServiceType::C]);
    
    queue_manager.add_window("窗口1".to_string(), window1_types);
    queue_manager.add_window("窗口2".to_string(), window2_types);
    queue_manager.add_window("窗口3".to_string(), window3_types);
    
    let state: AppState = Arc::new(RwLock::new(queue_manager));
    
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);
    
    let app = Router::new()
        .route("/health", get(health_check))
        .route("/windows", get(list_windows))
        .route("/windows/:id", get(get_window))
        .route("/windows/:id/call", post(call_next))
        .route("/windows/:id/recall", post(recall_ticket))
        .route("/windows/:id/arrived", post(mark_arrived))
        .route("/windows/:id/complete", post(complete_ticket))
        .route("/windows/:id/close", post(close_window))
        .route("/windows/:id/open", post(open_window))
        .route("/tickets/generate", post(generate_ticket))
        .route("/tickets/:display", get(get_ticket))
        .route("/queues/:service_type", get(get_queue_length))
        .with_state(state)
        .layer(cors);
    
    let addr = std::net::SocketAddr::from(([0, 0, 0, 0], args.port));
    println!("Server running on {}", addr);
    
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn health_check() -> &'static str {
    "OK"
}

async fn list_windows(State(state): State<AppState>) -> Json<Vec<WindowDto>> {
    let manager = state.read().await;
    let windows: Vec<_> = manager.get_all_windows().iter().map(|w| w.to_dto()).collect();
    Json(windows)
}

async fn get_window(
    State(state): State<AppState>,
    Path(id): Path<u32>,
) -> Response {
    let manager = state.read().await;
    match manager.get_window(id) {
        Some(window) => json_response(StatusCode::OK, &window.to_dto()),
        None => error_response(StatusCode::NOT_FOUND, "Window not found"),
    }
}

#[derive(serde::Deserialize)]
struct GenerateTicketRequest {
    service_type: String,
    is_vip: bool,
}

async fn generate_ticket(
    State(state): State<AppState>,
    Json(req): Json<GenerateTicketRequest>,
) -> Response {
    let service_type = match req.service_type.parse::<ServiceType>() {
        Ok(st) => st,
        Err(e) => return error_response(StatusCode::BAD_REQUEST, &e),
    };
    
    let mut manager = state.write().await;
    let ticket = manager.generate_ticket(service_type, req.is_vip);
    json_response(StatusCode::CREATED, &ticket.to_dto())
}

async fn get_ticket(
    State(state): State<AppState>,
    Path(display): Path<String>,
) -> Response {
    let manager = state.read().await;
    match manager.get_ticket(&display) {
        Some(ticket) => json_response(StatusCode::OK, &ticket.to_dto()),
        None => error_response(StatusCode::NOT_FOUND, "Ticket not found"),
    }
}

async fn get_queue_length(
    State(state): State<AppState>,
    Path(service_type): Path<String>,
) -> Response {
    let st = match service_type.parse::<ServiceType>() {
        Ok(st) => st,
        Err(e) => return error_response(StatusCode::BAD_REQUEST, &e),
    };
    
    let manager = state.read().await;
    let length = manager.get_queue_length(&st);
    let body = serde_json::json!({"length": length});
    json_response(StatusCode::OK, &body)
}

async fn call_next(
    State(state): State<AppState>,
    Path(id): Path<u32>,
) -> Response {
    let mut manager = state.write().await;
    match manager.call_next(id) {
        Ok(ticket) => json_response(StatusCode::OK, &ticket.to_dto()),
        Err(e) => error_response(StatusCode::BAD_REQUEST, &e.to_string()),
    }
}

async fn recall_ticket(
    State(state): State<AppState>,
    Path(id): Path<u32>,
) -> Response {
    let mut manager = state.write().await;
    match manager.recall_ticket(id) {
        Ok(ticket) => json_response(StatusCode::OK, &ticket.to_dto()),
        Err(e) => error_response(StatusCode::BAD_REQUEST, &e.to_string()),
    }
}

async fn mark_arrived(
    State(state): State<AppState>,
    Path(id): Path<u32>,
) -> Response {
    let mut manager = state.write().await;
    match manager.mark_ticket_arrived(id) {
        Ok(ticket) => json_response(StatusCode::OK, &ticket.to_dto()),
        Err(e) => error_response(StatusCode::BAD_REQUEST, &e.to_string()),
    }
}

async fn complete_ticket(
    State(state): State<AppState>,
    Path(id): Path<u32>,
) -> Response {
    let mut manager = state.write().await;
    match manager.complete_ticket(id) {
        Ok(ticket) => json_response(StatusCode::OK, &ticket.to_dto()),
        Err(e) => error_response(StatusCode::BAD_REQUEST, &e.to_string()),
    }
}

async fn close_window(
    State(state): State<AppState>,
    Path(id): Path<u32>,
) -> Response {
    let mut manager = state.write().await;
    match manager.close_window(id) {
        Ok(()) => {
            let body = serde_json::json!({"status": "closed"});
            json_response(StatusCode::OK, &body)
        }
        Err(e) => error_response(StatusCode::BAD_REQUEST, &e.to_string()),
    }
}

async fn open_window(
    State(state): State<AppState>,
    Path(id): Path<u32>,
) -> Response {
    let mut manager = state.write().await;
    match manager.open_window(id) {
        Ok(()) => {
            let body = serde_json::json!({"status": "opened"});
            json_response(StatusCode::OK, &body)
        }
        Err(e) => error_response(StatusCode::BAD_REQUEST, &e.to_string()),
    }
}
