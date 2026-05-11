use std::net::SocketAddr;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::*,
    Json, Router,
};
use clap::Parser;
use decl_core::*;
use serde::Serialize;
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "SERVER_HOST", default_value = "0.0.0.0")]
    host: String,

    #[arg(long, env = "SERVER_PORT", default_value_t = 3000)]
    port: u16,
}

#[derive(Debug, Serialize)]
struct ErrorResponse {
    error: String,
}

struct ApiError(AppError);

impl IntoResponse for ApiError {
    fn into_response(self) -> Response {
        let status = StatusCode::from_u16(self.0.http_status()).unwrap_or(StatusCode::INTERNAL_SERVER_ERROR);
        let body = Json(ErrorResponse {
            error: self.0.to_string(),
        });
        (status, body).into_response()
    }
}

impl From<AppError> for ApiError {
    fn from(err: AppError) -> Self {
        ApiError(err)
    }
}

type AppResult<T> = Result<Json<T>, ApiError>;

async fn list_exchange_rates(State(state): State<AppState>) -> AppResult<Vec<ExchangeRate>> {
    let rates = service::list_exchange_rates(&state)?;
    Ok(Json(rates))
}

async fn add_exchange_rate(
    State(state): State<AppState>,
    Json(req): Json<CreateExchangeRateRequest>,
) -> AppResult<ExchangeRate> {
    let rate = service::add_exchange_rate(&state, req)?;
    Ok(Json(rate))
}

async fn list_declarations(State(state): State<AppState>) -> AppResult<Vec<DeclarationSummary>> {
    let decls = service::list_declarations(&state)?;
    Ok(Json(decls))
}

async fn get_declaration(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> AppResult<Declaration> {
    let decl = service::get_declaration(&state, id)?
        .ok_or_else(|| AppError::DeclarationNotFound(id.to_string()))?;
    Ok(Json(decl))
}

async fn create_declaration(
    State(state): State<AppState>,
    Json(req): Json<CreateDeclarationRequest>,
) -> AppResult<Declaration> {
    let decl = service::create_declaration(&state, req)?;
    Ok(Json(decl))
}

async fn update_declaration(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
    Json(req): Json<UpdateDeclarationRequest>,
) -> AppResult<Declaration> {
    let decl = service::update_declaration(&state, id, req)?;
    Ok(Json(decl))
}

async fn delete_declaration(
    State(state): State<AppState>,
    Path(id): Path<Uuid>,
) -> Result<StatusCode, ApiError> {
    service::delete_declaration(&state, id)?;
    Ok(StatusCode::NO_CONTENT)
}

async fn merge_declarations(
    State(state): State<AppState>,
    Json(req): Json<MergeDeclarationRequest>,
) -> AppResult<Declaration> {
    let decl = service::merge_declarations(&state, req)?;
    Ok(Json(decl))
}

async fn declare(
    State(state): State<AppState>,
    Json(req): Json<DeclareRequest>,
) -> AppResult<Declaration> {
    let decl = service::declare(&state, req)?;
    Ok(Json(decl))
}

fn router(state: AppState) -> Router {
    Router::new()
        .route("/api/rates", get(list_exchange_rates).post(add_exchange_rate))
        .route("/api/declarations", get(list_declarations).post(create_declaration))
        .route(
            "/api/declarations/:id",
            get(get_declaration).put(update_declaration).delete(delete_declaration),
        )
        .route("/api/declarations/merge", post(merge_declarations))
        .route("/api/declarations/declare", post(declare))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    let state = AppState::new();

    let app = router(state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port)
        .parse()
        .expect("Invalid address");

    println!("Server listening on http://{}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
