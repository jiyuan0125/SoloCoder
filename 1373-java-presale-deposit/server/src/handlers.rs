use crate::middleware::map_error_to_response;
use crate::state::AppState;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    Json,
};
use chrono::{DateTime, Utc};
use presale_core::{
    models::{ActivitySummary, OrderDetail, PresaleConfig},
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use uuid::Uuid;

#[derive(Debug, Deserialize)]
pub struct CreateProductRequest {
    pub name: String,
    pub original_price: u64,
}

#[derive(Debug, Deserialize)]
pub struct CreateActivityRequest {
    pub product_id: Uuid,
    pub deposit_amount: u64,
    pub inflation_rate: u32,
    pub max_participants: u32,
    pub start_time: DateTime<Utc>,
    pub end_time: DateTime<Utc>,
    pub final_payment_deadline_hours: Option<i64>,
}

#[derive(Debug, Deserialize)]
pub struct PayDepositRequest {
    pub user_id: String,
    pub activity_id: Uuid,
}

#[derive(Debug, Deserialize)]
pub struct PayFinalRequest {
    pub paid_amount: u64,
}

#[derive(Debug, Serialize)]
pub struct SimpleResponse {
    pub success: bool,
}

#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
}

pub async fn health_check() -> impl IntoResponse {
    (StatusCode::OK, Json(HealthResponse { status: "ok".to_string() }))
}

pub async fn create_product(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateProductRequest>,
) -> impl IntoResponse {
    let product = state.service.create_product(req.name, req.original_price);
    (StatusCode::CREATED, Json(product))
}

pub async fn list_products(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let products = state.service.list_products();
    (StatusCode::OK, Json(products))
}

pub async fn get_product(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_product(id) {
        Ok(product) => (StatusCode::OK, Json(Some(product))).into_response(),
        Err(e) => map_error_to_response(e),
    }
}

pub async fn create_activity(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateActivityRequest>,
) -> impl IntoResponse {
    let config = PresaleConfig {
        deposit_amount: req.deposit_amount,
        inflation_rate: req.inflation_rate,
    };

    match state.service.create_activity(
        req.product_id,
        config,
        req.max_participants,
        req.start_time,
        req.end_time,
        req.final_payment_deadline_hours,
    ) {
        Ok(activity) => (StatusCode::CREATED, Json(ActivitySummary::from(&activity))).into_response(),
        Err(e) => map_error_to_response(e),
    }
}

pub async fn list_activities(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let activities = state.service.list_activities();
    (StatusCode::OK, Json(activities))
}

pub async fn get_activity(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_activity(id) {
        Ok(summary) => (StatusCode::OK, Json(summary)).into_response(),
        Err(e) => map_error_to_response(e),
    }
}

pub async fn start_activity(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.start_activity(id) {
        Ok(_) => (StatusCode::OK, Json(SimpleResponse { success: true })).into_response(),
        Err(e) => map_error_to_response(e),
    }
}

pub async fn end_activity(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.end_activity(id) {
        Ok(_) => (StatusCode::OK, Json(SimpleResponse { success: true })).into_response(),
        Err(e) => map_error_to_response(e),
    }
}

pub async fn cancel_activity(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    let now = Utc::now();
    match state.service.cancel_activity(id, now) {
        Ok(_) => (StatusCode::OK, Json(SimpleResponse { success: true })).into_response(),
        Err(e) => map_error_to_response(e),
    }
}

pub async fn pay_deposit(
    State(state): State<Arc<AppState>>,
    Json(req): Json<PayDepositRequest>,
) -> impl IntoResponse {
    let now = Utc::now();
    match state.service.pay_deposit(req.user_id, req.activity_id, now) {
        Ok(order) => (StatusCode::CREATED, Json(OrderDetail::from(&order))).into_response(),
        Err(e) => map_error_to_response(e),
    }
}

pub async fn pay_final(
    State(state): State<Arc<AppState>>,
    Path(order_id): Path<Uuid>,
    Json(req): Json<PayFinalRequest>,
) -> impl IntoResponse {
    let now = Utc::now();
    match state.service.pay_final(order_id, req.paid_amount, now) {
        Ok(order) => (StatusCode::OK, Json(OrderDetail::from(&order))).into_response(),
        Err(e) => map_error_to_response(e),
    }
}

pub async fn get_order(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> impl IntoResponse {
    match state.service.get_order(id) {
        Ok(detail) => (StatusCode::OK, Json(detail)).into_response(),
        Err(e) => map_error_to_response(e),
    }
}

pub async fn list_user_orders(
    State(state): State<Arc<AppState>>,
    Path(user_id): Path<String>,
) -> impl IntoResponse {
    let orders = state.service.list_orders_by_user(&user_id);
    (StatusCode::OK, Json(orders))
}
