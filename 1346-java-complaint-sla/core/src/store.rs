use std::collections::HashMap;
use std::sync::{Arc, RwLock};

use uuid::Uuid;

use crate::models::{Complaint, Handler};
use crate::errors::Result;

#[derive(Debug, Clone)]
pub struct InMemoryStore {
    complaints: Arc<RwLock<HashMap<Uuid, Complaint>>>,
    handlers: Arc<RwLock<HashMap<String, Handler>>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        let store = Self {
            complaints: Arc::new(RwLock::new(HashMap::new())),
            handlers: Arc::new(RwLock::new(HashMap::new())),
        };
        store.init_default_handlers();
        store
    }

    fn init_default_handlers(&self) {
        let handlers = vec![
            Handler::new("handler_1", "张经理", 1),
            Handler::new("handler_2", "李主管", 2),
            Handler::new("handler_3", "王总监", 3),
            Handler::new("handler_4", "赵副总", 4),
        ];
        
        let mut store = self.handlers.write().unwrap();
        for handler in handlers {
            store.insert(handler.id.clone(), handler);
        }
    }

    pub fn get_complaint(&self, id: Uuid) -> Result<Option<Complaint>> {
        let store = self.complaints.read().unwrap();
        Ok(store.get(&id).cloned())
    }

    pub fn list_complaints(&self) -> Result<Vec<Complaint>> {
        let store = self.complaints.read().unwrap();
        Ok(store.values().cloned().collect())
    }

    pub fn save_complaint(&self, complaint: Complaint) -> Result<()> {
        let mut store = self.complaints.write().unwrap();
        store.insert(complaint.id, complaint);
        Ok(())
    }

    pub fn get_handler(&self, id: &str) -> Result<Option<Handler>> {
        let store = self.handlers.read().unwrap();
        Ok(store.get(id).cloned())
    }

    pub fn list_handlers(&self) -> Result<Vec<Handler>> {
        let store = self.handlers.read().unwrap();
        Ok(store.values().cloned().collect())
    }

    pub fn add_handler(&self, handler: Handler) -> Result<()> {
        let mut store = self.handlers.write().unwrap();
        store.insert(handler.id.clone(), handler);
        Ok(())
    }

    pub fn get_higher_level_handler(&self, current_level: u32) -> Result<Option<Handler>> {
        let store = self.handlers.read().unwrap();
        let mut candidates: Vec<Handler> = store
            .values()
            .filter(|h| h.level > current_level)
            .cloned()
            .collect();
        
        candidates.sort_by_key(|h| h.level);
        Ok(candidates.first().cloned())
    }
}

impl Default for InMemoryStore {
    fn default() -> Self {
        Self::new()
    }
}
