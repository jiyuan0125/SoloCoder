use presale_core::PresaleService;
use std::sync::Arc;

pub struct AppState {
    pub service: Arc<PresaleService>,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            service: Arc::new(PresaleService::new()),
        }
    }
}

impl Default for AppState {
    fn default() -> Self {
        Self::new()
    }
}
