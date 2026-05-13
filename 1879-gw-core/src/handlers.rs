use crate::models::*;
use crate::proxy::forward_request;
use crate::state::AppState;
use actix_web::{
    web::{self, Data, Bytes, Path},
    HttpRequest, HttpResponse, Responder,
};
use reqwest::Client;
use uuid::Uuid;

pub async fn create_route(
    req: web::Json<CreateRouteRequest>,
    state: Data<AppState>,
) -> impl Responder {
    let route = Route {
        id: Uuid::new_v4(),
        path: req.path.clone(),
        match_type: req.match_type,
        backend_name: req.backend_name.clone(),
        auth_policy: req.auth_policy,
    };
    
    state.add_route(route.clone()).await;
    HttpResponse::Created().json(route)
}

pub async fn delete_route(id: Path<String>, state: Data<AppState>) -> impl Responder {
    let uuid = match Uuid::parse_str(&id) {
        Ok(u) => u,
        Err(_) => return HttpResponse::BadRequest().body("Invalid UUID format"),
    };
    
    if state.remove_route(&uuid).await {
        HttpResponse::NoContent().finish()
    } else {
        HttpResponse::NotFound().body("Route not found")
    }
}

pub async fn get_routes(state: Data<AppState>) -> impl Responder {
    let routes = state.get_all_routes().await;
    HttpResponse::Ok().json(routes)
}

pub async fn create_api_key(
    req: web::Json<CreateApiKeyRequest>,
    state: Data<AppState>,
) -> impl Responder {
    let key = match &req.key {
        Some(k) => k.clone(),
        None => Uuid::new_v4().to_string(),
    };
    
    let api_key = state.register_api_key(key).await;
    HttpResponse::Created().json(api_key)
}

pub async fn get_api_keys(state: Data<AppState>) -> impl Responder {
    let keys = state.get_all_api_keys().await;
    HttpResponse::Ok().json(keys)
}

pub async fn get_services(state: Data<AppState>) -> impl Responder {
    let backends = state.get_all_backends().await;
    let response: Vec<ServiceStatusResponse> = backends
        .into_iter()
        .map(|b| ServiceStatusResponse {
            name: b.name,
            url: b.url,
            status: b.status,
        })
        .collect();
    
    HttpResponse::Ok().json(response)
}

pub async fn get_stats(state: Data<AppState>) -> impl Responder {
    let stats = state.get_all_stats().await;
    let mut response = Vec::new();
    
    for (name, stat) in stats {
        let error_rate = if stat.total_requests > 0 {
            stat.error_requests as f64 / stat.total_requests as f64
        } else {
            0.0
        };
        
        response.push(StatsResponse {
            backend_name: name,
            total_requests: stat.total_requests,
            error_requests: stat.error_requests,
            error_rate,
        });
    }
    
    HttpResponse::Ok().json(response)
}

pub async fn register_backend(
    req: web::Json<RegisterBackendRequest>,
    state: Data<AppState>,
) -> impl Responder {
    let backend = state.register_backend(req.name.clone(), req.url.clone()).await;
    HttpResponse::Created().json(backend)
}

pub async fn set_backend_offline(
    name: Path<String>,
    state: Data<AppState>,
) -> impl Responder {
    if state.get_backend(&name).await.is_none() {
        return HttpResponse::NotFound().body("Backend not found");
    }
    
    state.update_backend_status(&name, BackendStatus::Offline).await;
    HttpResponse::Ok().json(serde_json::json!({
        "name": name.to_string(),
        "status": "offline"
    }))
}

pub async fn proxy_handler(
    req: HttpRequest,
    body: Bytes,
    state: Data<AppState>,
    client: Data<Client>,
) -> impl Responder {
    let method = req.method().clone();
    let path = req.uri().path().to_string();
    
    let api_key_header = req
        .headers()
        .get("X-API-Key")
        .and_then(|v| v.to_str().ok())
        .map(|s| s.to_string());
    
    forward_request(state, client, method, path, api_key_header, body).await
}
