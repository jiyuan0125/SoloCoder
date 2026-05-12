use actix_web::{get, post, put, web, HttpResponse, Responder};
use serde::Deserialize;
use std::sync::Arc;
use tokio::sync::Mutex;
use uuid::Uuid;

use crate::channel::{ChannelConfig, ChannelManager};
use crate::gateway::{GatewayError, GatewayRequest, invoke_grpc_service};

type Manager = web::Data<Arc<Mutex<ChannelManager>>>;

#[derive(Debug, Deserialize)]
pub struct CreateChannelRequest {
    pub service: Option<String>,
    pub method: Option<String>,
    pub backend_address: Option<String>,
    pub proto: Option<String>,
}

#[derive(Debug, Deserialize)]
pub struct ConfigureChannelRequest {
    pub service: String,
    pub method: String,
    pub backend_address: String,
    pub proto: String,
}

#[get("/channels")]
pub async fn list_channels(manager: Manager) -> impl Responder {
    let mgr = manager.lock().await;
    let channels = mgr.list();
    HttpResponse::Ok().json(channels)
}

#[post("/channels")]
pub async fn create_channel(req: web::Json<CreateChannelRequest>, manager: Manager) -> impl Responder {
    let mut mgr = manager.lock().await;

    let config = if let (Some(s), Some(m), Some(a), Some(p)) = (
        req.service.clone(),
        req.method.clone(),
        req.backend_address.clone(),
        req.proto.clone(),
    ) {
        Some(ChannelConfig {
            service: s,
            method: m,
            backend_address: a,
            proto: p,
        })
    } else {
        None
    };

    let channel = mgr.create(config);
    HttpResponse::Created().json(channel)
}

#[get("/channels/{id}")]
pub async fn get_channel(path: web::Path<Uuid>, manager: Manager) -> impl Responder {
    let id = path.into_inner();
    let mgr = manager.lock().await;

    match mgr.get(&id) {
        Some(channel) => HttpResponse::Ok().json(channel),
        None => HttpResponse::NotFound().body("Channel not found"),
    }
}

#[put("/channels/{id}/configure")]
pub async fn configure_channel(
    path: web::Path<Uuid>,
    req: web::Json<ConfigureChannelRequest>,
    manager: Manager,
) -> impl Responder {
    let id = path.into_inner();
    let mut mgr = manager.lock().await;

    let config = ChannelConfig {
        service: req.service.clone(),
        method: req.method.clone(),
        backend_address: req.backend_address.clone(),
        proto: req.proto.clone(),
    };

    match mgr.configure(&id, config) {
        Ok(channel) => HttpResponse::Ok().json(channel),
        Err(e) => HttpResponse::NotFound().body(e),
    }
}

#[put("/channels/{id}/reset")]
pub async fn reset_channel(path: web::Path<Uuid>, manager: Manager) -> impl Responder {
    let id = path.into_inner();
    let mut mgr = manager.lock().await;

    match mgr.reset(&id) {
        Ok(channel) => HttpResponse::Ok().json(channel),
        Err(e) => HttpResponse::NotFound().body(e),
    }
}

#[get("/channels/{id}/stats")]
pub async fn get_channel_stats(path: web::Path<Uuid>, manager: Manager) -> impl Responder {
    let id = path.into_inner();
    let mgr = manager.lock().await;

    match mgr.get(&id) {
        Some(channel) => HttpResponse::Ok().json(channel.stats),
        None => HttpResponse::NotFound().body("Channel not found"),
    }
}

#[post("/v1/{service}/{method}")]
pub async fn gateway_request(
    path: web::Path<(String, String)>,
    body: web::Bytes,
    manager: Manager,
) -> impl Responder {
    let (service, method) = path.into_inner();

    let channel_id = {
        let mgr = manager.lock().await;
        let channel = match mgr.get_by_service_method(&service, &method) {
            Some(c) => c,
            None => return HttpResponse::NotFound().body("Channel not found for this service/method"),
        };

        if channel.status != crate::channel::ChannelStatus::Ready {
            return HttpResponse::ServiceUnavailable().body(format!("Channel is not ready: {:?}", channel.status));
        }

        channel.id
    };

    let config = {
        let mut mgr = manager.lock().await;
        match mgr.start_transforming(&channel_id) {
            Ok(channel) => channel.config.clone(),
            Err(e) => return HttpResponse::ServiceUnavailable().body(e),
        }
    };

    let config = match config {
        Some(c) => c,
        None => {
            let mut mgr = manager.lock().await;
            let _ = mgr.complete_error(&channel_id, "No config".to_string());
            return HttpResponse::InternalServerError().body("Channel not configured");
        }
    };

    let json_body: serde_json::Value = match serde_json::from_slice(&body) {
        Ok(v) => v,
        Err(e) => {
            let mut mgr = manager.lock().await;
            let _ = mgr.complete_error(&channel_id, format!("JSON parse error: {}", e));
            return HttpResponse::BadRequest().body(format!("Invalid JSON: {}", e));
        }
    };

    let gateway_req = GatewayRequest {
        backend_address: config.backend_address.clone(),
        service: config.service.clone(),
        method: config.method.clone(),
        json_body,
        proto_definition: config.proto.clone(),
    };

    match invoke_grpc_service(gateway_req).await {
        Ok(response) => {
            let mut mgr = manager.lock().await;
            let _ = mgr.complete_success(&channel_id);
            HttpResponse::Ok().content_type("application/json").body(response.to_string())
        }
        Err(e) => {
            let mut mgr = manager.lock().await;
            let _ = mgr.complete_error(&channel_id, format!("{:?}", e));

            match e {
                GatewayError::BackendUnavailable(_) => {
                    HttpResponse::BadGateway().body("Backend unavailable")
                }
                GatewayError::GrpcStatus(code, msg) => match code.as_str() {
                    "OK" => HttpResponse::Ok().body(msg),
                    "INVALID_ARGUMENT" => HttpResponse::BadRequest().body(msg),
                    "NOT_FOUND" => HttpResponse::NotFound().body(msg),
                    "UNAVAILABLE" => HttpResponse::ServiceUnavailable().body(msg),
                    "DEADLINE_EXCEEDED" => HttpResponse::GatewayTimeout().body(msg),
                    _ => HttpResponse::InternalServerError().body(msg),
                },
                _ => HttpResponse::InternalServerError().body(format!("{:?}", e)),
            }
        }
    }
}
