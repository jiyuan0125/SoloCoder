use actix_web::{web, HttpResponse, Responder};
use serde::{Deserialize, Serialize};

use crate::app_state::AppState;
use crate::models::{AddBackendRequest, DeleteBackendRequest};

#[derive(Debug, Serialize, Deserialize)]
struct ErrorResponse {
    error: String,
    position: String,
}

pub async fn add_backend(
    req: web::Json<AddBackendRequest>,
    state: web::Data<AppState>,
) -> impl Responder {
    let pattern = req.pattern.clone();
    let regex = match AppState::validate_regex(&pattern).await {
        Ok(r) => r,
        Err((msg, pos)) => {
            return HttpResponse::BadRequest().json(ErrorResponse {
                error: msg,
                position: pos,
            });
        }
    };

    let targets: Vec<(String, u32)> = req
        .targets
        .iter()
        .map(|t| (t.address.clone(), t.weight))
        .collect();

    state.add_route(pattern, regex, targets).await;
    HttpResponse::Ok().json("后端添加成功")
}

pub async fn delete_backend(
    req: web::Json<DeleteBackendRequest>,
    state: web::Data<AppState>,
) -> impl Responder {
    let pattern = req.pattern.clone();
    let address = req.address.clone();

    if state.mark_for_deletion(&pattern, &address).await {
        HttpResponse::Ok().json("后端已标记为删除，将在请求完成后移除")
    } else {
        HttpResponse::NotFound().json("未找到指定的后端")
    }
}

pub async fn list_status(state: web::Data<AppState>) -> impl Responder {
    let status = state.get_all_status().await;
    HttpResponse::Ok().json(status)
}
