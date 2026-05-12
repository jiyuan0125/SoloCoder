use actix_web::{web, HttpResponse, Responder};
use serde::{Deserialize, Serialize};
use std::time::Duration;
use crate::app_state::AppState;

#[derive(Debug, Deserialize)]
pub struct AddBackendRequest {
    pub path_prefix: String,
    pub target_url: String,
}

#[derive(Debug, Serialize)]
pub struct AddBackendResponse {
    pub id: String,
    pub path_prefix: String,
    pub target_url: String,
}

#[derive(Debug, Deserialize)]
pub struct SetTimeoutRequest {
    pub seconds: u64,
}

pub async fn add_backend(
    state: web::Data<AppState>,
    req: web::Json<AddBackendRequest>,
) -> impl Responder {
    let backend = state.add_backend(req.path_prefix.clone(), req.target_url.clone()).await;
    HttpResponse::Created().json(AddBackendResponse {
        id: backend.id.clone(),
        path_prefix: backend.path_prefix.clone(),
        target_url: backend.target_url.clone(),
    })
}

pub async fn delete_backend(state: web::Data<AppState>, path: web::Path<String>) -> impl Responder {
    let id = path.into_inner();
    if state.remove_backend(&id).await {
        HttpResponse::NoContent().finish()
    } else {
        HttpResponse::NotFound().finish()
    }
}

pub async fn list_backends(state: web::Data<AppState>) -> impl Responder {
    let backends = state.backends.read().await;
    let response: Vec<_> = backends.iter().map(|b| b.to_external()).collect();
    HttpResponse::Ok().json(response)
}

pub async fn get_stats(state: web::Data<AppState>) -> impl Responder {
    let stats = state.get_all_stats().await;
    HttpResponse::Ok().json(stats)
}

pub async fn set_timeout(
    state: web::Data<AppState>,
    req: web::Json<SetTimeoutRequest>,
) -> impl Responder {
    if req.seconds == 0 {
        return HttpResponse::BadRequest().body("Timeout must be greater than 0");
    }
    state.config.set_timeout(Duration::from_secs(req.seconds));
    HttpResponse::Ok().json(serde_json::json!({ "timeout_seconds": req.seconds }))
}
