use std::net::SocketAddr;
use std::sync::Arc;

use axum::{
    extract::{Path, Query, State},
    http::StatusCode,
    response::{IntoResponse, Json},
    routing::{delete, get, post},
    Router,
};
use clap::Parser;
use review_core::models::{
    CreateFollowUpRequest, CreateInitialReviewRequest, CreateReplyRequest, DeleteReviewRequest,
    Rating, ReviewFilter,
};
use review_core::ReviewService;
use serde::Deserialize;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
    
    #[arg(long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

#[derive(Clone)]
struct AppState {
    service: Arc<ReviewService>,
}

#[derive(Debug, Deserialize)]
struct QueryFilter {
    rating: Option<u32>,
    has_follow_up: Option<bool>,
    has_reply: Option<bool>,
    as_merchant: Option<bool>,
}

#[derive(Debug, Deserialize)]
struct ReviewIdQuery {
    as_merchant: Option<bool>,
}

async fn create_initial_review(
    State(state): State<AppState>,
    Json(req): Json<CreateInitialReviewRequest>,
) -> Result<impl IntoResponse, AppError> {
    let review = state.service.create_initial_review(req).await?;
    Ok((StatusCode::CREATED, Json(review)))
}

async fn create_follow_up(
    State(state): State<AppState>,
    Json(req): Json<CreateFollowUpRequest>,
) -> Result<impl IntoResponse, AppError> {
    let review = state.service.create_follow_up_review(req).await?;
    Ok((StatusCode::OK, Json(review)))
}

async fn create_reply(
    State(state): State<AppState>,
    Json(req): Json<CreateReplyRequest>,
) -> Result<impl IntoResponse, AppError> {
    let review = state.service.create_merchant_reply(req).await?;
    Ok((StatusCode::OK, Json(review)))
}

async fn create_supplement_reply(
    State(state): State<AppState>,
    Json(req): Json<CreateReplyRequest>,
) -> Result<impl IntoResponse, AppError> {
    let review = state.service.create_supplement_reply(req).await?;
    Ok((StatusCode::OK, Json(review)))
}

async fn delete_review(
    State(state): State<AppState>,
    Json(req): Json<DeleteReviewRequest>,
) -> Result<impl IntoResponse, AppError> {
    state.service.delete_review(req).await?;
    Ok(StatusCode::NO_CONTENT)
}

async fn get_reviews(
    State(state): State<AppState>,
    Path(product_id): Path<String>,
    Query(query_filter): Query<QueryFilter>,
) -> Result<impl IntoResponse, AppError> {
    let as_merchant = query_filter.as_merchant.unwrap_or(false);
    
    let review_filter = ReviewFilter {
        rating: query_filter.rating.and_then(Rating::from_u32),
        has_follow_up: query_filter.has_follow_up,
        has_reply: query_filter.has_reply,
    };

    let filter = if review_filter.rating.is_some()
        || review_filter.has_follow_up.is_some()
        || review_filter.has_reply.is_some()
    {
        Some(review_filter)
    } else {
        None
    };

    let reviews = state
        .service
        .get_reviews_for_product(&product_id, filter, as_merchant)
        .await;

    Ok((StatusCode::OK, Json(reviews)))
}

async fn get_review_by_id(
    State(state): State<AppState>,
    Path(review_id): Path<uuid::Uuid>,
    Query(query): Query<ReviewIdQuery>,
) -> Result<impl IntoResponse, AppError> {
    let as_merchant = query.as_merchant.unwrap_or(false);
    
    match state.service.get_review_by_id(review_id, as_merchant).await {
        Some(review) => Ok((StatusCode::OK, Json(review))),
        None => Err(AppError::NotFound),
    }
}

#[tokio::main]
async fn main() {
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "review_server=debug,tower_http=debug,axum=info".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    let args = Args::parse();

    let service = ReviewService::new();
    let app_state = AppState { service };

    let app = Router::new()
        .route("/api/reviews", post(create_initial_review))
        .route("/api/reviews/follow-up", post(create_follow_up))
        .route("/api/reviews/reply", post(create_reply))
        .route("/api/reviews/supplement-reply", post(create_supplement_reply))
        .route("/api/reviews", delete(delete_review))
        .route("/api/reviews/product/:product_id", get(get_reviews))
        .route("/api/reviews/:review_id", get(get_review_by_id))
        .with_state(app_state);

    let addr: SocketAddr = format!("{}:{}", args.host, args.port)
        .parse()
        .expect("Invalid address");

    tracing::info!("Server listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}

enum AppError {
    NotFound,
    Service(review_core::ReviewError),
}

impl From<review_core::ReviewError> for AppError {
    fn from(err: review_core::ReviewError) -> Self {
        AppError::Service(err)
    }
}

impl IntoResponse for AppError {
    fn into_response(self) -> axum::response::Response {
        match self {
            AppError::NotFound => (
                StatusCode::NOT_FOUND,
                Json(serde_json::json!({ "error": "Not Found" })),
            )
                .into_response(),
            AppError::Service(err) => {
                let status = match err {
                    review_core::ReviewError::ReviewNotFound => StatusCode::NOT_FOUND,
                    review_core::ReviewError::FollowUpWindowClosed => StatusCode::BAD_REQUEST,
                    review_core::ReviewError::FollowUpAlreadyExists => StatusCode::CONFLICT,
                    review_core::ReviewError::ReplyAlreadyExists => StatusCode::CONFLICT,
                    review_core::ReviewError::InvalidRating => StatusCode::BAD_REQUEST,
                    review_core::ReviewError::EmptyContent => StatusCode::BAD_REQUEST,
                    review_core::ReviewError::PermissionDenied => StatusCode::FORBIDDEN,
                    review_core::ReviewError::EmptyProductId => StatusCode::BAD_REQUEST,
                    review_core::ReviewError::EmptyCustomerId => StatusCode::BAD_REQUEST,
                };
                (status, Json(serde_json::json!({ "error": err.to_string() }))).into_response()
            }
        }
    }
}
