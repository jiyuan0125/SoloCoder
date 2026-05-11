use std::sync::Arc;
use cert_core::CertificationService;

#[derive(Clone)]
pub struct AppState {
    pub service: Arc<CertificationService>,
}

impl AppState {
    pub fn new(service: CertificationService) -> Self {
        Self {
            service: Arc::new(service),
        }
    }
}
