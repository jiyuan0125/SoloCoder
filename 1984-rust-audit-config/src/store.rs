use std::collections::{BTreeMap, HashMap};
use std::sync::Arc;

use parking_lot::RwLock;
use time::OffsetDateTime;
use uuid::Uuid;

use crate::errors::ConfigError;
use crate::models::{AuditLog, ChangeRequest, Configuration, DiffItem, NamespaceConfig};

#[derive(Clone)]
pub struct MemoryStore {
    inner: Arc<RwLock<StoreInner>>,
}

struct StoreInner {
    namespaces: HashMap<String, NamespaceConfig>,
    current_configs: HashMap<String, HashMap<String, Configuration>>,
    approved_versions: BTreeMap<u64, HashMap<String, HashMap<String, Configuration>>>,
    version_counter: u64,
    change_requests: HashMap<Uuid, ChangeRequest>,
    audit_logs: Vec<AuditLog>,
}

impl MemoryStore {
    pub fn new() -> Self {
        let mut inner = StoreInner {
            namespaces: HashMap::new(),
            current_configs: HashMap::new(),
            approved_versions: BTreeMap::new(),
            version_counter: 0,
            change_requests: HashMap::new(),
            audit_logs: Vec::new(),
        };
        
        let default_ns = NamespaceConfig {
            name: "default".to_string(),
            approvers: vec!["admin1".to_string(), "admin2".to_string()],
        };
        inner.namespaces.insert("default".to_string(), default_ns);
        inner.current_configs.insert("default".to_string(), HashMap::new());
        
        MemoryStore {
            inner: Arc::new(RwLock::new(inner)),
        }
    }

    pub fn add_namespace(&self, namespace: NamespaceConfig) {
        let name = namespace.name.clone();
        let mut inner = self.inner.write();
        if !inner.current_configs.contains_key(&name) {
            inner.current_configs.insert(name.clone(), HashMap::new());
        }
        inner.namespaces.insert(name, namespace);
    }

    pub fn get_namespace(&self, name: &str) -> Option<NamespaceConfig> {
        let inner = self.inner.read();
        inner.namespaces.get(name).cloned()
    }

    pub fn is_approver(&self, namespace: &str, user: &str) -> bool {
        let inner = self.inner.read();
        inner.namespaces
            .get(namespace)
            .map(|ns| ns.approvers.contains(&user.to_string()))
            .unwrap_or(false)
    }

    pub fn get_current_configs(&self, namespace: &str) -> Option<HashMap<String, Configuration>> {
        let inner = self.inner.read();
        inner.current_configs.get(namespace).cloned()
    }

    pub fn snapshot_current_as_approved(&self, operator: &str) -> u64 {
        let mut inner = self.inner.write();
        let version = inner.version_counter + 1;
        let snapshot: HashMap<String, HashMap<String, Configuration>> = inner.current_configs
            .iter()
            .map(|(ns, configs)| (ns.clone(), configs.clone()))
            .collect();
        inner.approved_versions.insert(version, snapshot);
        inner.version_counter = version;
        
        for (_ns, configs) in inner.current_configs.iter_mut() {
            for config in configs.values_mut() {
                config.version = version;
                config.updated_by = operator.to_string();
                config.updated_at = OffsetDateTime::now_utc();
            }
        }
        
        version
    }

    pub fn get_last_approved_version(&self) -> Option<u64> {
        let inner = self.inner.read();
        inner.approved_versions.keys().last().copied()
    }

    pub fn rollback_to_version(&self, version: u64) -> Result<Vec<DiffItem>, ConfigError> {
        let mut inner = self.inner.write();
        let snapshot = inner.approved_versions
            .get(&version)
            .ok_or_else(|| ConfigError::Internal(format!("Version {} not found", version)))?
            .clone();
        
        let mut diffs = Vec::new();
        for (ns, configs) in snapshot.iter() {
            let current_ns = inner.current_configs
                .entry(ns.clone())
                .or_insert_with(HashMap::new);
            
            for (key, old_config) in configs {
                let before = current_ns.get(key).map(|c| c.config.clone());
                let after = old_config.config.clone();
                if before != Some(after.clone()) {
                    diffs.push(DiffItem {
                        key: format!("{}/{}", ns, key),
                        before,
                        after: Some(after),
                    });
                }
                current_ns.insert(key.clone(), old_config.clone());
            }
        }
        
        let all_current_keys: Vec<(String, String)> = inner.current_configs
            .iter()
            .flat_map(|(ns, configs)| {
                configs.keys().map(|k| (ns.clone(), k.clone())).collect::<Vec<_>>()
            })
            .collect();
        
        for (ns, key) in all_current_keys {
            if !snapshot.get(&ns).map(|c| c.contains_key(&key)).unwrap_or(false) {
                let current_ns = inner.current_configs.get_mut(&ns).unwrap();
                if let Some(removed) = current_ns.remove(&key) {
                    diffs.push(DiffItem {
                        key: format!("{}/{}", ns, key),
                        before: Some(removed.config),
                        after: None,
                    });
                }
            }
        }
        
        Ok(diffs)
    }

    pub fn insert_change_request(&self, request: ChangeRequest) {
        let mut inner = self.inner.write();
        inner.change_requests.insert(request.id, request);
    }

    pub fn get_change_request(&self, id: &Uuid) -> Option<ChangeRequest> {
        let inner = self.inner.read();
        inner.change_requests.get(id).cloned()
    }

    pub fn update_change_request(&self, id: &Uuid, request: ChangeRequest) {
        let mut inner = self.inner.write();
        inner.change_requests.insert(*id, request);
    }

    pub fn apply_changes_to_current(&self, namespace: &str, diffs: &[DiffItem]) -> Vec<DiffItem> {
        let mut inner = self.inner.write();
        let current_version = inner.version_counter;
        let current_ns = inner.current_configs
            .entry(namespace.to_string())
            .or_insert_with(HashMap::new);
        
        let mut applied_diffs = Vec::new();
        
        for diff in diffs {
            let parts: Vec<&str> = diff.key.splitn(2, '/').collect();
            let key = if parts.len() == 2 { parts[1] } else { &diff.key };
            
            if let Some(after) = &diff.after {
                let new_config = Configuration {
                    namespace: namespace.to_string(),
                    key: key.to_string(),
                    config: after.clone(),
                    version: current_version,
                    updated_at: OffsetDateTime::now_utc(),
                    updated_by: "system".to_string(),
                };
                let before = current_ns.get(key).map(|c| c.config.clone());
                if before != Some(after.clone()) {
                    applied_diffs.push(DiffItem {
                        key: diff.key.clone(),
                        before,
                        after: Some(after.clone()),
                    });
                }
                current_ns.insert(key.to_string(), new_config);
            } else {
                if let Some(removed) = current_ns.remove(key) {
                    applied_diffs.push(DiffItem {
                        key: diff.key.clone(),
                        before: Some(removed.config),
                        after: None,
                    });
                }
            }
        }
        
        applied_diffs
    }

    pub fn record_audit_log(&self, log: AuditLog) {
        let mut inner = self.inner.write();
        inner.audit_logs.push(log);
    }

    pub fn get_audit_logs(&self, namespace: Option<&str>) -> Vec<AuditLog> {
        let inner = self.inner.read();
        let mut logs = inner.audit_logs.clone();
        if let Some(ns) = namespace {
            logs.retain(|log| log.namespace == ns);
        }
        logs.sort_by(|a, b| b.timestamp.cmp(&a.timestamp));
        logs
    }
}
