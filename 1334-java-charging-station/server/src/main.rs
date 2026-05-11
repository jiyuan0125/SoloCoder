use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post},
    Router,
};
use clap::Parser;
use serde::Deserialize;
use std::net::SocketAddr;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};
use uuid::Uuid;

use charging_core::{
    errors::ChargingError,
    models::{Charger, CreateChargerRequest, CreateOrderRequest, EndOrderRequest, Order},
    services::{create_charger, create_order, end_order},
    storage::{create_shared_state, get_all_chargers, get_all_orders, get_charger, get_order, initialize_demo_data, SharedState},
};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "CHARGING_PORT", default_value_t = 3000)]
    port: u16,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "charging_server=debug,tower_http=debug,axum::rejection=trace".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();
    let shared_state = create_shared_state();
    
    initialize_demo_data(shared_state.clone()).await;

    let app = Router::new()
        .route("/health", get(health_check))
        .route("/chargers", get(list_chargers))
        .route("/chargers", post(add_charger))
        .route("/chargers/:id", get(get_charger_handler))
        .route("/orders", get(list_orders))
        .route("/orders", post(start_charging))
        .route("/orders/:id", get(get_order_handler))
        .route("/orders/:id/end", post(end_charging))
        .with_state(shared_state);

    let addr = SocketAddr::from(([127, 0, 0, 1], args.port));
    tracing::debug!("listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

async fn health_check() -> &'static str {
    "OK"
}

async fn list_chargers(State(state): State<SharedState>) -> Json<Vec<Charger>> {
    let chargers = get_all_chargers(&state).await;
    let mut result: Vec<Charger> = chargers.into_values().collect();
    result.sort_by(|a, b| a.name.cmp(&b.name));
    Json(result)
}

async fn get_charger_handler(
    State(state): State<SharedState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Charger>, AppError> {
    match get_charger(&state, id).await {
        Some(charger) => Ok(Json(charger)),
        None => Err(ChargingError::ChargerNotFound(id.to_string()).into()),
    }
}

async fn add_charger(
    State(state): State<SharedState>,
    Json(payload): Json<CreateChargerRequest>,
) -> Json<Charger> {
    let charger = create_charger(&state, payload.name, payload.power_kw).await;
    Json(charger)
}

async fn list_orders(State(state): State<SharedState>) -> Json<Vec<Order>> {
    let orders = get_all_orders(&state).await;
    let mut result: Vec<Order> = orders.into_values().collect();
    result.sort_by(|a, b| b.start_time.cmp(&a.start_time));
    Json(result)
}

async fn get_order_handler(
    State(state): State<SharedState>,
    Path(id): Path<Uuid>,
) -> Result<Json<Order>, AppError> {
    match get_order(&state, id).await {
        Some(order) => Ok(Json(order)),
        None => Err(ChargingError::OrderNotFound(id.to_string()).into()),
    }
}

async fn start_charging(
    State(state): State<SharedState>,
    Json(payload): Json<CreateOrderRequest>,
) -> Result<Json<Order>, AppError> {
    let order = create_order(&state, payload.charger_id, payload.user_id).await?;
    Ok(Json(order))
}

async fn end_charging(
    State(state): State<SharedState>,
    Path(order_id): Path<Uuid>,
) -> Result<Json<Order>, AppError> {
    let order = end_order(&state, order_id).await?;
    Ok(Json(order))
}

struct AppError(anyhow::Error);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let (status, error_message) = match self.0.downcast_ref::<ChargingError>() {
            Some(ChargingError::ChargerNotFound(_)) => (StatusCode::NOT_FOUND, self.0.to_string()),
            Some(ChargingError::ChargerBusy(_)) => (StatusCode::CONFLICT, self.0.to_string()),
            Some(ChargingError::OrderNotFound(_)) => (StatusCode::NOT_FOUND, self.0.to_string()),
            Some(ChargingError::OrderAlreadyCompleted(_)) => (StatusCode::BAD_REQUEST, self.0.to_string()),
            Some(ChargingError::InvalidRequest(_)) => (StatusCode::BAD_REQUEST, self.0.to_string()),
            None => (StatusCode::INTERNAL_SERVER_ERROR, "Internal server error".to_string()),
        };

        (status, Json(serde_json::json!({ "error": error_message }))).into_response()
    }
}

impl<E> From<E> for AppError
where
    E: Into<anyhow::Error>,
{
    fn from(err: E) -> Self {
        Self(err.into())
    }
}
