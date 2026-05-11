extern crate core_app as core_module;

use std::sync::Arc;

use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::IntoResponse,
    routing::{get, post},
    Json, Router,
};
use clap::Parser;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

use core_module::{
    Document, DocumentType, UrgencyLevel, WorkflowHistory,
    WorkflowService,
};

#[derive(Parser, Debug, Clone)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[arg(short, long, env = "PORT", default_value_t = 3000)]
    port: u16,
    #[arg(short, long, env = "HOST", default_value = "127.0.0.1")]
    host: String,
}

type AppState = Arc<WorkflowService>;

#[derive(Debug, Deserialize)]
struct CreateDocumentRequest {
    title: String,
    content: String,
    doc_type: String,
    urgency: String,
    author: String,
}

#[derive(Debug, Deserialize)]
struct SubmitRequest {
    operator: String,
}

#[derive(Debug, Deserialize)]
struct ReviewRequest {
    operator: String,
    approve: bool,
    reason: Option<String>,
}

#[derive(Debug, Deserialize)]
struct CountersignRequest {
    operator: String,
    approve: bool,
    comment: Option<String>,
    reason: Option<String>,
}

#[derive(Debug, Deserialize)]
struct UpdateResubmitRequest {
    operator: String,
    title: Option<String>,
    content: Option<String>,
}

#[derive(Debug, Serialize)]
struct ApiResponse<T> {
    success: bool,
    data: Option<T>,
    error: Option<String>,
}

impl<T> ApiResponse<T> {
    fn success(data: T) -> Self {
        ApiResponse {
            success: true,
            data: Some(data),
            error: None,
        }
    }

    fn error(msg: String) -> Self {
        ApiResponse {
            success: false,
            data: None,
            error: Some(msg),
        }
    }
}

fn parse_document_type(s: &str) -> Result<DocumentType, String> {
    DocumentType::from_str(s).ok_or_else(|| format!("Invalid document type: {}", s))
}

fn parse_urgency(s: &str) -> Result<UrgencyLevel, String> {
    UrgencyLevel::from_str(s).ok_or_else(|| format!("Invalid urgency level: {}", s))
}

async fn create_document(
    State(state): State<AppState>,
    Json(req): Json<CreateDocumentRequest>,
) -> impl IntoResponse {
    let doc_type = match parse_document_type(&req.doc_type) {
        Ok(t) => t,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e))),
    };
    let urgency = match parse_urgency(&req.urgency) {
        Ok(u) => u,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e))),
    };

    let doc = state.create_document(req.title, req.content, doc_type, urgency, req.author);
    (StatusCode::OK, Json(ApiResponse::<Document>::success(doc)))
}

async fn submit_document(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<SubmitRequest>,
) -> impl IntoResponse {
    let doc_id = match Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    };

    match state.submit_document(doc_id, req.operator) {
        Ok(doc) => (StatusCode::OK, Json(ApiResponse::<Document>::success(doc))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    }
}

async fn department_review(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<ReviewRequest>,
) -> impl IntoResponse {
    let doc_id = match Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    };

    match state.department_review(doc_id, req.operator, req.approve, req.reason) {
        Ok(doc) => (StatusCode::OK, Json(ApiResponse::<Document>::success(doc))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    }
}

async fn countersign(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<CountersignRequest>,
) -> impl IntoResponse {
    let doc_id = match Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    };

    match state.countersign(doc_id, req.operator, req.approve, req.comment, req.reason) {
        Ok(doc) => (StatusCode::OK, Json(ApiResponse::<Document>::success(doc))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    }
}

async fn issue_document(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<ReviewRequest>,
) -> impl IntoResponse {
    let doc_id = match Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    };

    match state.issue_document(doc_id, req.operator, req.approve, req.reason) {
        Ok(doc) => (StatusCode::OK, Json(ApiResponse::<Document>::success(doc))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    }
}

async fn update_and_resubmit(
    State(state): State<AppState>,
    Path(id): Path<String>,
    Json(req): Json<UpdateResubmitRequest>,
) -> impl IntoResponse {
    let doc_id = match Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    };

    match state.update_and_resubmit(doc_id, req.operator, req.title, req.content) {
        Ok(doc) => (StatusCode::OK, Json(ApiResponse::<Document>::success(doc))),
        Err(e) => (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    }
}

async fn get_document(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let doc_id = match Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<Document>::error(e.to_string()))),
    };

    match state.get_document(doc_id) {
        Some(doc) => (StatusCode::OK, Json(ApiResponse::<Document>::success(doc))),
        None => (StatusCode::NOT_FOUND, Json(ApiResponse::<Document>::error("Document not found".to_string()))),
    }
}

async fn list_documents(State(state): State<AppState>) -> impl IntoResponse {
    let docs = state.list_documents();
    (StatusCode::OK, Json(ApiResponse::<Vec<Document>>::success(docs)))
}

async fn get_history(
    State(state): State<AppState>,
    Path(id): Path<String>,
) -> impl IntoResponse {
    let doc_id = match Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(e) => return (StatusCode::BAD_REQUEST, Json(ApiResponse::<Vec<WorkflowHistory>>::error(e.to_string()))),
    };

    let history = state.get_history(doc_id);
    (StatusCode::OK, Json(ApiResponse::<Vec<WorkflowHistory>>::success(history)))
}

async fn check_overdue(State(state): State<AppState>) -> impl IntoResponse {
    let overdue = state.get_overdue_documents();
    (StatusCode::OK, Json(ApiResponse::<Vec<Document>>::success(overdue)))
}

#[tokio::main]
async fn main() {
    let args = Args::parse();
    let service = Arc::new(WorkflowService::new());

    let app = Router::new()
        .route("/api/documents", post(create_document).get(list_documents))
        .route("/api/documents/:id", get(get_document))
        .route("/api/documents/:id/submit", post(submit_document))
        .route("/api/documents/:id/review", post(department_review))
        .route("/api/documents/:id/countersign", post(countersign))
        .route("/api/documents/:id/issue", post(issue_document))
        .route("/api/documents/:id/resubmit", post(update_and_resubmit))
        .route("/api/documents/:id/history", get(get_history))
        .route("/api/overdue", get(check_overdue))
        .with_state(service);

    let addr = format!("{}:{}", args.host, args.port);
    println!("Server listening on {}", addr);

    axum::Server::bind(&addr.parse().unwrap())
        .serve(app.into_make_service())
        .await
        .unwrap();
}
