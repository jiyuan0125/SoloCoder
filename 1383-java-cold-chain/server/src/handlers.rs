use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::IntoResponse,
    Json,
};
use cold_chain_core::{ColdChainError, ColdChainService};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use uuid::Uuid;

#[derive(Debug, Deserialize)]
pub struct CreateShipmentRequest {
    pub cargo_type: String,
    pub vehicle_id: String,
    pub min_temp: f64,
    pub max_temp: f64,
}

#[derive(Debug, Deserialize)]
pub struct ReportTemperatureRequest {
    pub temperature: f64,
}

#[derive(Debug, Deserialize)]
pub struct ListAlertsQuery {
    pub shipment_id: Option<String>,
}

#[derive(Debug, Serialize)]
pub struct ErrorResponse {
    pub error: String,
}

pub async fn create_shipment(
    State(service): State<Arc<ColdChainService>>,
    Json(req): Json<CreateShipmentRequest>,
) -> impl IntoResponse {
    let shipment = service.create_shipment(
        req.cargo_type,
        req.vehicle_id,
        req.min_temp,
        req.max_temp,
    );
    (StatusCode::CREATED, Json(shipment))
}

pub async fn list_shipments(State(service): State<Arc<ColdChainService>>) -> impl IntoResponse {
    let shipments = service.list_shipments();
    Json(shipments)
}

pub async fn get_shipment(
    State(service): State<Arc<ColdChainService>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match service.get_shipment(&uuid) {
            Ok(shipment) => (StatusCode::OK, Json(shipment)).into_response(),
            Err(e) => handle_error(e),
        },
        Err(_) => (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Invalid UUID format".to_string(),
            }),
        )
            .into_response(),
    }
}

pub async fn start_shipment(
    State(service): State<Arc<ColdChainService>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match service.start_shipment(&uuid) {
            Ok(shipment) => (StatusCode::OK, Json(shipment)).into_response(),
            Err(e) => handle_error(e),
        },
        Err(_) => (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Invalid UUID format".to_string(),
            }),
        )
            .into_response(),
    }
}

pub async fn report_temperature(
    State(service): State<Arc<ColdChainService>>,
    Path(id): Path<String>,
    Json(req): Json<ReportTemperatureRequest>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match service.report_temperature(&uuid, req.temperature) {
            Ok(reading) => (StatusCode::OK, Json(reading)).into_response(),
            Err(e) => handle_error(e),
        },
        Err(_) => (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Invalid UUID format".to_string(),
            }),
        )
            .into_response(),
    }
}

pub async fn complete_shipment(
    State(service): State<Arc<ColdChainService>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match service.complete_shipment(&uuid) {
            Ok(shipment) => (StatusCode::OK, Json(shipment)).into_response(),
            Err(e) => handle_error(e),
        },
        Err(_) => (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Invalid UUID format".to_string(),
            }),
        )
            .into_response(),
    }
}

pub async fn accept_shipment(
    State(service): State<Arc<ColdChainService>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match service.accept_shipment(&uuid) {
            Ok(shipment) => (StatusCode::OK, Json(shipment)).into_response(),
            Err(e) => handle_error(e),
        },
        Err(_) => (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Invalid UUID format".to_string(),
            }),
        )
            .into_response(),
    }
}

pub async fn reject_shipment(
    State(service): State<Arc<ColdChainService>>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    match Uuid::parse_str(&id) {
        Ok(uuid) => match service.reject_shipment(&uuid) {
            Ok(shipment) => (StatusCode::OK, Json(shipment)).into_response(),
            Err(e) => handle_error(e),
        },
        Err(_) => (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: "Invalid UUID format".to_string(),
            }),
        )
            .into_response(),
    }
}

pub async fn list_alerts(
    State(service): State<Arc<ColdChainService>>,
    Query(query): Query<ListAlertsQuery>,
) -> impl IntoResponse {
    let shipment_uuid = match query.shipment_id {
        Some(id) => match Uuid::parse_str(&id) {
            Ok(uuid) => Some(uuid),
            Err(_) => {
                return (
                    StatusCode::BAD_REQUEST,
                    Json(ErrorResponse {
                        error: "Invalid UUID format".to_string(),
                    }),
                )
                    .into_response()
            }
        },
        None => None,
    };

    let alerts = service.list_alerts(shipment_uuid.as_ref());
    Json(alerts).into_response()
}

pub async fn get_vehicle_stats(
    State(service): State<Arc<ColdChainService>>,
    Path(params): Path<HashMap<String, String>>,
) -> impl IntoResponse {
    let vehicle_id = match params.get("vehicle_id") {
        Some(id) => id.clone(),
        None => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    error: "Missing vehicle_id".to_string(),
                }),
            )
                .into_response()
        }
    };

    let year = match params
        .get("year")
        .and_then(|y| y.parse::<i32>().ok())
    {
        Some(y) => y,
        None => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    error: "Invalid or missing year".to_string(),
                }),
            )
                .into_response()
        }
    };

    let month = match params
        .get("month")
        .and_then(|m| m.parse::<u32>().ok())
    {
        Some(m) if (1..=12).contains(&m) => m,
        _ => {
            return (
                StatusCode::BAD_REQUEST,
                Json(ErrorResponse {
                    error: "Invalid or missing month (1-12)".to_string(),
                }),
            )
                .into_response()
        }
    };

    let stats = service.get_vehicle_monthly_stats(&vehicle_id, year, month);
    Json(stats).into_response()
}

pub async fn get_cargo_stats(
    State(service): State<Arc<ColdChainService>>,
    Path(cargo_type): Path<String>,
) -> impl IntoResponse {
    let stats = service.get_cargo_type_stats(&cargo_type);
    Json(stats)
}

fn handle_error(error: ColdChainError) -> axum::response::Response {
    match error {
        ColdChainError::ShipmentNotFound(_) => (
            StatusCode::NOT_FOUND,
            Json(ErrorResponse {
                error: error.to_string(),
            }),
        ),
        ColdChainError::InvalidStateTransition(_) => (
            StatusCode::CONFLICT,
            Json(ErrorResponse {
                error: error.to_string(),
            }),
        ),
        ColdChainError::ShipmentNotInTransit => (
            StatusCode::BAD_REQUEST,
            Json(ErrorResponse {
                error: error.to_string(),
            }),
        ),
        ColdChainError::BrokenChainCannotAccept => (
            StatusCode::CONFLICT,
            Json(ErrorResponse {
                error: error.to_string(),
            }),
        ),
        ColdChainError::Internal(_) => (
            StatusCode::INTERNAL_SERVER_ERROR,
            Json(ErrorResponse {
                error: error.to_string(),
            }),
        ),
    }
    .into_response()
}
