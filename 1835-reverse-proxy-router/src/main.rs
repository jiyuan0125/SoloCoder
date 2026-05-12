use actix_web::{web, App, HttpResponse, HttpServer, Responder, Error, HttpRequest};
use actix_web::http::StatusCode;
use reqwest::Client as ReqwestClient;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Mutex;
use std::time::Duration;
use uuid::Uuid;
use chrono::{DateTime, Utc};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
enum BackendStatus {
    ComingUp,
    Active,
    Down,
}

#[derive(Debug, Clone, Serialize)]
struct Backend {
    id: String,
    path_prefix: String,
    address: String,
    status: BackendStatus,
    last_check_time: Option<DateTime<Utc>>,
    consecutive_failures: u32,
}

#[derive(Debug, Deserialize)]
struct AddBackendRequest {
    path_prefix: String,
    address: String,
}

#[derive(Clone)]
struct AppState {
    backends: web::Data<Mutex<HashMap<String, Backend>>>,
    next_index: web::Data<Mutex<HashMap<String, usize>>>,
}

fn normalize_prefix(prefix: &str) -> String {
    let mut p = prefix.trim().to_string();
    if !p.starts_with('/') {
        p.insert(0, '/');
    }
    if p.ends_with('/') && p.len() > 1 {
        p.pop();
    }
    p
}

fn path_matches(request_path: &str, prefix: &str) -> bool {
    if request_path == prefix {
        return true;
    }
    let with_slash = if prefix.ends_with('/') {
        prefix.to_string()
    } else {
        format!("{}/", prefix)
    };
    request_path.starts_with(&with_slash)
}

async fn add_backend(
    data: web::Data<AppState>,
    req: web::Json<AddBackendRequest>,
) -> impl Responder {
    let path_prefix = normalize_prefix(&req.path_prefix);
    let address = req.address.trim().trim_end_matches('/').to_string();

    let id = Uuid::new_v4().to_string();
    let backend = Backend {
        id: id.clone(),
        path_prefix: path_prefix.clone(),
        address: address.clone(),
        status: BackendStatus::ComingUp,
        last_check_time: None,
        consecutive_failures: 0,
    };

    let mut backends = data.backends.lock().unwrap();
    backends.insert(id.clone(), backend);

    HttpResponse::Created().json(serde_json::json!({
        "id": id,
        "path_prefix": path_prefix,
        "address": address,
        "status": "ComingUp"
    }))
}

async fn delete_backend(
    data: web::Data<AppState>,
    path: web::Path<String>,
) -> impl Responder {
    let id = path.into_inner();
    let mut backends = data.backends.lock().unwrap();
    
    if backends.remove(&id).is_some() {
        HttpResponse::NoContent().finish()
    } else {
        HttpResponse::NotFound().json(serde_json::json!({
            "error": "Backend not found"
        }))
    }
}

async fn list_backends(data: web::Data<AppState>) -> impl Responder {
    let backends = data.backends.lock().unwrap();
    let list: Vec<&Backend> = backends.values().collect();
    HttpResponse::Ok().json(list)
}

fn select_backend<'a>(
    backends: &'a HashMap<String, Backend>,
    request_path: &str,
    next_index: &mut HashMap<String, usize>,
) -> Option<&'a Backend> {
    let mut matched_prefixes: Vec<String> = Vec::new();
    
    for backend in backends.values() {
        if path_matches(request_path, &backend.path_prefix) {
            if !matched_prefixes.contains(&backend.path_prefix) {
                matched_prefixes.push(backend.path_prefix.clone());
            }
        }
    }
    
    if matched_prefixes.is_empty() {
        return None;
    }
    
    matched_prefixes.sort_by(|a, b| b.len().cmp(&a.len()));
    let best_prefix = &matched_prefixes[0];
    
    let active_backends: Vec<&Backend> = backends
        .values()
        .filter(|b| b.path_prefix == *best_prefix && b.status == BackendStatus::Active)
        .collect();
    
    if active_backends.is_empty() {
        return None;
    }
    
    let idx = next_index.entry(best_prefix.clone()).or_insert(0);
    let selected = active_backends[*idx % active_backends.len()];
    *idx = (*idx + 1) % active_backends.len();
    
    Some(selected)
}

async fn health_check_task(state: AppState) {
    let client = reqwest::Client::builder()
        .timeout(Duration::from_secs(5))
        .build()
        .unwrap();

    loop {
        tokio::time::sleep(Duration::from_secs(10)).await;
        
        let ids_to_check: Vec<String> = {
            let backends = state.backends.lock().unwrap();
            backends.keys().cloned().collect()
        };
        
        for id in ids_to_check {
            let address = {
                let backends = state.backends.lock().unwrap();
                match backends.get(&id) {
                    Some(b) => b.address.clone(),
                    None => continue,
                }
            };
            
            let health_url = format!("{}/health", address);
            let result = client.get(&health_url).send().await;
            
            let is_healthy = match result {
                Ok(resp) => resp.status().is_success(),
                Err(_) => false,
            };
            
            let mut backends = state.backends.lock().unwrap();
            if let Some(backend) = backends.get_mut(&id) {
                backend.last_check_time = Some(Utc::now());
                
                if is_healthy {
                    backend.consecutive_failures = 0;
                    match backend.status {
                        BackendStatus::ComingUp => {
                            backend.status = BackendStatus::Active;
                        }
                        BackendStatus::Down => {
                            backend.status = BackendStatus::Active;
                        }
                        BackendStatus::Active => {}
                    }
                } else {
                    backend.consecutive_failures += 1;
                    if backend.consecutive_failures >= 3 {
                        backend.status = BackendStatus::Down;
                    }
                }
            }
        }
    }
}

async fn proxy_handler(
    req: HttpRequest,
    body: web::Bytes,
    data: web::Data<AppState>,
    client: web::Data<ReqwestClient>,
) -> Result<HttpResponse, Error> {
    let request_path = req.path().to_string();
    
    let backend = {
        let mut next_index = data.next_index.lock().unwrap();
        let backends = data.backends.lock().unwrap();
        match select_backend(&backends, &request_path, &mut next_index) {
            Some(b) => Some(b.clone()),
            None => None,
        }
    };
    
    let backend = match backend {
        Some(b) => b,
        None => {
            return Ok(HttpResponse::BadGateway()
                .content_type("application/json")
                .body(serde_json::to_string(&serde_json::json!({
                    "error": "No available backend for this path"
                })).unwrap()));
        }
    };
    
    let query_string = req.query_string();
    let mut forward_url = format!("{}{}", backend.address, request_path);
    if !query_string.is_empty() {
        forward_url = format!("{}?{}", forward_url, query_string);
    }
    
    let request_id = Uuid::new_v4().to_string();
    let client_ip = req.peer_addr()
        .map(|addr| addr.ip().to_string())
        .unwrap_or_else(|| "unknown".to_string());
    
    let forwarded_for = match req.headers().get("X-Forwarded-For") {
        Some(existing) => {
            let existing_str = existing.to_str().unwrap_or("");
            if existing_str.is_empty() {
                client_ip.clone()
            } else {
                format!("{}, {}", existing_str, client_ip)
            }
        }
        None => client_ip.clone(),
    };
    
    let method = match req.method().as_str() {
        "GET" => reqwest::Method::GET,
        "POST" => reqwest::Method::POST,
        "PUT" => reqwest::Method::PUT,
        "DELETE" => reqwest::Method::DELETE,
        "PATCH" => reqwest::Method::PATCH,
        "HEAD" => reqwest::Method::HEAD,
        "OPTIONS" => reqwest::Method::OPTIONS,
        _ => reqwest::Method::GET,
    };
    
    let mut forwarded_req = client.request(method, &forward_url);
    
    for (key, value) in req.headers() {
        let key_lower = key.as_str().to_lowercase();
        if key_lower != "host" 
            && key_lower != "content-length"
            && key_lower != "x-forwarded-for"
            && key_lower != "x-request-id" {
            forwarded_req = forwarded_req.header(key.as_str(), value.as_bytes());
        }
    }
    
    forwarded_req = forwarded_req
        .header("X-Forwarded-For", forwarded_for)
        .header("X-Request-ID", request_id)
        .body(body.to_vec());
    
    let response = match forwarded_req.send().await {
        Ok(resp) => resp,
        Err(_) => {
            return Ok(HttpResponse::BadGateway()
                .content_type("application/json")
                .body(serde_json::to_string(&serde_json::json!({
                    "error": "Backend connection failed"
                })).unwrap()));
        }
    };
    
    let status_code = response.status().as_u16();
    let status: StatusCode = StatusCode::from_u16(status_code).unwrap_or(StatusCode::INTERNAL_SERVER_ERROR);
    let mut client_resp = HttpResponse::build(status);
    
    for (key, value) in response.headers() {
        let key_lower = key.as_str().to_lowercase();
        if key_lower != "content-length" 
            && key_lower != "connection"
            && key_lower != "keep-alive"
            && key_lower != "proxy-authenticate"
            && key_lower != "proxy-authorization"
            && key_lower != "te"
            && key_lower != "trailers"
            && key_lower != "transfer-encoding"
            && key_lower != "upgrade" {
            client_resp.append_header((key.as_str(), value.as_bytes()));
        }
    }
    
    let resp_body = response.bytes().await.map_err(|e| {
        actix_web::error::ErrorInternalServerError(format!("Failed to read response body: {}", e))
    })?;
    
    Ok(client_resp.body(resp_body))
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    let port: u16 = std::env::var("PORT")
        .unwrap_or_else(|_| "8080".to_string())
        .parse()
        .expect("PORT must be a valid port number");
    
    let backends: HashMap<String, Backend> = HashMap::new();
    let next_index: HashMap<String, usize> = HashMap::new();
    
    let shared_backends = web::Data::new(Mutex::new(backends));
    let shared_next_index = web::Data::new(Mutex::new(next_index));
    
    let app_state = AppState {
        backends: shared_backends.clone(),
        next_index: shared_next_index.clone(),
    };
    
    let health_state = app_state.clone();
    tokio::spawn(async move {
        health_check_task(health_state).await;
    });
    
    let app_state_data = web::Data::new(app_state);
    
    println!("Reverse proxy router starting on port {}", port);
    
    HttpServer::new(move || {
        App::new()
            .app_data(app_state_data.clone())
            .app_data(web::Data::new(ReqwestClient::new()))
            .service(
                web::resource("/backends")
                    .route(web::post().to(add_backend))
                    .route(web::get().to(list_backends)),
            )
            .service(
                web::resource("/backends/{id}")
                    .route(web::delete().to(delete_backend)),
            )
            .default_service(
                web::route().to(proxy_handler),
            )
    })
    .bind(("0.0.0.0", port))?
    .run()
    .await
}
