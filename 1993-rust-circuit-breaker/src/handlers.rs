use std::sync::Arc;
use actix_web::{web, HttpResponse, Responder, http::StatusCode};
use serde_json::json;
use crate::breaker::{CircuitBreakerRegistry, CircuitConfigUpdate};
use crate::models::{CreateBreakerRequest, UpdateConfigRequest, ProxyRequest};

#[actix_web::post("/breakers")]
pub async fn create_breaker(
    registry: web::Data<Arc<CircuitBreakerRegistry>>,
    req: web::Json<CreateBreakerRequest>,
) -> impl Responder {
    let service_name = req.service_name.clone();
    
    if service_name.is_empty() {
        return HttpResponse::BadRequest().json(json!({
            "error": "service_name cannot be empty"
        }));
    }

    match registry.create(service_name.clone(), req.config.clone()) {
        Some(_) => {
            log::info!("Created circuit breaker for service: {}", service_name);
            HttpResponse::Created().json(json!({
                "message": "Circuit breaker created",
                "service_name": service_name
            }))
        }
        None => HttpResponse::Conflict().json(json!({
            "error": format!("Circuit breaker for service '{}' already exists", service_name)
        })),
    }
}

#[actix_web::get("/breakers/{service_name}")]
pub async fn get_breaker(
    registry: web::Data<Arc<CircuitBreakerRegistry>>,
    path: web::Path<String>,
) -> impl Responder {
    let service_name = path.into_inner();

    match registry.get(&service_name) {
        Some(breaker) => HttpResponse::Ok().json(breaker.get_status()),
        None => HttpResponse::NotFound().json(json!({
            "error": format!("Circuit breaker for service '{}' not found", service_name)
        })),
    }
}

#[actix_web::get("/breakers")]
pub async fn list_breakers(
    registry: web::Data<Arc<CircuitBreakerRegistry>>,
) -> impl Responder {
    let breakers = registry.list();
    let statuses: Vec<_> = breakers.iter().map(|b| b.get_status()).collect();
    
    HttpResponse::Ok().json(json!({
        "total": statuses.len(),
        "breakers": statuses
    }))
}

#[actix_web::patch("/breakers/{service_name}/config")]
pub async fn update_config(
    registry: web::Data<Arc<CircuitBreakerRegistry>>,
    path: web::Path<String>,
    req: web::Json<UpdateConfigRequest>,
) -> impl Responder {
    let service_name = path.into_inner();

    match registry.get(&service_name) {
        Some(breaker) => {
            let update = CircuitConfigUpdate {
                failure_threshold: req.failure_threshold,
                window_size: req.window_size,
                cooling_period_secs: req.cooling_period_secs,
                notify_url: req.notify_url.clone(),
                target_url: req.target_url.clone(),
            };
            
            breaker.update_config(&update);
            
            log::info!("Updated config for service: {}", service_name);
            
            HttpResponse::Ok().json(breaker.get_status())
        }
        None => HttpResponse::NotFound().json(json!({
            "error": format!("Circuit breaker for service '{}' not found", service_name)
        })),
    }
}

#[actix_web::post("/breakers/{service_name}/force-open")]
pub async fn force_open(
    registry: web::Data<Arc<CircuitBreakerRegistry>>,
    path: web::Path<String>,
) -> impl Responder {
    let service_name = path.into_inner();

    match registry.get(&service_name) {
        Some(breaker) => {
            breaker.force_open();
            log::info!("Forced open circuit breaker for service: {}", service_name);
            HttpResponse::Ok().json(breaker.get_status())
        }
        None => HttpResponse::NotFound().json(json!({
            "error": format!("Circuit breaker for service '{}' not found", service_name)
        })),
    }
}

#[actix_web::post("/breakers/{service_name}/force-close")]
pub async fn force_close(
    registry: web::Data<Arc<CircuitBreakerRegistry>>,
    path: web::Path<String>,
) -> impl Responder {
    let service_name = path.into_inner();

    match registry.get(&service_name) {
        Some(breaker) => {
            breaker.force_close();
            log::info!("Forced close circuit breaker for service: {}", service_name);
            HttpResponse::Ok().json(breaker.get_status())
        }
        None => HttpResponse::NotFound().json(json!({
            "error": format!("Circuit breaker for service '{}' not found", service_name)
        })),
    }
}

#[actix_web::delete("/breakers/{service_name}")]
pub async fn delete_breaker(
    registry: web::Data<Arc<CircuitBreakerRegistry>>,
    path: web::Path<String>,
) -> impl Responder {
    let service_name = path.into_inner();

    if registry.delete(&service_name) {
        log::info!("Deleted circuit breaker for service: {}", service_name);
        HttpResponse::Ok().json(json!({
            "message": format!("Circuit breaker for service '{}' deleted", service_name)
        }))
    } else {
        HttpResponse::NotFound().json(json!({
            "error": format!("Circuit breaker for service '{}' not found", service_name)
        }))
    }
}

#[actix_web::post("/breakers/{service_name}/proxy")]
pub async fn proxy_request(
    registry: web::Data<Arc<CircuitBreakerRegistry>>,
    path: web::Path<String>,
    req: Option<web::Json<ProxyRequest>>,
) -> impl Responder {
    let service_name = path.into_inner();

    let breaker = match registry.get(&service_name) {
        Some(b) => b,
        None => return HttpResponse::NotFound().json(json!({
            "error": format!("Circuit breaker for service '{}' not found", service_name)
        })),
    };

    let config = breaker.get_config();
    let target_url = match &config.target_url {
        Some(url) => url.clone(),
        None => return HttpResponse::BadRequest().json(json!({
            "error": format!("No target_url configured for service '{}'", service_name)
        })),
    };

    if !breaker.try_acquire() {
        return HttpResponse::ServiceUnavailable().json(json!({
            "error": "Circuit breaker is open or half-open with probe in progress",
            "state": breaker.get_state().as_str(),
            "service_name": service_name
        }));
    }

    let client = reqwest::Client::builder()
        .timeout(std::time::Duration::from_secs(30))
        .build()
        .unwrap();

    let body = req.and_then(|r| r.body.clone());

    let result = match body {
        Some(b) => client.post(&target_url).json(&b).send().await,
        None => client.get(&target_url).send().await,
    };

    match result {
        Ok(response) => {
            if response.status().is_success() {
                breaker.on_success();
                let status = response.status();
                let body = response.text().await.unwrap_or_default();
                
                HttpResponse::build(StatusCode::from_u16(status.as_u16()).unwrap())
                    .content_type("application/json")
                    .body(body)
            } else {
                breaker.on_failure();
                let status = response.status();
                let body = response.text().await.unwrap_or_default();
                
                HttpResponse::build(StatusCode::from_u16(status.as_u16()).unwrap())
                    .content_type("application/json")
                    .body(body)
            }
        }
        Err(e) => {
            breaker.on_failure();
            log::error!("Proxy request failed for service {}: {}", service_name, e);
            
            HttpResponse::BadGateway().json(json!({
                "error": "Backend service unavailable",
                "service_name": service_name,
                "details": e.to_string()
            }))
        }
    }
}
