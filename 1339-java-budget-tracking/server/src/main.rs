use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{delete, get, post, put},
    Json, Router,
};
use clap::Parser;
use serde::Deserialize;
use tower_http::cors::CorsLayer;
use uuid::Uuid;

use budget_tracking_core::{
    AlertConfig, BudgetService, CreateBudgetCategoryRequest, CreateExpenseRequest,
    UpdateBudgetCategoryRequest,
};

#[derive(Parser, Debug)]
#[command(version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "BUDGET_SERVER_PORT")]
    port: Option<u16>,
    
    #[arg(short = 'H', long, env = "BUDGET_SERVER_HOST", default_value = "0.0.0.0")]
    host: String,
}

struct AppState {
    service: Arc<BudgetService>,
}

#[derive(Debug, Deserialize)]
struct ExecutionQuery {
    category_id: Option<Uuid>,
    start_date: Option<String>,
    end_date: Option<String>,
}

async fn create_category(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateBudgetCategoryRequest>,
) -> Result<impl IntoResponse, AppError> {
    let category = state.service.create_category(req)?;
    Ok((StatusCode::CREATED, Json(category)))
}

async fn list_categories(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let categories = state.service.list_categories();
    Json(categories)
}

async fn get_category_tree(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let tree = state.service.get_category_tree();
    Json(tree)
}

async fn update_category(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
    Json(req): Json<UpdateBudgetCategoryRequest>,
) -> Result<impl IntoResponse, AppError> {
    let category = state.service.update_category(id, req)?;
    Ok(Json(category))
}

async fn delete_category(
    State(state): State<Arc<AppState>>,
    Path(id): Path<Uuid>,
) -> Result<impl IntoResponse, AppError> {
    state.service.delete_category(id)?;
    Ok(StatusCode::NO_CONTENT)
}

async fn create_expense(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateExpenseRequest>,
) -> Result<impl IntoResponse, AppError> {
    let expense = state.service.create_expense(req)?;
    Ok((StatusCode::CREATED, Json(expense)))
}

async fn list_expenses(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let expenses = state.service.list_expenses();
    Json(expenses)
}

async fn get_alert_config(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let config = state.service.get_alert_config();
    Json(config)
}

async fn set_alert_config(
    State(state): State<Arc<AppState>>,
    Json(config): Json<AlertConfig>,
) -> Result<impl IntoResponse, AppError> {
    state.service.set_alert_config(config)?;
    Ok(StatusCode::NO_CONTENT)
}

async fn get_execution_summary(
    State(state): State<Arc<AppState>>,
    Query(query): Query<ExecutionQuery>,
) -> Result<impl IntoResponse, AppError> {
    let start_date = match query.start_date {
        Some(s) => Some(s.parse().map_err(|_| AppError::InvalidDate)?),
        None => None,
    };
    let end_date = match query.end_date {
        Some(s) => Some(s.parse().map_err(|_| AppError::InvalidDate)?),
        None => None,
    };

    let summaries = state
        .service
        .get_execution_summary(query.category_id, start_date, end_date)?;
    Ok(Json(summaries))
}

#[derive(Debug)]
enum AppError {
    Budget(budget_tracking_core::BudgetError),
    InvalidDate,
}

impl From<budget_tracking_core::BudgetError> for AppError {
    fn from(err: budget_tracking_core::BudgetError) -> Self {
        AppError::Budget(err)
    }
}

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        match self {
            AppError::Budget(err) => {
                use budget_tracking_core::BudgetError::*;
                let status = match err {
                    CategoryNotFound(_) => StatusCode::NOT_FOUND,
                    ParentNotFound(_) => StatusCode::BAD_REQUEST,
                    NonLeafCannotHaveBudget => StatusCode::BAD_REQUEST,
                    LeafMustHaveBudget => StatusCode::BAD_REQUEST,
                    CannotSetBudgetOnNonLeaf => StatusCode::BAD_REQUEST,
                    InvalidExpenseAmount => StatusCode::BAD_REQUEST,
                    InvalidDateRange => StatusCode::BAD_REQUEST,
                    InvalidThreshold { .. } => StatusCode::BAD_REQUEST,
                };
                (status, Json(serde_json::json!({ "error": err.to_string() }))).into_response()
            }
            AppError::InvalidDate => (
                StatusCode::BAD_REQUEST,
                Json(serde_json::json!({ "error": "Invalid date format" })),
            )
                .into_response(),
        }
    }
}

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let port = args.port.unwrap_or_else(|| {
        std::env::var("BUDGET_SERVER_PORT")
            .ok()
            .and_then(|s| s.parse().ok())
            .unwrap_or(3000)
    });

    let service = Arc::new(BudgetService::new());
    let state = Arc::new(AppState { service });

    let app = Router::new()
        .route("/api/categories", post(create_category).get(list_categories))
        .route("/api/categories/tree", get(get_category_tree))
        .route(
            "/api/categories/:id",
            put(update_category).delete(delete_category),
        )
        .route("/api/expenses", post(create_expense).get(list_expenses))
        .route("/api/alert-config", get(get_alert_config).put(set_alert_config))
        .route("/api/execution", get(get_execution_summary))
        .layer(CorsLayer::permissive())
        .with_state(state);

    let addr: SocketAddr = format!("{}:{}", args.host, port)
        .parse()
        .expect("Invalid address");

    println!("Budget tracking server running on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
