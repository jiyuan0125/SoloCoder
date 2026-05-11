use std::sync::Arc;
use dorm_core::DormService;

pub struct AppState {
    pub service: Arc<DormService>,
}
