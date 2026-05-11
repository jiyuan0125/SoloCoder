use axum::{routing::{get, post}, Router};
use cold_chain_core::ColdChainService;
use std::sync::Arc;

pub fn create_routes(service: Arc<ColdChainService>) -> Router {
    Router::new()
        .route("/api/shipments", post(crate::handlers::create_shipment).get(crate::handlers::list_shipments))
        .route("/api/shipments/:id", get(crate::handlers::get_shipment))
        .route("/api/shipments/:id/start", post(crate::handlers::start_shipment))
        .route("/api/shipments/:id/temperature", post(crate::handlers::report_temperature))
        .route("/api/shipments/:id/complete", post(crate::handlers::complete_shipment))
        .route("/api/shipments/:id/accept", post(crate::handlers::accept_shipment))
        .route("/api/shipments/:id/reject", post(crate::handlers::reject_shipment))
        .route("/api/alerts", get(crate::handlers::list_alerts))
        .route(
            "/api/stats/vehicle/:vehicle_id/:year/:month",
            get(crate::handlers::get_vehicle_stats),
        )
        .route("/api/stats/cargo/:cargo_type", get(crate::handlers::get_cargo_stats))
        .with_state(service)
}
