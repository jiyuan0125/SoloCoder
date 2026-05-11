use std::sync::Arc;

use archive_core::{ArchiveService, InMemoryRepository};
use tokio::sync::Mutex;

#[derive(Clone)]
pub struct AppState {
    pub service: Arc<Mutex<ArchiveService<InMemoryRepository>>>,
}
