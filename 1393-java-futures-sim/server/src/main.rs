use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use futures_sim_core::{
    Contract, PositionType, TradingEngine, TradingError, INITIAL_FUNDS,
};
use serde::{Deserialize, Serialize};
use std::sync::{Arc, Mutex};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(long, env = "PORT", default_value_t = 3000)]
    port: u16,
    #[arg(long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    engine: Arc<Mutex<TradingEngine>>,
}

#[derive(Serialize)]
struct ApiResponse<T: Serialize> {
    success: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    data: Option<T>,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<String>,
}

impl<T: Serialize> ApiResponse<T> {
    fn ok(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            error: None,
        }
    }
}

impl ApiResponse<()> {
    fn ok_empty() -> Self {
        Self {
            success: true,
            data: None,
            error: None,
        }
    }

    fn err(msg: String) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(msg),
        }
    }
}

struct AppError(TradingError);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::<()>::err(self.0.to_string())),
        )
            .into_response()
    }
}

impl From<TradingError> for AppError {
    fn from(err: TradingError) -> Self {
        AppError(err)
    }
}

#[derive(Deserialize)]
struct CreateContractRequest {
    code: String,
    price: f64,
    multiplier: f64,
    margin_ratio: f64,
}

#[derive(Deserialize)]
struct OpenPositionRequest {
    user_id: String,
    contract_code: String,
    position_type: String,
    lots: i32,
}

#[derive(Deserialize)]
struct ClosePositionRequest {
    user_id: String,
    contract_code: String,
    position_type: String,
    lots: i32,
}

#[derive(Deserialize)]
struct UpdatePriceRequest {
    price: f64,
}

fn parse_position_type(s: &str) -> Result<PositionType, AppError> {
    match s.to_lowercase().as_str() {
        "long" => Ok(PositionType::Long),
        "short" => Ok(PositionType::Short),
        _ => Err(TradingError::InvalidLots.into()),
    }
}

async fn list_contracts(State(state): State<AppState>) -> impl IntoResponse {
    let engine = state.engine.lock().unwrap();
    let contracts: Vec<_> = engine.get_contracts().iter().map(|c| (*c).clone()).collect();
    Json(ApiResponse::ok(contracts))
}

async fn get_contract(
    State(state): State<AppState>,
    Path(code): Path<String>,
) -> Result<impl IntoResponse, impl IntoResponse> {
    let engine = state.engine.lock().unwrap();
    match engine.get_contract(&code) {
        Some(c) => Ok(Json(ApiResponse::ok(c.clone()))),
        None => Err((
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<()>::err(format!("合约不存在: {}", code))),
        )),
    }
}

async fn create_contract(
    State(state): State<AppState>,
    Json(req): Json<CreateContractRequest>,
) -> impl IntoResponse {
    let contract = Contract::new(&req.code, req.price, req.multiplier, req.margin_ratio);
    let mut engine = state.engine.lock().unwrap();
    engine.add_contract(contract.clone());
    Json(ApiResponse::ok(contract))
}

async fn update_price(
    State(state): State<AppState>,
    Path(code): Path<String>,
    Json(req): Json<UpdatePriceRequest>,
) -> Result<impl IntoResponse, AppError> {
    let mut engine = state.engine.lock().unwrap();
    let results = engine.update_price(&code, req.price)?;
    Ok(Json(ApiResponse::ok(results)))
}

async fn list_users(State(state): State<AppState>) -> impl IntoResponse {
    let engine = state.engine.lock().unwrap();
    let users: Vec<_> = engine.get_users().iter().map(|u| (*u).clone()).collect();
    Json(ApiResponse::ok(users))
}

async fn get_user(
    State(state): State<AppState>,
    Path(user_id): Path<String>,
) -> Result<impl IntoResponse, impl IntoResponse> {
    let engine = state.engine.lock().unwrap();
    match engine.get_user(&user_id) {
        Some(u) => Ok(Json(ApiResponse::ok(u.clone()))),
        None => Err((
            StatusCode::NOT_FOUND,
            Json(ApiResponse::<()>::err(format!("用户不存在: {}", user_id))),
        )),
    }
}

async fn register_user(
    State(state): State<AppState>,
    Path(user_id): Path<String>,
) -> impl IntoResponse {
    let mut engine = state.engine.lock().unwrap();
    engine.register_user(&user_id);
    Json(ApiResponse::ok(serde_json::json!({
        "user_id": user_id,
        "initial_funds": INITIAL_FUNDS,
    })))
}

async fn open_position(
    State(state): State<AppState>,
    Json(req): Json<OpenPositionRequest>,
) -> Result<impl IntoResponse, AppError> {
    let pt = parse_position_type(&req.position_type)?;
    let mut engine = state.engine.lock().unwrap();
    if engine.get_user(&req.user_id).is_none() {
        engine.register_user(&req.user_id);
    }
    let result = engine.open_position(&req.user_id, &req.contract_code, pt, req.lots)?;
    Ok(Json(ApiResponse::ok(result)))
}

async fn close_position(
    State(state): State<AppState>,
    Json(req): Json<ClosePositionRequest>,
) -> Result<impl IntoResponse, AppError> {
    let pt = parse_position_type(&req.position_type)?;
    let mut engine = state.engine.lock().unwrap();
    let results = engine.close_position(&req.user_id, &req.contract_code, pt, req.lots)?;
    Ok(Json(ApiResponse::ok(results)))
}

async fn settle(State(state): State<AppState>) -> impl IntoResponse {
    let mut engine = state.engine.lock().unwrap();
    let results = engine.settle();
    Json(ApiResponse::ok(results))
}

fn build_router(state: AppState) -> Router {
    Router::new()
        .route("/api/contracts", get(list_contracts).post(create_contract))
        .route("/api/contracts/:code", get(get_contract))
        .route("/api/contracts/:code/price", post(update_price))
        .route("/api/users", get(list_users))
        .route("/api/users/:user_id", get(get_user).post(register_user))
        .route("/api/positions/open", post(open_position))
        .route("/api/positions/close", post(close_position))
        .route("/api/settle", post(settle))
        .with_state(state)
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let engine = Arc::new(Mutex::new(TradingEngine::new()));
    let state = AppState { engine };

    let app = build_router(state);

    let addr = format!("{}:{}", args.host, args.port);
    println!("期货模拟交易服务已启动: http://{}", addr);

    let listener = tokio::net::TcpListener::bind(&addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
