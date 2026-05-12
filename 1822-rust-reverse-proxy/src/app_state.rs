use std::collections::HashMap;
use std::sync::Arc;

use regex::Regex;
use tokio::sync::Mutex;

use crate::models::{BackendInner, BackendStatus, Route, RouteStatus};

#[derive(Clone)]
pub struct AppState {
    inner: Arc<Mutex<AppStateInner>>,
}

pub struct AppStateInner {
    routes: HashMap<String, Route>,
    backends_by_route: HashMap<String, Vec<BackendInner>>,
}

impl AppState {
    pub fn new() -> Self {
        Self {
            inner: Arc::new(Mutex::new(AppStateInner {
                routes: HashMap::new(),
                backends_by_route: HashMap::new(),
            })),
        }
    }

    pub async fn validate_regex(pattern: &str) -> Result<Regex, (String, String)> {
        match Regex::new(pattern) {
            Ok(re) => Ok(re),
            Err(e) => {
                let pos = match e {
                    regex::Error::Syntax(ref msg) => {
                        let parts: Vec<&str> = msg.split("at position").collect();
                        if parts.len() > 1 {
                            parts[1].trim().to_string()
                        } else {
                            "未知位置".to_string()
                        }
                    }
                    _ => "未知位置".to_string(),
                };
                Err((e.to_string(), pos))
            }
        }
    }

    pub async fn add_route(
        &self,
        pattern: String,
        regex: Regex,
        targets: Vec<(String, u32)>,
    ) {
        let mut guard = self.inner.lock().await;

        let backends: Vec<BackendInner> = targets
            .into_iter()
            .map(|(addr, weight)| BackendInner::new(addr, weight))
            .collect();

        guard
            .routes
            .insert(pattern.clone(), Route { _pattern: pattern.clone(), regex });

        if let Some(existing) = guard.backends_by_route.get_mut(&pattern) {
            for new_backend in backends {
                if let Some(existing_backend) =
                    existing.iter_mut().find(|b| b.address == new_backend.address)
                {
                    existing_backend.original_weight = new_backend.original_weight;
                    existing_backend.current_weight = new_backend.current_weight;
                } else {
                    existing.push(new_backend);
                }
            }
        } else {
            guard.backends_by_route.insert(pattern, backends);
        }
    }

    pub async fn find_matching_route(
        &self,
        path: &str,
    ) -> Option<(String, Vec<(String, u32, bool, bool, u64)>)> {
        let guard = self.inner.lock().await;
        for (pattern, route) in guard.routes.iter() {
            if route.regex.is_match(path) {
                if let Some(backends) = guard.backends_by_route.get(pattern) {
                    let info: Vec<(String, u32, bool, bool, u64)> = backends
                        .iter()
                        .map(|b| {
                            (
                                b.address.clone(),
                                b.current_weight,
                                b.is_healthy,
                                b.marked_for_deletion,
                                b.pending_requests,
                            )
                        })
                        .collect();
                    return Some((pattern.clone(), info));
                }
            }
        }
        None
    }

    pub async fn get_captures(&self, pattern: &str, path: &str) -> Option<HashMap<String, String>> {
        let guard = self.inner.lock().await;
        guard.routes.get(pattern).and_then(|route| {
            route.regex.captures(path).map(|caps| {
                let mut result = HashMap::new();
                for (i, _name) in route.regex.capture_names().enumerate() {
                    if let Some(cap) = caps.get(i) {
                        let key = if i == 0 {
                            continue;
                        } else {
                            format!("group{}", i)
                        };
                        result.insert(key, cap.as_str().to_string());
                    }
                }
                result
            })
        })
    }

    pub async fn mark_for_deletion(&self, pattern: &str, address: &str) -> bool {
        let mut guard = self.inner.lock().await;
        if let Some(backends) = guard.backends_by_route.get_mut(pattern) {
            for backend in backends.iter_mut() {
                if backend.address == address {
                    backend.marked_for_deletion = true;
                    return true;
                }
            }
        }
        false
    }

    pub async fn remove_marked_backends(&self) {
        let mut guard = self.inner.lock().await;
        for backends in guard.backends_by_route.values_mut() {
            backends.retain(|b| !b.marked_for_deletion || b.pending_requests > 0);
        }
        let empty_patterns: Vec<String> = guard
            .backends_by_route
            .iter()
            .filter(|(_, v)| v.is_empty())
            .map(|(k, _)| k.clone())
            .collect();
        for p in empty_patterns {
            guard.backends_by_route.remove(&p);
            guard.routes.remove(&p);
        }
    }

    pub async fn increment_request(&self, pattern: &str, address: &str) {
        let mut guard = self.inner.lock().await;
        if let Some(backends) = guard.backends_by_route.get_mut(pattern) {
            for backend in backends.iter_mut() {
                if backend.address == address {
                    backend.total_requests += 1;
                    backend.pending_requests += 1;
                    break;
                }
            }
        }
    }

    pub async fn decrement_pending(&self, pattern: &str, address: &str) {
        let mut guard = self.inner.lock().await;
        if let Some(backends) = guard.backends_by_route.get_mut(pattern) {
            for backend in backends.iter_mut() {
                if backend.address == address {
                    if backend.pending_requests > 0 {
                        backend.pending_requests -= 1;
                    }
                    break;
                }
            }
        }
    }

    pub async fn get_all_backends(&self) -> Vec<(String, Vec<(String, u32, u32, bool, u64, u64, bool)>)>
    {
        let guard = self.inner.lock().await;
        guard
            .backends_by_route
            .iter()
            .map(|(pattern, backends)| {
                let backend_info: Vec<_> = backends
                    .iter()
                    .map(|b| {
                        (
                            b.address.clone(),
                            b.original_weight,
                            b.current_weight,
                            b.is_healthy,
                            b.total_requests,
                            b.pending_requests,
                            b.marked_for_deletion,
                        )
                    })
                    .collect();
                (pattern.clone(), backend_info)
            })
            .collect()
    }

    pub async fn get_all_status(&self) -> Vec<RouteStatus> {
        let guard = self.inner.lock().await;
        guard
            .backends_by_route
            .iter()
            .map(|(pattern, backends)| RouteStatus {
                pattern: pattern.clone(),
                backends: backends
                    .iter()
                    .map(|b| BackendStatus {
                        address: b.address.clone(),
                        original_weight: b.original_weight,
                        current_weight: b.current_weight,
                        is_healthy: b.is_healthy,
                        total_requests: b.total_requests,
                        pending_requests: b.pending_requests,
                        marked_for_deletion: b.marked_for_deletion,
                    })
                    .collect(),
            })
            .collect()
    }

    pub async fn update_health<F>(&self, pattern: &str, address: &str, updater: F)
    where
        F: FnOnce(&mut BackendInner),
    {
        let mut guard = self.inner.lock().await;
        if let Some(backends) = guard.backends_by_route.get_mut(pattern) {
            for backend in backends.iter_mut() {
                if backend.address == address {
                    updater(backend);
                    break;
                }
            }
        }
    }
}
