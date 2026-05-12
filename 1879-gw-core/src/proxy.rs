use crate::models::{AuthPolicy, BackendStatus};
use crate::state::AppState;
use actix_web::{
    http::Method,
    web::{Bytes, Data},
    HttpResponse,
};
use reqwest::Client;
use std::time::Duration;

pub async fn forward_request(
    state: Data<AppState>,
    client: Data<Client>,
    method: Method,
    path: String,
    api_key_header: Option<String>,
    body: Bytes,
) -> HttpResponse {
    let route = match state.find_route(&path).await {
        Some(r) => r,
        None => {
            return HttpResponse::NotFound().body("Route not found");
        }
    };

    if route.auth_policy == AuthPolicy::ApiKey {
        let valid = match api_key_header {
            Some(key) => state.validate_api_key(&key).await,
            None => false,
        };
        if !valid {
            return HttpResponse::Unauthorized().body("Invalid or missing X-API-Key");
        }
    }

    let backend = match state.get_backend(&route.backend_name).await {
        Some(b) => b,
        None => {
            return HttpResponse::ServiceUnavailable().body("Backend service not found");
        }
    };

    match backend.status {
        BackendStatus::Pending | BackendStatus::Offline => {
            return HttpResponse::ServiceUnavailable().body("Backend service unavailable");
        }
        BackendStatus::Degraded => {
            if method != Method::GET {
                state.record_request(&backend.name, true).await;
                return HttpResponse::ServiceUnavailable().body("Service degraded: read-only mode");
            }
        }
        BackendStatus::Online => {}
    }

    let target_url = build_target_url(&backend.url, &path);
    let reqwest_method = convert_method(&method);
    
    let mut request_builder = client
        .request(reqwest_method, &target_url)
        .timeout(Duration::from_secs(30));

    if !body.is_empty() {
        request_builder = request_builder.body(body);
    }

    let response = match request_builder.send().await {
        Ok(resp) => resp,
        Err(e) => {
            state.record_request(&backend.name, true).await;
            return HttpResponse::BadGateway().body(format!("Backend error: {}", e));
        }
    };

    let status = response.status();
    let status_code = actix_web::http::StatusCode::from_u16(status.as_u16())
        .unwrap_or(actix_web::http::StatusCode::INTERNAL_SERVER_ERROR);
    let is_5xx = status.is_server_error();
    
    state.record_request(&backend.name, is_5xx).await;

    let mut builder = HttpResponse::build(status_code);
    for (key, value) in response.headers().iter() {
        if let Ok(v) = value.to_str() {
            builder.insert_header((key.as_str(), v));
        }
    }

    let body = match response.bytes().await {
        Ok(b) => b,
        Err(e) => {
            return HttpResponse::BadGateway().body(format!("Error reading response: {}", e));
        }
    };

    builder.body(body)
}

fn build_target_url(backend_url: &str, path: &str) -> String {
    if backend_url.ends_with('/') && path.starts_with('/') {
        format!("{}{}", backend_url.trim_end_matches('/'), path)
    } else if !backend_url.ends_with('/') && !path.starts_with('/') {
        format!("{}/{}", backend_url, path)
    } else {
        format!("{}{}", backend_url, path)
    }
}

fn convert_method(method: &Method) -> reqwest::Method {
    match *method {
        Method::GET => reqwest::Method::GET,
        Method::POST => reqwest::Method::POST,
        Method::PUT => reqwest::Method::PUT,
        Method::DELETE => reqwest::Method::DELETE,
        Method::HEAD => reqwest::Method::HEAD,
        Method::OPTIONS => reqwest::Method::OPTIONS,
        Method::CONNECT => reqwest::Method::CONNECT,
        Method::PATCH => reqwest::Method::PATCH,
        Method::TRACE => reqwest::Method::TRACE,
        _ => reqwest::Method::GET,
    }
}
