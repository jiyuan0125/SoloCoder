use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use serde::Deserialize;
use serde_json::json;
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

use bargain_core::*;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "BARGAIN_PORT", default_value_t = 8080)]
    port: u16,
    
    #[arg(short, long, env = "BARGAIN_HOST", default_value = "0.0.0.0")]
    host: String,
}

struct AppState {
    service: BargainService,
}

#[derive(Deserialize)]
struct CreateProductBody {
    name: String,
    original_price: f64,
    floor_price: f64,
}

#[derive(Deserialize)]
struct CreateUserBody {
    name: String,
}

#[derive(Deserialize)]
struct StartBargainBody {
    product_id: Uuid,
    user_id: Uuid,
}

#[derive(Deserialize)]
struct HelpBargainBody {
    activity_id: Uuid,
    user_id: Uuid,
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let repo = InMemoryRepository::new();
    let service = BargainService::new(repo);
    let state = Arc::new(AppState { service });
    
    let cors = CorsLayer::new()
        .allow_methods(Any)
        .allow_origin(Any)
        .allow_headers(Any);
    
    let app = Router::new()
        .route("/api/products", get(list_products).post(create_product))
        .route("/api/products/:id", get(get_product))
        .route("/api/users", post(create_user))
        .route("/api/users/:id", get(get_user))
        .route("/api/bargain/start", post(start_bargain))
        .route("/api/bargain/help", post(help_bargain))
        .route("/api/bargain/:id", get(get_activity))
        .route("/api/bargain/:id/purchase", post(purchase))
        .route("/api/bargain/:id/abandon", post(abandon))
        .with_state(state)
        .layer(cors);
    
    let addr = format!("{}:{}", args.host, args.port);
    println!("Bargain Server listening on {}", addr);
    
    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

fn handle_error(e: BargainError) -> (StatusCode, impl IntoResponse) {
    let status = match e {
        BargainError::ProductNotFound => StatusCode::NOT_FOUND,
        BargainError::ActivityNotFound => StatusCode::NOT_FOUND,
        BargainError::DuplicateActivity => StatusCode::CONFLICT,
        BargainError::AlreadyBargained => StatusCode::CONFLICT,
        BargainError::ActivityExpired => StatusCode::GONE,
        BargainError::ActivityNotActive => StatusCode::BAD_REQUEST,
        BargainError::InvalidPrice => StatusCode::BAD_REQUEST,
        BargainError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
    };
    (status, Json(json!({ "error": e.to_string() })))
}

async fn list_products(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let products = state.service.list_products();
    Json(products)
}

async fn create_product(
    State(state): State<Arc<AppState>>,
    Json(body): Json<CreateProductBody>,
) -> impl IntoResponse {
    match state.service.create_product(CreateProductRequest {
        name: body.name,
        original_price: body.original_price,
        floor_price: body.floor_price,
    }) {
        Ok(product) => (StatusCode::CREATED, Json(product)).into_response(),
        Err(e) => {
            let (status, body) = handle_error(e);
            (status, body).into_response()
        }
    }
}

async fn get_product(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_product(id) {
        Ok(product) => Json(product).into_response(),
        Err(e) => {
            let (status, body) = handle_error(e);
            (status, body).into_response()
        }
    }
}

async fn create_user(
    State(state): State<Arc<AppState>>,
    Json(body): Json<CreateUserBody>,
) -> impl IntoResponse {
    let user = state.service.create_user(body.name);
    (StatusCode::CREATED, Json(user))
}

async fn get_user(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_user(id) {
        Some(user) => Json(user).into_response(),
        None => (StatusCode::NOT_FOUND, Json(json!({ "error": "User not found" }))).into_response(),
    }
}

async fn start_bargain(
    State(state): State<Arc<AppState>>,
    Json(body): Json<StartBargainBody>,
) -> impl IntoResponse {
    match state.service.start_bargain(StartBargainRequest {
        product_id: body.product_id,
        user_id: body.user_id,
    }) {
        Ok(detail) => (StatusCode::CREATED, Json(detail)).into_response(),
        Err(e) => {
            let (status, body) = handle_error(e);
            (status, body).into_response()
        }
    }
}

async fn help_bargain(
    State(state): State<Arc<AppState>>,
    Json(body): Json<HelpBargainBody>,
) -> impl IntoResponse {
    match state.service.help_bargain(BargainRequest {
        activity_id: body.activity_id,
        user_id: body.user_id,
    }) {
        Ok(detail) => Json(detail).into_response(),
        Err(e) => {
            let (status, body) = handle_error(e);
            (status, body).into_response()
        }
    }
}

async fn get_activity(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_activity_detail(id) {
        Ok(detail) => Json(detail).into_response(),
        Err(e) => {
            let (status, body) = handle_error(e);
            (status, body).into_response()
        }
    }
}

async fn purchase(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.purchase(id) {
        Ok(detail) => Json(detail).into_response(),
        Err(e) => {
            let (status, body) = handle_error(e);
            (status, body).into_response()
        }
    }
}

async fn abandon(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.abandon(id) {
        Ok(detail) => Json(detail).into_response(),
        Err(e) => {
            let (status, body) = handle_error(e);
            (status, body).into_response()
        }
    }
}
