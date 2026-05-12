use crate::models::{BackendService, BackendStatus};
use crate::state::{AppState, DEGRADED_RECOVERY_CHECKS, ERROR_RATE_THRESHOLD};
use log::{error, info};
use reqwest::Client;

pub async fn perform_health_check(client: &Client, backend: &BackendService) -> bool {
    let health_url = if backend.url.ends_with('/') {
        format!("{}health", backend.url)
    } else {
        format!("{}/health", backend.url)
    };
    
    match client
        .get(&health_url)
        .timeout(std::time::Duration::from_secs(5))
        .send()
        .await
    {
        Ok(response) => response.status().is_success(),
        Err(e) => {
            error!("Health check failed for {}: {}", backend.name, e);
            false
        }
    }
}

pub async fn process_health_check_result(state: &AppState, client: &Client, backend: BackendService) {
    let is_healthy = perform_health_check(client, &backend).await;
    
    match backend.status {
        BackendStatus::Pending => {
            if is_healthy {
                info!(
                    "Backend '{}' passed health check, transitioning to Online",
                    backend.name
                );
                state.update_backend_status(&backend.name, BackendStatus::Online).await;
            }
        }
        
        BackendStatus::Online => {
            let error_rate = state.calculate_error_rate(&backend.name).await;
            if error_rate >= ERROR_RATE_THRESHOLD {
                info!(
                    "Backend '{}' error rate {:.1}% exceeds threshold {:.0}%, transitioning to Degraded",
                    backend.name,
                    error_rate * 100.0,
                    ERROR_RATE_THRESHOLD * 100.0
                );
                state.update_backend_status(&backend.name, BackendStatus::Degraded).await;
                state.reset_healthy_checks(&backend.name).await;
            }
        }
        
        BackendStatus::Degraded => {
            if is_healthy {
                let count = state.increment_healthy_checks(&backend.name).await;
                info!(
                    "Backend '{}' health check passed ({}/5)",
                    backend.name, count
                );
                if count >= DEGRADED_RECOVERY_CHECKS {
                    info!(
                        "Backend '{}' recovered {} consecutive checks, transitioning to Online",
                        backend.name, DEGRADED_RECOVERY_CHECKS
                    );
                    state.update_backend_status(&backend.name, BackendStatus::Online).await;
                }
            } else {
                state.reset_healthy_checks(&backend.name).await;
            }
        }
        
        BackendStatus::Offline => {
        }
    }
}
