use actix_web::{
    http::{header::HeaderName, StatusCode},
    web, Error, HttpRequest, HttpResponse,
};
use futures_util::TryStreamExt;
use std::time::Instant;
use tokio::time::timeout;
use tracing::{error, info};
use uuid::Uuid;

use crate::app_state::AppState;

pub async fn handle_request(
    req: HttpRequest,
    body: web::Payload,
    state: web::Data<AppState>,
) -> Result<HttpResponse, Error> {
    let path = req.path().to_string();
    let method = req.method().clone();
    
    let request_id = Uuid::new_v4().to_string();

    let backend = match state.find_backend(&path).await {
        Some(b) => b,
        None => {
            return Ok(HttpResponse::NotFound()
                .insert_header(("X-Request-ID", request_id))
                .body("No backend found for this path"));
        }
    };

    let target_path = path.strip_prefix(&backend.path_prefix).unwrap_or(&path);
    let target_url = format!(
        "{}{}{}",
        backend.target_url.trim_end_matches('/'),
        if target_path.starts_with('/') { "" } else { "/" },
        target_path
    );

    let client_ip = req
        .connection_info()
        .realip_remote_addr()
        .map(|ip| ip.to_string())
        .unwrap_or_else(|| "unknown".to_string());

    let timeout_duration = state.config.get_timeout();
    let start_time = Instant::now();
    let stats = state.get_stats(&backend.id).await;

    info!(
        "Proxying {} {} -> {} (request_id: {})",
        method, path, target_url, request_id
    );

    let result = timeout(timeout_duration, async {
        let mut proxy_req = state
            .http_client
            .request(method.clone(), &target_url);

        for (key, value) in req.headers().iter() {
            if key != "host" {
                proxy_req = proxy_req.header(key.as_str(), value.as_bytes());
            }
        }

        proxy_req = proxy_req
            .header("X-Forwarded-For", client_ip)
            .header("X-Request-ID", &request_id);

        let mut body_bytes = Vec::new();
        let mut body_stream = body;
        while let Some(chunk) = body_stream.try_next().await? {
            body_bytes.extend_from_slice(&chunk);
        }
        proxy_req = proxy_req.body(body_bytes);

        let response = proxy_req.send().await.map_err(|e| {
            error!("Backend request error: {}", e);
            actix_web::error::ErrorInternalServerError(e)
        })?;

        let status = response.status();
        let mut builder = HttpResponse::build(StatusCode::from_u16(status.as_u16()).unwrap());
        
        for (key, value) in response.headers().iter() {
            if let Ok(header_name) = HeaderName::from_bytes(key.as_str().as_bytes()) {
                builder.insert_header((header_name, value.as_bytes()));
            }
        }
        builder.insert_header(("X-Request-ID", request_id.clone()));

        let resp_body = response.bytes().await.map_err(|e| {
            error!("Response body error: {}", e);
            actix_web::error::ErrorInternalServerError(e)
        })?;

        Ok::<HttpResponse, Error>(builder.body(resp_body))
    }).await;

    let elapsed = start_time.elapsed().as_millis() as u64;

    match result {
        Ok(Ok(response)) => {
            if let Some(stat) = stats {
                stat.record_request(elapsed);
            }
            info!(
                "Request completed: {} -> {} ({}ms, request_id: {})",
                path, target_url, elapsed, request_id
            );
            Ok(response)
        }
        Ok(Err(e)) => {
            if let Some(stat) = stats {
                stat.record_request(elapsed);
            }
            error!("Proxy error: {} (request_id: {})", e, request_id);
            Ok(HttpResponse::BadGateway()
                .insert_header(("X-Request-ID", request_id))
                .body("Backend error"))
        }
        Err(_) => {
            if let Some(stat) = stats {
                stat.record_timeout();
            }
            error!(
                "Request timeout: {} -> {} (request_id: {})",
                path, target_url, request_id
            );
            Ok(HttpResponse::GatewayTimeout()
                .insert_header(("X-Request-ID", request_id))
                .body("Gateway timeout"))
        }
    }
}
