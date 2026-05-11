use std::env;
use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use tower_http::cors::{Any, CorsLayer};
use uuid::Uuid;

use rental_core::{
    CheckoutRequest, CreateContractRequest, CreatePropertyRequest, CreateTenantRequest,
    InMemoryStorage, RenewContractRequest, RentalError, RentalService,
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long)]
    port: Option<u16>,
}

#[derive(Clone)]
struct AppState {
    service: RentalService<InMemoryStorage>,
}

#[derive(Debug, Serialize, Deserialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

impl<T> ApiResponse<T> {
    fn success(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            error: None,
        }
    }

    fn error(message: String) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(message),
        }
    }
}

#[derive(Debug, Deserialize)]
struct ExpiringQuery {
    days: Option<i64>,
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    
    let port = args
        .port
        .or_else(|| {
            env::var("PORT")
                .ok()
                .and_then(|p| p.parse::<u16>().ok())
        })
        .unwrap_or(3000);

    let storage = InMemoryStorage::new();
    let service = RentalService::new(storage);
    
    let app_state = AppState { service };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/properties", get(list_properties).post(create_property))
        .route("/properties/:id", get(get_property))
        .route("/tenants", get(list_tenants).post(create_tenant))
        .route("/tenants/:id", get(get_tenant))
        .route("/contracts", get(list_contracts).post(create_contract))
        .route("/contracts/:id", get(get_contract))
        .route("/contracts/expiring", get(get_expiring_contracts))
        .route("/contracts/renew", post(renew_contract))
        .route("/contracts/checkout", post(checkout))
        .route("/checkout/:contract_id", get(get_checkout_record))
        .with_state(Arc::new(app_state))
        .layer(cors);

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    println!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

fn map_error(err: RentalError) -> (StatusCode, Json<ApiResponse<()>>) {
    let status = match &err {
        RentalError::PropertyNotFound(_) => StatusCode::NOT_FOUND,
        RentalError::TenantNotFound(_) => StatusCode::NOT_FOUND,
        RentalError::ContractNotFound(_) => StatusCode::NOT_FOUND,
        RentalError::CheckoutRecordNotFound(_) => StatusCode::NOT_FOUND,
        RentalError::PropertyNoExists(_) => StatusCode::CONFLICT,
        RentalError::TenantPhoneExists(_) => StatusCode::CONFLICT,
        RentalError::InvalidDate(_) => StatusCode::BAD_REQUEST,
        RentalError::ContractNotActive(_) => StatusCode::BAD_REQUEST,
        RentalError::PropertyOccupied => StatusCode::CONFLICT,
        RentalError::InvalidDamageCost => StatusCode::BAD_REQUEST,
        RentalError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
    };
    (status, Json(ApiResponse::error(err.to_string())))
}

async fn list_properties(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let properties = state.service.list_properties();
    Json(ApiResponse::success(properties))
}

async fn get_property(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_property(&id) {
        Ok(property) => Json(ApiResponse::success(property)).into_response(),
        Err(e) => {
            let (status, response) = map_error(e);
            (status, response).into_response()
        }
    }
}

async fn create_property(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreatePropertyRequest>,
) -> impl IntoResponse {
    match state.service.create_property(req) {
        Ok(property) => (StatusCode::CREATED, Json(ApiResponse::success(property))).into_response(),
        Err(e) => {
            let (status, response) = map_error(e);
            (status, response).into_response()
        }
    }
}

async fn list_tenants(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let tenants = state.service.list_tenants();
    Json(ApiResponse::success(tenants))
}

async fn get_tenant(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_tenant(&id) {
        Ok(tenant) => Json(ApiResponse::success(tenant)).into_response(),
        Err(e) => {
            let (status, response) = map_error(e);
            (status, response).into_response()
        }
    }
}

async fn create_tenant(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateTenantRequest>,
) -> impl IntoResponse {
    match state.service.create_tenant(req) {
        Ok(tenant) => (StatusCode::CREATED, Json(ApiResponse::success(tenant))).into_response(),
        Err(e) => {
            let (status, response) = map_error(e);
            (status, response).into_response()
        }
    }
}

async fn list_contracts(
    State(state): State<Arc<AppState>>,
) -> impl IntoResponse {
    let contracts = state.service.list_contracts();
    Json(ApiResponse::success(contracts))
}

async fn get_contract(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_contract(&id) {
        Ok(contract) => Json(ApiResponse::success(contract)).into_response(),
        Err(e) => {
            let (status, response) = map_error(e);
            (status, response).into_response()
        }
    }
}

async fn get_expiring_contracts(
    State(state): State<Arc<AppState>>,
    Query(params): Query<ExpiringQuery>,
) -> impl IntoResponse {
    let days = params.days.unwrap_or(30);
    let contracts = state.service.get_expiring_contracts(days);
    Json(ApiResponse::success(contracts))
}

async fn create_contract(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateContractRequest>,
) -> impl IntoResponse {
    match state.service.create_contract(req) {
        Ok(contract) => (StatusCode::CREATED, Json(ApiResponse::success(contract))).into_response(),
        Err(e) => {
            let (status, response) = map_error(e);
            (status, response).into_response()
        }
    }
}

async fn renew_contract(
    State(state): State<Arc<AppState>>,
    Json(req): Json<RenewContractRequest>,
) -> impl IntoResponse {
    match state.service.renew_contract(req) {
        Ok(contract) => (StatusCode::CREATED, Json(ApiResponse::success(contract))).into_response(),
        Err(e) => {
            let (status, response) = map_error(e);
            (status, response).into_response()
        }
    }
}

async fn checkout(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CheckoutRequest>,
) -> impl IntoResponse {
    match state.service.checkout(req) {
        Ok(record) => (StatusCode::CREATED, Json(ApiResponse::success(record))).into_response(),
        Err(e) => {
            let (status, response) = map_error(e);
            (status, response).into_response()
        }
    }
}

async fn get_checkout_record(
    State(state): State<Arc<AppState>>,
    Path(contract_id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_checkout_record(&contract_id) {
        Ok(record) => Json(ApiResponse::success(record)).into_response(),
        Err(e) => {
            let (status, response) = map_error(e);
            (status, response).into_response()
        }
    }
}
