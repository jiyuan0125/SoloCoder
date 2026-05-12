use std::sync::atomic::Ordering;
use std::time::Duration;

use tokio::net::TcpStream;
use tokio::time::timeout;
use tracing::{error, info};

use crate::load_balancer::LoadBalancer;
use crate::types::BackendStatus;

const HEALTH_CHECK_INTERVAL: Duration = Duration::from_secs(5);
const HEALTH_CHECK_TIMEOUT: Duration = Duration::from_secs(3);
const FAILURE_THRESHOLD: u64 = 3;
const SUCCESS_THRESHOLD: u64 = 2;

pub async fn run_health_checks(lb: LoadBalancer) {
    info!("Health check task started, checking every {:?}", HEALTH_CHECK_INTERVAL);
    
    loop {
        let backends = lb.get_backends();
        
        for backend in backends {
            let lb_clone = lb.clone();
            let address = backend.address.clone();
            
            tokio::spawn(async move {
                let result = check_backend_health(&address).await;
                
                match result {
                    Ok(_) => {
                        backend.consecutive_failures.store(0, Ordering::Relaxed);
                        let successes = backend.consecutive_successes.fetch_add(1, Ordering::Relaxed) + 1;
                        
                        let current_status = *backend.status.read();
                        if current_status == BackendStatus::Unhealthy && successes >= SUCCESS_THRESHOLD {
                            info!("Backend recovered: {}", address);
                            lb_clone.update_backend_status(&address, BackendStatus::Healthy);
                        }
                    }
                    Err(e) => {
                        backend.consecutive_successes.store(0, Ordering::Relaxed);
                        let failures = backend.consecutive_failures.fetch_add(1, Ordering::Relaxed) + 1;
                        
                        let current_status = *backend.status.read();
                        if current_status == BackendStatus::Healthy && failures >= FAILURE_THRESHOLD {
                            error!("Backend marked unhealthy: {} - {}", address, e);
                            lb_clone.update_backend_status(&address, BackendStatus::Unhealthy);
                        }
                    }
                }
            });
        }
        
        lb.clean_up_draining();
        
        tokio::time::sleep(HEALTH_CHECK_INTERVAL).await;
    }
}

async fn check_backend_health(address: &str) -> Result<(), String> {
    match timeout(HEALTH_CHECK_TIMEOUT, TcpStream::connect(address)).await {
        Ok(Ok(_stream)) => Ok(()),
        Ok(Err(e)) => Err(format!("Connection failed: {}", e)),
        Err(_) => Err("Connection timed out".to_string()),
    }
}
