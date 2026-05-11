use crate::state::AppState;
use axum::{
    Router,
    routing::{get, post, put},
};
use std::sync::Arc;

use crate::handlers::*;

pub fn create_router() -> Router<Arc<AppState>> {
    Router::new()
        .route("/products", post(create_product))
        .route("/products", get(list_products))
        .route("/products/:id", get(get_product))
        .route("/activities", post(create_activity))
        .route("/activities", get(list_activities))
        .route("/activities/:id", get(get_activity))
        .route("/activities/:id/start", put(start_activity))
        .route("/activities/:id/end", put(end_activity))
        .route("/activities/:id/cancel", put(cancel_activity))
        .route("/orders/deposit", post(pay_deposit))
        .route("/orders/:id/final", post(pay_final))
        .route("/orders/:id", get(get_order))
        .route("/users/:user_id/orders", get(list_user_orders))
        .route("/health", get(health_check))
}
