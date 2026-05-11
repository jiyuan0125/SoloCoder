use clap::Parser;
use std::net::SocketAddr;
use std::sync::Arc;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{get, post, put},
    Router,
};
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use insurance_claim_core::*;

#[derive(Parser, Debug)]
#[command(name = "insurance-claim-server")]
struct Cli {
    #[arg(short, long, env = "PORT", default_value = "3000")]
    port: u16,

    #[arg(short, long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    service: Arc<ClaimService>,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

struct AppError(ClaimError);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let status = match &self.0 {
            ClaimError::ClaimClosed => StatusCode::BAD_REQUEST,
            ClaimError::ClaimNotFound => StatusCode::NOT_FOUND,
            ClaimError::PolicyNotFound => StatusCode::NOT_FOUND,
            ClaimError::InvalidLiabilityRatio => StatusCode::BAD_REQUEST,
            ClaimError::InvalidLiabilityRatioSum(_) => StatusCode::BAD_REQUEST,
            ClaimError::InvalidAmount => StatusCode::BAD_REQUEST,
            ClaimError::DuplicateParty => StatusCode::BAD_REQUEST,
        };
        
        (status, Json(ErrorResponse { error: self.0.to_string() })).into_response()
    }
}

impl From<ClaimError> for AppError {
    fn from(err: ClaimError) -> Self {
        AppError(err)
    }
}

type AppResult<T> = Result<Json<T>, AppError>;

#[derive(Debug, Deserialize)]
struct UpdateLossRequest {
    total_loss: Decimal,
}

async fn create_policy(
    State(state): State<AppState>,
    Json(req): Json<CreatePolicyRequest>,
) -> AppResult<Policy> {
    let policy = state.service.create_policy(req).await?;
    Ok(Json(policy))
}

async fn get_policies(State(state): State<AppState>) -> Json<Vec<Policy>> {
    let policies = state.service.get_all_policies().await;
    Json(policies)
}

async fn get_policy(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> AppResult<Policy> {
    let policy = state.service.get_policy(&id).await?;
    Ok(Json(policy))
}

async fn create_claim(
    State(state): State<AppState>,
    Json(req): Json<CreateClaimRequest>,
) -> AppResult<Claim> {
    let claim = state.service.create_claim(req).await?;
    Ok(Json(claim))
}

async fn get_claims(State(state): State<AppState>) -> Json<Vec<Claim>> {
    let claims = state.service.get_all_claims().await;
    Json(claims)
}

async fn get_claim(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> AppResult<Claim> {
    let claim = state.service.get_claim(&id).await?;
    Ok(Json(claim))
}

async fn close_claim(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> AppResult<Claim> {
    let claim = state.service.close_claim(&id).await?;
    Ok(Json(claim))
}

async fn update_claim_parties(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(parties): Json<Vec<CreatePartyRequest>>,
) -> AppResult<Claim> {
    let claim = state.service.update_claim_parties(&id, parties).await?;
    Ok(Json(claim))
}

async fn update_claim_loss(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<UpdateLossRequest>,
) -> AppResult<Claim> {
    let claim = state.service.update_claim_total_loss(&id, req.total_loss).await?;
    Ok(Json(claim))
}

fn router(state: AppState) -> Router {
    Router::new()
        .route("/policies", post(create_policy).get(get_policies))
        .route("/policies/:id", get(get_policy))
        .route("/claims", post(create_claim).get(get_claims))
        .route("/claims/:id", get(get_claim).post(close_claim))
        .route("/claims/:id/parties", put(update_claim_parties))
        .route("/claims/:id/loss", put(update_claim_loss))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    let cli = Cli::parse();
    
    let storage = InMemoryStorage::new();
    let service = ClaimService::new(storage);
    
    let state = AppState {
        service: Arc::new(service),
    };
    
    let app = router(state);
    
    let addr: SocketAddr = format!("{}:{}", cli.host, cli.port)
        .parse()
        .expect("Invalid address");
    
    println!("Insurance Claim Server starting on {}", addr);
    println!("Port configured via --port or PORT environment variable");
    
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
