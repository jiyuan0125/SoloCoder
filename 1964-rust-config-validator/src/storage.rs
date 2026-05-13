use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use uuid::Uuid;
use chrono::{DateTime, Utc};
use crate::models::{ConfigVersion, DeployRecord, DeployStatus, ServiceMapping, ServiceInfo};

const MAX_VERSION_HISTORY: usize = 50;

#[derive(Clone, Default)]
pub struct ConfigStore {
    inner: Arc<Mutex<InnerStore>>,
}

#[derive(Default)]
struct InnerStore {
    versions: HashMap<String, Vec<ConfigVersion>>,
    full_versions: HashMap<String, Uuid>,
    canary_versions: HashMap<String, Uuid>,
    deploy_records: Vec<DeployRecord>,
    service_mappings: HashMap<String, ServiceMapping>,
}

impl ConfigStore {
    pub fn new() -> Self {
        Self {
            inner: Arc::new(Mutex::new(InnerStore::default())),
        }
    }
    
    pub fn add_version(&self, version: ConfigVersion) {
        let mut inner = self.inner.lock().unwrap();
        let versions = inner.versions.entry(version.key.clone()).or_insert_with(Vec::new);
        versions.push(version);
        
        if versions.len() > MAX_VERSION_HISTORY {
            versions.remove(0);
        }
    }
    
    pub fn get_version(&self, key: &str, version_id: Uuid) -> Option<ConfigVersion> {
        let inner = self.inner.lock().unwrap();
        inner.versions.get(key).and_then(|versions| {
            versions.iter().find(|v| v.version == version_id).cloned()
        })
    }
    
    pub fn get_versions(&self, key: &str) -> Vec<ConfigVersion> {
        let inner = self.inner.lock().unwrap();
        inner.versions.get(key).cloned().unwrap_or_default()
    }
    
    pub fn get_latest_version(&self, key: &str) -> Option<ConfigVersion> {
        let inner = self.inner.lock().unwrap();
        inner.versions.get(key).and_then(|versions| versions.last().cloned())
    }
    
    pub fn get_full_version(&self, key: &str) -> Option<ConfigVersion> {
        let inner = self.inner.lock().unwrap();
        inner.full_versions.get(key).and_then(|version_id| {
            inner.versions.get(key).and_then(|versions| {
                versions.iter().find(|v| v.version == *version_id).cloned()
            })
        })
    }
    
    pub fn get_canary_version(&self, key: &str) -> Option<ConfigVersion> {
        let inner = self.inner.lock().unwrap();
        inner.canary_versions.get(key).and_then(|version_id| {
            inner.versions.get(key).and_then(|versions| {
                versions.iter().find(|v| v.version == *version_id).cloned()
            })
        })
    }
    
    pub fn set_full_version(&self, key: &str, version_id: Uuid) {
        let mut inner = self.inner.lock().unwrap();
        inner.full_versions.insert(key.to_string(), version_id);
    }
    
    pub fn set_canary_version(&self, key: &str, version_id: Uuid) {
        let mut inner = self.inner.lock().unwrap();
        inner.canary_versions.insert(key.to_string(), version_id);
    }
    
    pub fn clear_canary_version(&self, key: &str) {
        let mut inner = self.inner.lock().unwrap();
        inner.canary_versions.remove(key);
    }
    
    pub fn add_deploy_record(&self, record: DeployRecord) {
        let mut inner = self.inner.lock().unwrap();
        inner.deploy_records.push(record);
    }
    
    pub fn update_deploy_status(&self, deploy_id: Uuid, status: DeployStatus) {
        let mut inner = self.inner.lock().unwrap();
        if let Some(record) = inner.deploy_records.iter_mut().find(|r| r.id == deploy_id) {
            record.status = status.clone();
            if status == DeployStatus::Success || status == DeployStatus::Failed || status == DeployStatus::RolledBack {
                record.completed_at = Some(Utc::now());
            }
        }
    }
    
    pub fn get_deploy_records(&self, key: &str) -> Vec<DeployRecord> {
        let inner = self.inner.lock().unwrap();
        inner.deploy_records
            .iter()
            .filter(|r| r.key == key)
            .cloned()
            .collect()
    }
    
    pub fn get_service_mapping(&self, key: &str) -> Option<ServiceMapping> {
        let inner = self.inner.lock().unwrap();
        inner.service_mappings.get(key).cloned()
    }
    
    pub fn set_service_mapping(&self, mapping: ServiceMapping) {
        let mut inner = self.inner.lock().unwrap();
        inner.service_mappings.insert(mapping.key.clone(), mapping);
    }
    
    pub fn init_sample_data(&self) {
        self.set_service_mapping(ServiceMapping {
            key: "db.connection.timeout".to_string(),
            services: vec![
                ServiceInfo { name: "api-gateway".to_string(), instance_count: 10, environment: "production".to_string() },
                ServiceInfo { name: "user-service".to_string(), instance_count: 5, environment: "production".to_string() },
                ServiceInfo { name: "order-service".to_string(), instance_count: 8, environment: "production".to_string() },
            ],
        });
        
        self.set_service_mapping(ServiceMapping {
            key: "feature.flag.new-ui".to_string(),
            services: vec![
                ServiceInfo { name: "web-frontend".to_string(), instance_count: 20, environment: "production".to_string() },
            ],
        });
        
        self.set_service_mapping(ServiceMapping {
            key: "cache.ttl".to_string(),
            services: vec![
                ServiceInfo { name: "api-gateway".to_string(), instance_count: 10, environment: "production".to_string() },
                ServiceInfo { name: "auth-service".to_string(), instance_count: 3, environment: "production".to_string() },
            ],
        });
    }
}
