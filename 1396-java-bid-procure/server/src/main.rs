use std::env;
use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use bid_procure_core::{
    AcceptInvitationRequest, CreateProcurementRequest, InviteSuppliersRequest, ProcurementError,
    ProcurementService, QualificationLevel, SubmitBidRequest, WithdrawBidRequest,
};
use serde::Deserialize;
use tower_http::cors::{Any, CorsLayer};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Clone)]
struct AppState {
    service: Arc<ProcurementService>,
}

#[derive(Debug, Deserialize)]
struct CreateSupplierRequest {
    name: String,
    qualification: String,
}

#[derive(Debug, Deserialize)]
struct UpdateProcurementRequest {
    title: Option<String>,
    description: Option<String>,
    required_qualification: Option<QualificationLevel>,
}

struct AppError(ProcurementError);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let (status, message) = match self.0 {
            ProcurementError::ProcurementNotFound(_) => {
                (StatusCode::NOT_FOUND, self.0.to_string())
            }
            ProcurementError::SupplierNotFound(_) => {
                (StatusCode::NOT_FOUND, self.0.to_string())
            }
            ProcurementError::InvalidStatusTransition { .. } => {
                (StatusCode::BAD_REQUEST, self.0.to_string())
            }
            ProcurementError::InsufficientQualification { .. } => {
                (StatusCode::FORBIDDEN, self.0.to_string())
            }
            ProcurementError::NotInvited => (StatusCode::FORBIDDEN, self.0.to_string()),
            ProcurementError::SupplierWithdrawn => (StatusCode::FORBIDDEN, self.0.to_string()),
            ProcurementError::InvalidPrice => (StatusCode::BAD_REQUEST, self.0.to_string()),
            ProcurementError::InvalidDeliveryDate => (StatusCode::BAD_REQUEST, self.0.to_string()),
            ProcurementError::AlreadySubmitted => (StatusCode::BAD_REQUEST, self.0.to_string()),
            ProcurementError::NotInBiddingPhase => (StatusCode::BAD_REQUEST, self.0.to_string()),
            ProcurementError::PriceNotDecreased => (StatusCode::BAD_REQUEST, self.0.to_string()),
            ProcurementError::CannotModifyDuringBidding => {
                (StatusCode::BAD_REQUEST, self.0.to_string())
            }
            ProcurementError::InsufficientSuppliers => {
                (StatusCode::CONFLICT, self.0.to_string())
            }
            ProcurementError::RoundEnded => (StatusCode::BAD_REQUEST, self.0.to_string()),
            ProcurementError::NoValidBids => (StatusCode::CONFLICT, self.0.to_string()),
        };
        (status, Json(serde_json::json!({ "error": message }))).into_response()
    }
}

impl<E> From<E> for AppError
where
    E: Into<ProcurementError>,
{
    fn from(err: E) -> Self {
        AppError(err.into())
    }
}

async fn create_supplier(
    State(state): State<AppState>,
    Json(req): Json<CreateSupplierRequest>,
) -> Result<impl IntoResponse, AppError> {
    let qualification = req
        .qualification
        .parse::<QualificationLevel>()
        .map_err(|e| ProcurementError::SupplierNotFound(e))?;
    let supplier = state.service.create_supplier(req.name, qualification);
    Ok((StatusCode::CREATED, Json(supplier)))
}

async fn list_suppliers(State(state): State<AppState>) -> impl IntoResponse {
    let suppliers = state.service.list_suppliers();
    Json(suppliers)
}

async fn get_supplier(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, AppError> {
    let id = uuid::Uuid::parse_str(&id).map_err(|_| {
        ProcurementError::SupplierNotFound("无效的供应商ID".to_string())
    })?;
    let supplier = state
        .service
        .get_supplier(id)
        .ok_or_else(|| ProcurementError::SupplierNotFound(id.to_string()))?;
    Ok(Json(supplier))
}

async fn create_procurement(
    State(state): State<AppState>,
    Json(req): Json<CreateProcurementRequest>,
) -> Result<impl IntoResponse, AppError> {
    let procurement = state.service.create_procurement(
        req.title,
        req.description,
        req.required_qualification,
    );
    Ok((StatusCode::CREATED, Json(procurement)))
}

async fn list_procurements(State(state): State<AppState>) -> impl IntoResponse {
    let procurements = state.service.list_procurements();
    Json(procurements)
}

async fn get_procurement(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, AppError> {
    let id = uuid::Uuid::parse_str(&id).map_err(|_| {
        ProcurementError::ProcurementNotFound("无效的采购需求ID".to_string())
    })?;
    let procurement = state.service.get_procurement(id)?;
    Ok(Json(procurement))
}

async fn update_procurement(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<UpdateProcurementRequest>,
) -> Result<impl IntoResponse, AppError> {
    let id = uuid::Uuid::parse_str(&id).map_err(|_| {
        ProcurementError::ProcurementNotFound("无效的采购需求ID".to_string())
    })?;
    let procurement = state.service.update_procurement(
        id,
        req.title,
        req.description,
        req.required_qualification,
    )?;
    Ok(Json(procurement))
}

async fn invite_suppliers(
    State(state): State<AppState>,
    Json(req): Json<InviteSuppliersRequest>,
) -> Result<impl IntoResponse, AppError> {
    let procurement = state
        .service
        .invite_suppliers(req.procurement_id, req.supplier_ids)?;
    Ok(Json(procurement))
}

async fn accept_invitation(
    State(state): State<AppState>,
    Json(req): Json<AcceptInvitationRequest>,
) -> Result<impl IntoResponse, AppError> {
    let procurement = state
        .service
        .accept_invitation(req.procurement_id, req.supplier_id)?;
    Ok(Json(procurement))
}

async fn start_bidding(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, AppError> {
    let id = uuid::Uuid::parse_str(&id).map_err(|_| {
        ProcurementError::ProcurementNotFound("无效的采购需求ID".to_string())
    })?;
    let procurement = state.service.start_bidding(id)?;
    Ok(Json(procurement))
}

async fn submit_bid(
    State(state): State<AppState>,
    Json(req): Json<SubmitBidRequest>,
) -> Result<impl IntoResponse, AppError> {
    let procurement = state.service.submit_bid(
        req.procurement_id,
        req.supplier_id,
        req.price,
        req.delivery_date,
    )?;
    Ok(Json(procurement))
}

async fn withdraw_bid(
    State(state): State<AppState>,
    Json(req): Json<WithdrawBidRequest>,
) -> Result<impl IntoResponse, AppError> {
    let procurement = state
        .service
        .withdraw_from_bidding(req.procurement_id, req.supplier_id)?;
    Ok(Json(procurement))
}

async fn get_round_result(
    State(state): State<AppState>,
    Path((procurement_id, round)): Path<(String, u32)>,
) -> Result<impl IntoResponse, AppError> {
    let procurement_id = uuid::Uuid::parse_str(&procurement_id).map_err(|_| {
        ProcurementError::ProcurementNotFound("无效的采购需求ID".to_string())
    })?;
    let result = state
        .service
        .get_round_result(procurement_id, round)
        .ok_or_else(|| ProcurementError::RoundEnded)?;
    Ok(Json(result))
}

async fn advance_round(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, AppError> {
    let id = uuid::Uuid::parse_str(&id).map_err(|_| {
        ProcurementError::ProcurementNotFound("无效的采购需求ID".to_string())
    })?;
    let procurement = state.service.advance_round(id)?;
    Ok(Json(procurement))
}

async fn evaluate_winner(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> Result<impl IntoResponse, AppError> {
    let id = uuid::Uuid::parse_str(&id).map_err(|_| {
        ProcurementError::ProcurementNotFound("无效的采购需求ID".to_string())
    })?;
    let procurement = state.service.evaluate_winner(id)?;
    Ok(Json(procurement))
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "server=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let service = ProcurementService::new();
    let state = AppState { service };

    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    let app = Router::new()
        .route("/suppliers", post(create_supplier).get(list_suppliers))
        .route("/suppliers/:id", get(get_supplier))
        .route(
            "/procurements",
            post(create_procurement).get(list_procurements),
        )
        .route(
            "/procurements/:id",
            get(get_procurement).post(update_procurement),
        )
        .route("/procurements/invite", post(invite_suppliers))
        .route("/procurements/accept", post(accept_invitation))
        .route("/procurements/:id/start", post(start_bidding))
        .route("/procurements/bid", post(submit_bid))
        .route("/procurements/withdraw", post(withdraw_bid))
        .route("/procurements/:id/round/:round", get(get_round_result))
        .route("/procurements/:id/advance", post(advance_round))
        .route("/procurements/:id/evaluate", post(evaluate_winner))
        .layer(cors)
        .with_state(state);

    let port = env::var("PORT")
        .ok()
        .and_then(|p| p.parse().ok())
        .unwrap_or(8080);

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    tracing::debug!("listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
