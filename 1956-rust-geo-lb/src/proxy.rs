use crate::app_state::AppState;
use axum::extract::{ConnectInfo, Request};
use axum::http::{StatusCode, Uri};
use axum::response::{IntoResponse, Response};
use std::net::SocketAddr;
use std::sync::Arc;
use std::time::Instant;
use tracing::{debug, error, info, warn};

pub async fn proxy_request(
    state: Arc<AppState>,
    request: Request,
    client: &reqwest::Client,
) -> Response {
    let start = Instant::now();
    let client_ip = get_client_ip(&request);

    info!("Received request from IP: {:?}", client_ip);

    let (method, uri, headers, body) = (
        request.method().clone(),
        request.uri().clone(),
        request.headers().clone(),
        request.into_body(),
    );

    let (target_zone, backend_addr) = match select_target(&state, client_ip).await {
        Some((zone, backend)) => (zone, backend),
        None => {
            error!("No available zones or backends found");
            return (StatusCode::SERVICE_UNAVAILABLE, "No available backends").into_response();
        }
    };

    let backend_addr = backend_addr.clone();
    let zone_name = target_zone.name.clone();

    let target_url = match build_target_url(&backend_addr, &uri) {
        Ok(url) => url,
        Err(e) => {
            error!("Failed to build target URL: {}", e);
            return (StatusCode::INTERNAL_SERVER_ERROR, "Failed to build target URL").into_response();
        }
    };

    info!(
        "Forwarding request to {} (zone: {})",
        backend_addr, zone_name
    );

    let body_bytes = match axum::body::to_bytes(body, usize::MAX).await {
        Ok(b) => b,
        Err(e) => {
            error!("Failed to read request body: {}", e);
            return (StatusCode::INTERNAL_SERVER_ERROR, "Failed to read request body").into_response();
        }
    };

    let mut req_builder = client
        .request(method.clone(), &target_url)
        .body(body_bytes);

    for (key, value) in headers.iter() {
        req_builder = req_builder.header(key, value);
    }

    match req_builder.send().await {
        Ok(response) => {
            let status = response.status();
            let latency = start.elapsed();

            update_stats(&state, &zone_name, latency.as_millis() as u64).await;

            info!(
                "Request completed - zone: {}, backend: {}, status: {}, latency: {:?}",
                zone_name, backend_addr, status, latency
            );

            convert_response(response).await
        }
        Err(e) => {
            error!("Proxy request failed: {}", e);
            (
                StatusCode::BAD_GATEWAY,
                format!("Backend request failed: {}", e),
            )
                .into_response()
        }
    }
}

async fn select_target(
    state: &Arc<AppState>,
    client_ip: Option<std::net::IpAddr>,
) -> Option<(Arc<crate::models::ZoneState>, String)> {
    let preferred_zone = match client_ip {
        Some(ip) => {
            info!("Looking up zone for IP: {}", ip);
            let found = state.find_zone_for_ip(ip).await;
            if let Some(ref z) = found {
                info!("Found matching zone: {} for IP: {}", z.name, ip);
            } else {
                warn!("No zone found for IP: {}", ip);
            }
            found
        }
        None => {
            warn!("Client IP not available, cannot perform geo-routing");
            None
        }
    };

    let preferred_zone_name = preferred_zone
        .as_ref()
        .map(|z| z.name.clone())
        .unwrap_or_default();

    if let Some(ref zone) = preferred_zone {
        if zone.has_any_healthy() {
            let healthy_backends = zone.get_healthy_backends();
            if !healthy_backends.is_empty() {
                let index = state.get_next_backend_index(healthy_backends.len());
                info!(
                    "Using preferred zone {} with backend {}",
                    zone.name, healthy_backends[index].address
                );
                return Some((zone.clone(), healthy_backends[index].address.clone()));
            }
        } else {
            warn!("Preferred zone {} is completely unhealthy, switching to fallback", zone.name);
        }
    }

    let fallback_zones = state.get_next_available_zones(&preferred_zone_name).await;

    info!(
        "Available fallback zones ({} total): {:?}",
        fallback_zones.len(),
        fallback_zones.iter().map(|z| z.name.as_str()).collect::<Vec<_>>()
    );

    for fallback_zone in fallback_zones {
        if fallback_zone.name == preferred_zone_name {
            continue;
        }

        if fallback_zone.has_any_healthy() {
            let healthy_backends = fallback_zone.get_healthy_backends();
            if !healthy_backends.is_empty() {
                let index = state.get_next_backend_index(healthy_backends.len());
                info!(
                    "Using fallback zone {} (preferred: {})",
                    fallback_zone.name, preferred_zone_name
                );
                let addr = healthy_backends[index].address.clone();
                return Some((fallback_zone, addr));
            }
        }
    }

    None
}

fn build_target_url(backend: &str, uri: &Uri) -> Result<String, String> {
    let path_and_query = match (uri.path(), uri.query()) {
        (path, Some(query)) => format!("{}?{}", path, query),
        (path, None) => path.to_string(),
    };

    let backend = if backend.ends_with('/') {
        backend.trim_end_matches('/')
    } else {
        backend
    };

    Ok(format!("{}{}", backend, path_and_query))
}

async fn update_stats(state: &Arc<AppState>, zone_name: &str, latency_ms: u64) {
    let zones = state.zones.read().await;
    if let Some(zone) = zones.get(zone_name) {
        zone.request_count
            .fetch_add(1, std::sync::atomic::Ordering::SeqCst);
        zone.total_latency_ms
            .fetch_add(latency_ms, std::sync::atomic::Ordering::SeqCst);
    }
}

fn get_client_ip(request: &Request) -> Option<std::net::IpAddr> {
    debug!("Checking request extensions for ConnectInfo");

    if let Some(connect_info) = request.extensions().get::<ConnectInfo<SocketAddr>>() {
        let ip = connect_info.ip();
        info!("Extracted client IP from ConnectInfo: {}", ip);
        return Some(ip);
    }

    debug!("ConnectInfo not found in request extensions");

    None
}

async fn convert_response(response: reqwest::Response) -> Response {
    let status = response.status();
    let headers = response.headers().clone();

    match response.bytes().await {
        Ok(body_bytes) => {
            let mut builder = Response::builder().status(status);

            if let Some(h) = builder.headers_mut() {
                *h = headers;
            }

            match builder.body(axum::body::Body::from(body_bytes)) {
                Ok(resp) => resp,
                Err(e) => {
                    error!("Failed to build response: {}", e);
                    (StatusCode::INTERNAL_SERVER_ERROR, "Failed to build response").into_response()
                }
            }
        }
        Err(e) => {
            error!("Failed to read response body: {}", e);
            (StatusCode::BAD_GATEWAY, "Failed to read backend response").into_response()
        }
    }
}
