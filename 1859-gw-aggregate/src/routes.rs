use std::collections::HashMap;

use crate::types::Route;

#[derive(Debug, Default)]
pub struct RouteRegistry {
    routes: HashMap<String, Route>,
    by_path: HashMap<String, String>,
}

impl RouteRegistry {
    pub fn new() -> Self {
        Self {
            routes: HashMap::new(),
            by_path: HashMap::new(),
        }
    }

    pub fn add(&mut self, route: Route) {
        self.by_path.insert(route.path.clone(), route.id.clone());
        self.routes.insert(route.id.clone(), route);
    }

    pub fn remove(&mut self, id: &str) -> bool {
        if let Some(route) = self.routes.remove(id) {
            self.by_path.remove(&route.path);
            true
        } else {
            false
        }
    }

    pub fn get_by_id(&self, id: &str) -> Option<&Route> {
        self.routes.get(id)
    }

    pub fn get_by_path(&self, path: &str) -> Option<&Route> {
        self.by_path.get(path).and_then(|id| self.routes.get(id))
    }

    pub fn list(&self) -> Vec<Route> {
        self.routes.values().cloned().collect()
    }
}
