use crate::models::{ProcessRoute, ProductionTask};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

#[derive(Default, Clone)]
pub struct InMemoryStore {
    routes: Arc<Mutex<HashMap<String, ProcessRoute>>>,
    tasks: Arc<Mutex<HashMap<String, ProductionTask>>>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub fn get_route(&self, id: &str) -> Option<ProcessRoute> {
        let routes = self.routes.lock().unwrap();
        routes.get(id).cloned()
    }

    pub fn list_routes(&self) -> Vec<ProcessRoute> {
        let routes = self.routes.lock().unwrap();
        routes.values().cloned().collect()
    }

    pub fn save_route(&self, route: ProcessRoute) {
        let mut routes = self.routes.lock().unwrap();
        routes.insert(route.id.clone(), route);
    }

    pub fn delete_route(&self, id: &str) -> bool {
        let mut routes = self.routes.lock().unwrap();
        routes.remove(id).is_some()
    }

    pub fn get_task(&self, id: &str) -> Option<ProductionTask> {
        let tasks = self.tasks.lock().unwrap();
        tasks.get(id).cloned()
    }

    pub fn list_tasks(&self) -> Vec<ProductionTask> {
        let tasks = self.tasks.lock().unwrap();
        tasks.values().cloned().collect()
    }

    pub fn save_task(&self, task: ProductionTask) {
        let mut tasks = self.tasks.lock().unwrap();
        tasks.insert(task.id.clone(), task);
    }

    pub fn delete_task(&self, id: &str) -> bool {
        let mut tasks = self.tasks.lock().unwrap();
        tasks.remove(id).is_some()
    }

    pub fn update_task<F>(&self, id: &str, f: F) -> bool
    where
        F: FnOnce(&mut ProductionTask),
    {
        let mut tasks = self.tasks.lock().unwrap();
        if let Some(task) = tasks.get_mut(id) {
            f(task);
            true
        } else {
            false
        }
    }
}
