use std::net::SocketAddr;
use std::sync::Arc;

use axum::extract::{Query, State};
use axum::http::StatusCode;
use axum::response::IntoResponse;
use axum::routing::{get, post};
use axum::{Json, Router};
use chrono::NaiveDate;
use clap::Parser;
use invoice_core::config::InvoiceConfig;
use invoice_core::error::InvoiceError;
use invoice_core::models::{
    CreateInvoiceRequest, InvoiceStatus, Red冲InvoiceRequest, SearchInvoiceRequest,
    VoidInvoiceRequest,
};
use invoice_core::service::InvoiceService;
use rust_decimal::Decimal;
use serde::{Deserialize, Serialize};
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};
use uuid::Uuid;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "INVOICE_PORT", default_value_t = 3000)]
    port: u16,

    #[arg(long, env = "INVOICE_HOST", default_value = "127.0.0.1")]
    host: String,

    #[arg(long, env = "INVOICE_TAX_RATE")]
    tax_rate: Option<Decimal>,
}

struct AppState {
    service: InvoiceService,
}

#[derive(Debug, Serialize, Deserialize)]
struct ErrorResponse {
    error: String,
}

struct AppError(InvoiceError);

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        let status = match &self.0 {
            InvoiceError::InvoiceNumberExists(_, _) => StatusCode::CONFLICT,
            InvoiceError::FutureDate => StatusCode::BAD_REQUEST,
            InvoiceError::InvoiceNotFound(_) => StatusCode::NOT_FOUND,
            InvoiceError::AlreadyVoided => StatusCode::BAD_REQUEST,
            InvoiceError::AlreadyRed冲ed => StatusCode::BAD_REQUEST,
            InvoiceError::Red冲InvoiceCannotBeOperated => StatusCode::BAD_REQUEST,
            InvoiceError::AmountMustBePositive => StatusCode::BAD_REQUEST,
            InvoiceError::InvalidTaxRate => StatusCode::BAD_REQUEST,
            InvoiceError::Internal(_) => StatusCode::INTERNAL_SERVER_ERROR,
        };

        (status, Json(ErrorResponse { error: self.0.to_string() })).into_response()
    }
}

impl From<InvoiceError> for AppError {
    fn from(err: InvoiceError) -> Self {
        AppError(err)
    }
}

#[derive(Debug, Deserialize)]
struct SearchQuery {
    buyer_name: Option<String>,
    start_date: Option<String>,
    end_date: Option<String>,
    status: Option<String>,
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "invoice_server=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let config = if let Some(rate) = args.tax_rate {
        InvoiceConfig::new(rate).expect("无效的税率配置")
    } else {
        InvoiceConfig::default()
    };

    let service = InvoiceService::new(config);
    let app_state = Arc::new(AppState { service });

    let app = Router::new()
        .route("/invoices", post(create_invoice).get(list_invoices))
        .route("/invoices/search", get(search_invoices))
        .route("/invoices/:id", get(get_invoice))
        .route("/invoices/:id/void", post(void_invoice))
        .route("/invoices/:id/red冲", post(red冲_invoice))
        .route("/invoices/:id/chain", get(get_red冲_chain))
        .with_state(app_state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port).parse().unwrap();
    tracing::info!("发票管理服务启动于 {}", addr);

    axum::Server::bind(&addr)
        .serve(app.into_make_service())
        .await
        .unwrap();
}

async fn create_invoice(
    State(state): State<Arc<AppState>>,
    Json(req): Json<CreateInvoiceRequest>,
) -> Result<impl IntoResponse, AppError> {
    let invoice = state.service.create_invoice(req)?;
    Ok((StatusCode::CREATED, Json(invoice)))
}

async fn list_invoices(State(state): State<Arc<AppState>>) -> impl IntoResponse {
    let invoices = state.service.list_invoices();
    Json(invoices)
}

async fn search_invoices(
    State(state): State<Arc<AppState>>,
    Query(query): Query<SearchQuery>,
) -> impl IntoResponse {
    let start_date = query
        .start_date
        .and_then(|s| NaiveDate::parse_from_str(&s, "%Y-%m-%d").ok());
    let end_date = query
        .end_date
        .and_then(|s| NaiveDate::parse_from_str(&s, "%Y-%m-%d").ok());
    let status = query.status.and_then(|s| s.parse::<InvoiceStatus>().ok());

    let req = SearchInvoiceRequest {
        buyer_name: query.buyer_name,
        start_date,
        end_date,
        status,
    };

    let results = state.service.search_invoices(req);
    Json(results)
}

async fn get_invoice(
    State(state): State<Arc<AppState>>,
    axum::extract::Path(id): axum::extract::Path<String>,
) -> Result<impl IntoResponse, AppError> {
    let uuid = Uuid::parse_str(&id).map_err(|e| InvoiceError::Internal(e.to_string()))?;
    match state.service.get_invoice(uuid) {
        Some(inv) => Ok(Json(inv)),
        None => Err(InvoiceError::InvoiceNotFound(id).into()),
    }
}

async fn void_invoice(
    State(state): State<Arc<AppState>>,
    axum::extract::Path(id): axum::extract::Path<String>,
) -> Result<impl IntoResponse, AppError> {
    let uuid = Uuid::parse_str(&id).map_err(|e| InvoiceError::Internal(e.to_string()))?;
    let req = VoidInvoiceRequest { invoice_id: uuid };
    let invoice = state.service.void_invoice(req)?;
    Ok(Json(invoice))
}

async fn red冲_invoice(
    State(state): State<Arc<AppState>>,
    axum::extract::Path(id): axum::extract::Path<String>,
) -> Result<impl IntoResponse, AppError> {
    let uuid = Uuid::parse_str(&id).map_err(|e| InvoiceError::Internal(e.to_string()))?;
    let req = Red冲InvoiceRequest { invoice_id: uuid };
    let invoice = state.service.red冲_invoice(req)?;
    Ok(Json(invoice))
}

async fn get_red冲_chain(
    State(state): State<Arc<AppState>>,
    axum::extract::Path(id): axum::extract::Path<String>,
) -> Result<impl IntoResponse, AppError> {
    let uuid = Uuid::parse_str(&id).map_err(|e| InvoiceError::Internal(e.to_string()))?;
    match state.service.get_red冲_chain(uuid) {
        Some((original, current, red冲)) => Ok(Json(serde_json::json!({
            "original_invoice": original,
            "current_invoice": current,
            "red冲_invoice": red冲,
        }))),
        None => Err(InvoiceError::InvoiceNotFound(id).into()),
    }
}
