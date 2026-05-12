use std::collections::HashMap;
use std::sync::Arc;

use serde::{Deserialize, Serialize};
use tokio::sync::RwLock;
use uuid::Uuid;

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum MatchType {
    Exact,
    Prefix,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
pub enum AuthStrategy {
    Public,
    ApiKey,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Route {
    pub id: Uuid,
    pub path: String,
    pub match_type: MatchType,
    pub backend_url: String,
    pub auth_strategy: AuthStrategy,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CreateRouteRequest {
    pub path: String,
    pub match_type: MatchType,
    pub backend_url: String,
    pub auth_strategy: AuthStrategy,
}

#[derive(Debug, Clone)]
pub struct RouteStore {
    exact_routes: Arc<RwLock<HashMap<String, Route>>>,
    prefix_routes: Arc<RwLock<Vec<Route>>>,
}

impl RouteStore {
    pub fn new() -> Self {
        Self {
            exact_routes: Arc::new(RwLock::new(HashMap::new())),
            prefix_routes: Arc::new(RwLock::new(Vec::new())),
        }
    }

    pub async fn add(&self, req: CreateRouteRequest) -> Route {
        let route = Route {
            id: Uuid::new_v4(),
            path: req.path,
            match_type: req.match_type,
            backend_url: req.backend_url,
            auth_strategy: req.auth_strategy,
        };

        match route.match_type {
            MatchType::Exact => {
                let mut exact = self.exact_routes.write().await;
                exact.insert(route.path.clone(), route.clone());
            }
            MatchType::Prefix => {
                let mut prefix = self.prefix_routes.write().await;
                prefix.push(route.clone());
            }
        }

        route
    }

    pub async fn remove(&self, id: Uuid) -> bool {
        let mut exact = self.exact_routes.write().await;
        let exact_before = exact.len();
        exact.retain(|_, r| r.id != id);
        let exact_removed = exact.len() != exact_before;

        let mut prefix = self.prefix_routes.write().await;
        let prefix_before = prefix.len();
        prefix.retain(|r| r.id != id);
        let prefix_removed = prefix.len() != prefix_before;

        exact_removed || prefix_removed
    }

    pub async fn match_route(&self, path: &str) -> Option<Route> {
        let exact = self.exact_routes.read().await;
        if let Some(route) = exact.get(path) {
            return Some(route.clone());
        }

        let prefix = self.prefix_routes.read().await;
        for route in prefix.iter() {
            let prefix_path = route.path.trim_end_matches('*');
            if path.starts_with(prefix_path) {
                return Some(route.clone());
            }
        }

        None
    }
}
