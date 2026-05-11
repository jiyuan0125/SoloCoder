use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post, put},
    Json, Router,
};
use clap::Parser;
use serde::Serialize;
use std::sync::Arc;
use tokio::sync::Mutex;
use visitor_core::{
    errors::VisitorError,
    models::{
        ApprovalDecision, PasscodeVerification, VisitorRegistration, VisitorUpdate,
    },
    service::VisitorService,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "VISITOR_SERVER_PORT", default_value_t = 3000)]
    port: u16,
}

#[derive(Clone)]
struct AppState {
    service: Arc<Mutex<VisitorService>>,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

fn error_to_response(err: VisitorError) -> (StatusCode, Json<ErrorResponse>) {
    let status = match &err {
        VisitorError::VisitorNotFound => StatusCode::NOT_FOUND,
        VisitorError::PasscodeAlreadyUsed => StatusCode::BAD_REQUEST,
        VisitorError::PasscodeExpired => StatusCode::BAD_REQUEST,
        VisitorError::PasscodeInvalid => StatusCode::BAD_REQUEST,
        VisitorError::InvalidStatusTransition => StatusCode::BAD_REQUEST,
        VisitorError::VisitorRejected => StatusCode::FORBIDDEN,
        VisitorError::PasscodeGenerationFailed => StatusCode::INTERNAL_SERVER_ERROR,
        VisitorError::InvalidInput(_) => StatusCode::BAD_REQUEST,
    };
    (
        status,
        Json(ErrorResponse {
            error: err.to_string(),
        }),
    )
}

async fn register_visitor(
    State(state): State<AppState>,
    Json(registration): Json<VisitorRegistration>,
) -> impl IntoResponse {
    let mut service = state.service.lock().await;
    match service.register_visitor(registration) {
        Ok(record) => (StatusCode::CREATED, Json(record)).into_response(),
        Err(err) => {
            let (status, json) = error_to_response(err);
            (status, json).into_response()
        }
    }
}

async fn update_visitor(
    State(state): State<AppState>,
    Json(update): Json<VisitorUpdate>,
) -> impl IntoResponse {
    let mut service = state.service.lock().await;
    match service.update_visitor(update.id, update.registration) {
        Ok(record) => (StatusCode::OK, Json(record)).into_response(),
        Err(err) => {
            let (status, json) = error_to_response(err);
            (status, json).into_response()
        }
    }
}

async fn approve_visitor(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(decision): Json<ApprovalDecision>,
) -> impl IntoResponse {
    let id = match uuid::Uuid::parse_str(&id) {
        Ok(id) => id,
        Err(_) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    error: "无效的访客ID".to_string(),
                }),
            )
                .into_response();
        }
    };

    let mut service = state.service.lock().await;
    match service.approve_visitor(id, decision) {
        Ok(record) => (StatusCode::OK, Json(record)).into_response(),
        Err(err) => {
            let (status, json) = error_to_response(err);
            (status, json).into_response()
        }
    }
}

async fn verify_passcode(
    State(state): State<AppState>,
    Json(verification): Json<PasscodeVerification>,
) -> impl IntoResponse {
    let mut service = state.service.lock().await;
    match service.verify_passcode(verification) {
        Ok(record) => (StatusCode::OK, Json(record)).into_response(),
        Err(err) => {
            let (status, json) = error_to_response(err);
            (status, json).into_response()
        }
    }
}

async fn sign_out(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let id = match uuid::Uuid::parse_str(&id) {
        Ok(id) => id,
        Err(_) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    error: "无效的访客ID".to_string(),
                }),
            )
                .into_response();
        }
    };

    let mut service = state.service.lock().await;
    match service.sign_out(id) {
        Ok(record) => (StatusCode::OK, Json(record)).into_response(),
        Err(err) => {
            let (status, json) = error_to_response(err);
            (status, json).into_response()
        }
    }
}

async fn get_visitor(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let id = match uuid::Uuid::parse_str(&id) {
        Ok(id) => id,
        Err(_) => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    error: "无效的访客ID".to_string(),
                }),
            )
                .into_response();
        }
    };

    let service = state.service.lock().await;
    match service.get_visitor(id) {
        Some(record) => (StatusCode::OK, Json(record)).into_response(),
        None => (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: "访客记录未找到".to_string(),
            }),
        )
            .into_response(),
    }
}

async fn get_all_visitors(State(state): State<AppState>) -> impl IntoResponse {
    let service = state.service.lock().await;
    let visitors = service.get_all_visitors();
    (StatusCode::OK, Json(visitors)).into_response()
}

async fn get_visitors_by_host(
    State(state): State<AppState>,
    Path(host): Path<String>,
) -> impl IntoResponse {
    let service = state.service.lock().await;
    let visitors = service.get_visitors_by_host(&host);
    (StatusCode::OK, Json(visitors)).into_response()
}

fn create_router(state: AppState) -> Router {
    Router::new()
        .route("/visitors", post(register_visitor).get(get_all_visitors))
        .route("/visitors/update", put(update_visitor))
        .route("/visitors/:id", get(get_visitor))
        .route("/visitors/:id/approve", post(approve_visitor))
        .route("/visitors/:id/signout", post(sign_out))
        .route("/visitors/host/:host", get(get_visitors_by_host))
        .route("/passcode/verify", post(verify_passcode))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let state = AppState {
        service: Arc::new(Mutex::new(VisitorService::new())),
    };

    let app = create_router(state);

    let listener = tokio::net::TcpListener::bind(format!("0.0.0.0:{}", args.port))
        .await
        .expect("Failed to bind TCP listener");

    println!("🚀 Visitor Management Server running on port {}", args.port);
    println!("📋 Available endpoints:");
    println!("  POST   /visitors              - 登记访客");
    println!("  PUT    /visitors/update       - 更新访客信息");
    println!("  GET    /visitors              - 获取所有访客");
    println!("  GET    /visitors/:id          - 获取单个访客");
    println!("  POST   /visitors/:id/approve  - 审批访客");
    println!("  POST   /visitors/:id/signout  - 访客签退");
    println!("  GET    /visitors/host/:host   - 获取被访人的访客列表");
    println!("  POST   /passcode/verify       - 验证通行码");

    axum::serve(listener, app)
        .await
        .expect("Failed to start server");
}
