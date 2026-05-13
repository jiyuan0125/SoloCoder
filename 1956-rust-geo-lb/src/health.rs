use crate::app_state::AppState;
use std::sync::Arc;
use std::time::Instant;
use tokio::time::{self, Duration};
use tracing::{info, warn};

const CONSECUTIVE_FAILURE_THRESHOLD: u32 = 3;

pub async fn start_health_checker(state: Arc<AppState>) {
    loop {
        time::sleep(Duration::from_secs(5)).await;

        let zones = state.get_all_zones().await;

        for zone in zones {
            let health_interval = Duration::from_secs(zone.health_interval_sec);

            for backend in &zone.backends {
                let should_check = match backend.last_checked {
                    Some(last) => Instant::now().duration_since(last) >= health_interval,
                    None => true,
                };

                if should_check {
                    let backend_addr = backend.address.clone();
                    let zone_name = zone.name.clone();
                    let state_clone = state.clone();

                    tokio::spawn(async move {
                        check_backend_health(&state_clone, &zone_name, &backend_addr).await;
                    });
                }
            }
        }
    }
}

async fn check_backend_health(state: &Arc<AppState>, zone_name: &str, backend_addr: &str) {
    let url = format!("{}/health", backend_addr);

    let client = match reqwest::Client::builder()
        .timeout(Duration::from_secs(5))
        .build()
    {
        Ok(c) => c,
        Err(e) => {
            warn!("Failed to create HTTP client: {}", e);
            return;
        }
    };

    let start = Instant::now();
    let result = client.get(&url).send().await;
    let _elapsed = start.elapsed();

    let is_healthy = match result {
        Ok(resp) => resp.status().is_success(),
        Err(e) => {
            warn!("Health check failed for {}: {}", backend_addr, e);
            false
        }
    };

    update_backend_health(state, zone_name, backend_addr, is_healthy).await;
}

async fn update_backend_health(
    state: &Arc<AppState>,
    zone_name: &str,
    backend_addr: &str,
    is_healthy: bool,
) {
    let mut zones = state.zones.write().await;

    if let Some(zone) = zones.get_mut(zone_name) {
        if let Some(existing_zone) = Arc::get_mut(zone) {
            for backend in &mut existing_zone.backends {
                if backend.address == backend_addr {
                    let was_healthy = backend.healthy;
                    backend.last_checked = Some(Instant::now());

                    if is_healthy {
                        backend.consecutive_failures = 0;
                        if !was_healthy {
                            backend.healthy = true;
                            info!(
                                "Backend {} in zone {} is now healthy",
                                backend_addr, zone_name
                            );
                        }
                    } else {
                        backend.consecutive_failures += 1;
                        if backend.consecutive_failures >= CONSECUTIVE_FAILURE_THRESHOLD
                            && was_healthy
                        {
                            backend.healthy = false;
                            warn!(
                                "Backend {} in zone {} is now unhealthy ({} consecutive failures)",
                                backend_addr, zone_name, CONSECUTIVE_FAILURE_THRESHOLD
                            );
                        }
                    }

                    break;
                }
            }
        }
    }
}
