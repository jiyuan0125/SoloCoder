use grading_core::InMemoryStorage;

pub struct AppState {
    pub storage: InMemoryStorage,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            storage: InMemoryStorage::new(),
        }
    }
}

impl Default for AppState {
    fn default() -> Self {
        Self::new()
    }
}
