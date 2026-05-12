use actix_web::{
    web, HttpResponse, Responder,
};
use serde::{Deserialize, Serialize};

use crate::app_state::AppState;
use crate::sliding_window::SlidingWindowConfig;
use crate::token_bucket::TokenBucketConfig;

#[derive(Debug, Deserialize)]
pub struct SetSlidingWindowConfigRequest {
    pub window_secs: u64,
    pub limit: u64,
}

#[derive(Debug, Serialize)]
pub struct SlidingWindowConfigResponse {
    pub window_secs: u64,
    pub limit: u64,
}

#[derive(Debug, Deserialize)]
pub struct SetTokenBucketRuleRequest {
    pub path: String,
    pub capacity: u64,
    pub fill_rate: u64,
}

#[derive(Debug, Deserialize)]
pub struct RemoveTokenBucketRuleRequest {
    pub path: String,
}

#[derive(Debug, Deserialize)]
pub struct SetWebhookRequest {
    pub url: String,
}

#[derive(Debug, Serialize)]
pub struct ApiResponse<T: Serialize> {
    pub success: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub data: Option<T>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub message: Option<String>,
}

impl<T: Serialize> ApiResponse<T> {
    pub fn ok(data: T) -> Self {
        ApiResponse {
            success: true,
            data: Some(data),
            message: None,
        }
    }

    pub fn ok_msg(message: &str) -> ApiResponse<()> {
        ApiResponse {
            success: true,
            data: None,
            message: Some(message.to_string()),
        }
    }

    pub fn error(message: &str) -> ApiResponse<()> {
        ApiResponse {
            success: false,
            data: None,
            message: Some(message.to_string()),
        }
    }
}

pub async fn set_sliding_window_config(
    state: web::Data<AppState>,
    req: web::Json<SetSlidingWindowConfigRequest>,
) -> impl Responder {
    let config = SlidingWindowConfig {
        window_secs: req.window_secs,
        limit: req.limit,
    };
    state.sliding_window_limiter.update_config(config);
    HttpResponse::Ok().json(ApiResponse::<()>::ok_msg(
        "Sliding window config updated successfully",
    ))
}

pub async fn get_sliding_window_config(
    state: web::Data<AppState>,
) -> impl Responder {
    let config = state.sliding_window_limiter.get_config();
    let response = SlidingWindowConfigResponse {
        window_secs: config.window_secs,
        limit: config.limit,
    };
    HttpResponse::Ok().json(ApiResponse::ok(response))
}

pub async fn set_token_bucket_rule(
    state: web::Data<AppState>,
    req: web::Json<SetTokenBucketRuleRequest>,
) -> impl Responder {
    let config = TokenBucketConfig {
        capacity: req.capacity,
        fill_rate: req.fill_rate,
    };
    state.token_bucket_limiter.set_rule(&req.path, config);
    HttpResponse::Ok().json(ApiResponse::<()>::ok_msg(
        &format!("Token bucket rule for {} set successfully", req.path),
    ))
}

pub async fn remove_token_bucket_rule(
    state: web::Data<AppState>,
    req: web::Json<RemoveTokenBucketRuleRequest>,
) -> impl Responder {
    state.token_bucket_limiter.remove_rule(&req.path);
    HttpResponse::Ok().json(ApiResponse::<()>::ok_msg(
        &format!("Token bucket rule for {} removed", req.path),
    ))
}

pub async fn list_token_bucket_rules(
    state: web::Data<AppState>,
) -> impl Responder {
    let rules = state.token_bucket_limiter.get_all_rules();
    HttpResponse::Ok().json(ApiResponse::ok(rules))
}

pub async fn set_webhook(
    state: web::Data<AppState>,
    req: web::Json<SetWebhookRequest>,
) -> impl Responder {
    state.alert_manager.set_webhook(req.url.clone());
    HttpResponse::Ok().json(ApiResponse::<()>::ok_msg(
        "Webhook URL updated successfully",
    ))
}

pub async fn get_webhook(
    state: web::Data<AppState>,
) -> impl Responder {
    #[derive(Debug, Serialize)]
    struct WebhookResponse {
        url: Option<String>,
    }
    let url = state.alert_manager.get_webhook();
    HttpResponse::Ok().json(ApiResponse::ok(WebhookResponse { url }))
}

pub async fn health_check() -> impl Responder {
    HttpResponse::Ok().json(serde_json::json!({
        "status": "ok"
    }))
}
