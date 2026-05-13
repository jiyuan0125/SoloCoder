use log::{info, warn};
use std::time::Duration;
use tokio::net::TcpStream;
use tokio::time::timeout;
use url::Url;

use crate::app_state::AppState;

const HEALTH_CHECK_INTERVAL: u64 = 10;
const HEALTH_CHECK_TIMEOUT: u64 = 5;
const CONSECUTIVE_FAILURES_THRESHOLD: u32 = 3;
const CONSECUTIVE_SUCCESSES_THRESHOLD: u32 = 3;

pub async fn run_health_checks(state: AppState) {
    let mut interval = tokio::time::interval(Duration::from_secs(HEALTH_CHECK_INTERVAL));
    loop {
        interval.tick().await;
        perform_health_checks(&state).await;
        state.remove_marked_backends().await;
    }
}

async fn perform_health_checks(state: &AppState) {
    let all_backends = state.get_all_backends().await;
    for (pattern, backends) in &all_backends {
        for (address, original_weight, current_weight, is_healthy, _, _pending, marked) in backends {
            if *marked {
                continue;
            }

            let addr = address.clone();
            let healthy = check_backend(&addr).await;

            update_backend_status(
                state,
                pattern,
                &addr,
                healthy,
                *original_weight,
                *current_weight,
                *is_healthy,
            )
            .await;
        }
    }
}

fn extract_host_port(address: &str) -> Option<String> {
    if let Ok(url) = Url::parse(address) {
        let host = url.host_str()?;
        let port = url.port().unwrap_or_else(|| {
            if url.scheme() == "https" { 443 } else { 80 }
        });
        Some(format!("{}:{}", host, port))
    } else {
        Some(address.to_string())
    }
}

async fn check_backend(address: &str) -> bool {
    let target = match extract_host_port(address) {
        Some(t) => t,
        None => return false,
    };

    let duration = Duration::from_secs(HEALTH_CHECK_TIMEOUT);
    match timeout(duration, TcpStream::connect(target)).await {
        Ok(Ok(_)) => true,
        _ => false,
    }
}

async fn update_backend_status(
    state: &AppState,
    pattern: &str,
    address: &str,
    healthy: bool,
    original_weight: u32,
    current_weight: u32,
    _current_is_healthy: bool,
) {
    let pattern = pattern.to_string();
    let address = address.to_string();
    let pattern_clone = pattern.clone();
    let address_clone = address.clone();

    state
        .update_health(&pattern, &address, move |backend| {
            if healthy {
                backend.consecutive_successes += 1;
                backend.consecutive_failures = 0;

                if current_weight != original_weight
                    && backend.consecutive_successes >= CONSECUTIVE_SUCCESSES_THRESHOLD
                {
                    backend.current_weight = original_weight;
                    info!(
                        "后端 {} (路由 {}) 连续 {} 次健康检查通过，恢复原始权重 {}",
                        address_clone, pattern_clone, CONSECUTIVE_SUCCESSES_THRESHOLD, original_weight
                    );
                }
                backend.is_healthy = true;
            } else {
                backend.consecutive_failures += 1;
                backend.consecutive_successes = 0;

                if current_weight == original_weight
                    && backend.consecutive_failures >= CONSECUTIVE_FAILURES_THRESHOLD
                {
                    let new_weight = (original_weight as f64 * 0.1) as u32;
                    backend.current_weight = new_weight;
                    warn!(
                        "后端 {} (路由 {}) 连续 {} 次健康检查失败，权重降至 {}",
                        address_clone, pattern_clone, CONSECUTIVE_FAILURES_THRESHOLD, new_weight
                    );
                }
                if backend.consecutive_failures >= CONSECUTIVE_FAILURES_THRESHOLD {
                    backend.is_healthy = false;
                }
            }
        })
        .await;
}
