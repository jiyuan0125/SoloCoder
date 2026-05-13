use aes_gcm::aead::{Aead, KeyInit};
use aes_gcm::{Aes256Gcm, Nonce};
use base64::{engine::general_purpose, Engine as _};
use chrono::{DateTime, Utc};
use rand::RngCore;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::path::PathBuf;
use std::sync::{Arc, Mutex};

use crate::models::{
    AuditLog, ConfigItem, ConfigItemMeta, TransactionLog, TransactionLogEntry, TransactionStatus,
};

const CONFIG_FILE: &str = "config_store.json";
const TX_LOG_FILE: &str = "tx_logs.json";
const AUDIT_LOG_FILE: &str = "audit_logs.json";
const MAX_AUDIT_LOGS: usize = 2000;
const MASKED_VALUE: &str = "***";

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct StoredConfigItem {
    pub key: String,
    pub value: String,
    pub is_secret: bool,
    pub updated_at: DateTime<Utc>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PersistentState {
    pub configs: HashMap<String, StoredConfigItem>,
    pub tx_logs: Vec<TransactionLog>,
    pub audit_logs: Vec<AuditLog>,
    pub next_tx_id: u64,
    pub next_audit_id: u64,
}

impl Default for PersistentState {
    fn default() -> Self {
        Self {
            configs: HashMap::new(),
            tx_logs: Vec::new(),
            audit_logs: Vec::new(),
            next_tx_id: 1,
            next_audit_id: 1,
        }
    }
}

pub struct EncryptionService {
    cipher: Aes256Gcm,
}

impl EncryptionService {
    pub fn new(key: &[u8; 32]) -> Self {
        Self {
            cipher: Aes256Gcm::new(key.into()),
        }
    }

    pub fn encrypt(&self, plaintext: &str) -> Result<String, String> {
        let mut nonce_bytes = [0u8; 12];
        rand::thread_rng().fill_bytes(&mut nonce_bytes);
        let nonce = Nonce::from_slice(&nonce_bytes);

        let ciphertext = self
            .cipher
            .encrypt(nonce, plaintext.as_bytes())
            .map_err(|e| format!("加密失败: {}", e))?;

        let mut combined = Vec::with_capacity(12 + ciphertext.len());
        combined.extend_from_slice(&nonce_bytes);
        combined.extend_from_slice(&ciphertext);

        Ok(general_purpose::STANDARD.encode(&combined))
    }

    pub fn decrypt(&self, ciphertext_b64: &str) -> Result<String, String> {
        let combined = general_purpose::STANDARD
            .decode(ciphertext_b64)
            .map_err(|e| format!("Base64 解码失败: {}", e))?;

        if combined.len() < 12 {
            return Err("密文格式错误".to_string());
        }

        let (nonce_bytes, ciphertext) = combined.split_at(12);
        let nonce = Nonce::from_slice(nonce_bytes);

        let plaintext = self
            .cipher
            .decrypt(nonce, ciphertext)
            .map_err(|e| format!("解密失败: {}", e))?;

        String::from_utf8(plaintext).map_err(|e| format!("UTF-8 转换失败: {}", e))
    }
}

pub struct ConfigStore {
    state: Arc<Mutex<PersistentState>>,
    encryption: EncryptionService,
    data_dir: PathBuf,
}

impl ConfigStore {
    pub fn new(data_dir: PathBuf, encryption_key: [u8; 32]) -> Self {
        let encryption = EncryptionService::new(&encryption_key);
        let state = Arc::new(Mutex::new(PersistentState::default()));
        let store = Self {
            state,
            encryption,
            data_dir,
        };
        store.load_or_initialize();
        store.recover_pending_transactions();
        store
    }

    fn data_file(&self, name: &str) -> PathBuf {
        self.data_dir.join(name)
    }

    fn load_or_initialize(&self) {
        fs::create_dir_all(&self.data_dir).ok();
        
        let config_path = self.data_file(CONFIG_FILE);
        let tx_log_path = self.data_file(TX_LOG_FILE);
        let audit_path = self.data_file(AUDIT_LOG_FILE);

        let mut state = self.state.lock().unwrap();

        if config_path.exists() {
            if let Ok(contents) = fs::read_to_string(&config_path) {
                if let Ok(configs) = serde_json::from_str::<HashMap<String, StoredConfigItem>>(&contents) {
                    state.configs = configs;
                }
            }
        }

        if tx_log_path.exists() {
            if let Ok(contents) = fs::read_to_string(&tx_log_path) {
                if let Ok(tx_logs) = serde_json::from_str::<Vec<TransactionLog>>(&contents) {
                    state.tx_logs = tx_logs;
                }
            }
        }

        if audit_path.exists() {
            if let Ok(contents) = fs::read_to_string(&audit_path) {
                if let Ok(audit_logs) = serde_json::from_str::<Vec<AuditLog>>(&contents) {
                    state.audit_logs = audit_logs;
                    if let Some(max_id) = state.audit_logs.iter().map(|l| l.id).max() {
                        state.next_audit_id = max_id + 1;
                    }
                }
            }
        }

        if let Some(max_tx_id) = state.tx_logs.iter().map(|l| l.id).max() {
            state.next_tx_id = max_tx_id + 1;
        }
    }

    fn save_configs(&self, state: &PersistentState) {
        let path = self.data_file(CONFIG_FILE);
        let contents = serde_json::to_string_pretty(&state.configs).unwrap();
        fs::write(path, contents).ok();
    }

    fn save_tx_logs(&self, state: &PersistentState) {
        let path = self.data_file(TX_LOG_FILE);
        let contents = serde_json::to_string_pretty(&state.tx_logs).unwrap();
        fs::write(path, contents).ok();
    }

    fn save_audit_logs(&self, state: &PersistentState) {
        let path = self.data_file(AUDIT_LOG_FILE);
        let contents = serde_json::to_string_pretty(&state.audit_logs).unwrap();
        fs::write(path, contents).ok();
    }

    fn recover_pending_transactions(&self) {
        let mut state = self.state.lock().unwrap();
        let pending_tx_ids: Vec<u64> = state
            .tx_logs
            .iter()
            .filter(|tx| matches!(tx.status, TransactionStatus::Pending))
            .map(|tx| tx.id)
            .collect();

        for tx_id in pending_tx_ids {
            self.rollback_transaction_internal(&mut state, tx_id);
        }

        self.save_configs(&state);
        self.save_tx_logs(&state);
    }

    fn create_transaction_internal(
        &self,
        state: &mut PersistentState,
        entries: Vec<TransactionLogEntry>,
    ) -> u64 {
        let tx_id = state.next_tx_id;
        state.next_tx_id += 1;

        let tx_log = TransactionLog {
            id: tx_id,
            status: TransactionStatus::Pending,
            entries,
            created_at: Utc::now(),
        };

        state.tx_logs.push(tx_log);
        self.save_tx_logs(state);
        tx_id
    }

    fn commit_transaction_internal(
        &self,
        state: &mut PersistentState,
        tx_id: u64,
        user_id: &str,
    ) -> Result<(), String> {
        let tx_index = state
            .tx_logs
            .iter()
            .position(|t| t.id == tx_id)
            .ok_or_else(|| format!("事务 {} 不存在", tx_id))?;

        if !matches!(state.tx_logs[tx_index].status, TransactionStatus::Pending) {
            return Err(format!("事务 {} 不是待提交状态", tx_id));
        }

        let tx_entries = state.tx_logs[tx_index].entries.clone();

        for entry in &tx_entries {
            let old_item = state.configs.get(&entry.key).cloned();
            let old_value = old_item.as_ref().map(|i| i.value.clone());
            let new_value = if entry.is_secret {
                self.encryption.encrypt(&entry.new_value)?
            } else {
                entry.new_value.clone()
            };

            let new_item = StoredConfigItem {
                key: entry.key.clone(),
                value: new_value,
                is_secret: entry.is_secret,
                updated_at: Utc::now(),
            };

            let operation = if old_item.is_some() { "UPDATE" } else { "CREATE" };
            self.add_audit_log_internal(
                state,
                user_id,
                &entry.key,
                operation,
                if entry.is_secret {
                    old_value.map(|_| MASKED_VALUE.to_string())
                } else {
                    old_value
                },
                if entry.is_secret {
                    Some(MASKED_VALUE.to_string())
                } else {
                    Some(entry.new_value.clone())
                },
            );

            state.configs.insert(entry.key.clone(), new_item);
        }

        state.tx_logs[tx_index].status = TransactionStatus::Committed;
        self.save_configs(state);
        self.save_tx_logs(state);
        self.save_audit_logs(state);
        Ok(())
    }

    fn rollback_transaction_internal(&self, state: &mut PersistentState, tx_id: u64) {
        let tx = match state.tx_logs.iter_mut().find(|t| t.id == tx_id) {
            Some(tx) => tx,
            None => return,
        };

        if !matches!(tx.status, TransactionStatus::Pending) {
            return;
        }

        for entry in &tx.entries {
            if let Some(old_value) = &entry.old_value {
                state.configs.insert(
                    entry.key.clone(),
                    StoredConfigItem {
                        key: entry.key.clone(),
                        value: old_value.clone(),
                        is_secret: entry.is_secret,
                        updated_at: Utc::now(),
                    },
                );
            } else {
                state.configs.remove(&entry.key);
            }
        }

        tx.status = TransactionStatus::RolledBack;
    }

    fn add_audit_log_internal(
        &self,
        state: &mut PersistentState,
        user_id: &str,
        key: &str,
        operation: &str,
        old_value: Option<String>,
        new_value: Option<String>,
    ) {
        let log = AuditLog {
            id: state.next_audit_id,
            user_id: user_id.to_string(),
            key: key.to_string(),
            operation: operation.to_string(),
            old_value,
            new_value,
            timestamp: Utc::now(),
        };

        state.next_audit_id += 1;
        state.audit_logs.push(log);

        if state.audit_logs.len() > MAX_AUDIT_LOGS {
            let excess = state.audit_logs.len() - MAX_AUDIT_LOGS;
            state.audit_logs.drain(0..excess);
        }
    }

    pub fn batch_update(
        &self,
        items: &HashMap<String, (String, bool)>,
        user_id: &str,
    ) -> Result<(), String> {
        let mut state = self.state.lock().unwrap();

        let mut entries = Vec::with_capacity(items.len());
        for (key, (value, is_secret)) in items {
            let old_item = state.configs.get(key).cloned();
            let old_value = old_item.as_ref().map(|i| i.value.clone());

            entries.push(TransactionLogEntry {
                key: key.clone(),
                old_value,
                new_value: value.clone(),
                is_secret: *is_secret,
            });
        }

        let tx_id = self.create_transaction_internal(&mut state, entries);

        match self.commit_transaction_internal(&mut state, tx_id, user_id) {
            Ok(_) => Ok(()),
            Err(e) => {
                self.rollback_transaction_internal(&mut state, tx_id);
                self.save_tx_logs(&state);
                Err(e)
            }
        }
    }

    pub fn single_update(
        &self,
        key: &str,
        value: &str,
        is_secret: bool,
        user_id: &str,
    ) -> Result<(), String> {
        let mut items = HashMap::new();
        items.insert(key.to_string(), (value.to_string(), is_secret));
        self.batch_update(&items, user_id)
    }

    pub fn delete(&self, key: &str, user_id: &str) -> Result<bool, String> {
        let mut state = self.state.lock().unwrap();

        let old_item = match state.configs.remove(key) {
            Some(item) => item,
            None => return Ok(false),
        };

        let old_value = if old_item.is_secret {
            Some(MASKED_VALUE.to_string())
        } else {
            Some(old_item.value.clone())
        };

        self.add_audit_log_internal(&mut state, user_id, key, "DELETE", old_value, None);

        self.save_configs(&state);
        self.save_audit_logs(&state);
        Ok(true)
    }

    pub fn list(&self) -> Vec<ConfigItemMeta> {
        let state = self.state.lock().unwrap();
        state
            .configs
            .values()
            .map(|item| ConfigItemMeta {
                key: item.key.clone(),
                is_secret: item.is_secret,
                updated_at: item.updated_at,
            })
            .collect()
    }

    pub fn get(&self, key: &str) -> Result<Option<ConfigItem>, String> {
        let state = self.state.lock().unwrap();
        match state.configs.get(key) {
            Some(stored) => {
                let value = if stored.is_secret {
                    self.encryption.decrypt(&stored.value)?
                } else {
                    stored.value.clone()
                };

                Ok(Some(ConfigItem {
                    key: stored.key.clone(),
                    value,
                    is_secret: stored.is_secret,
                    updated_at: stored.updated_at,
                }))
            }
            None => Ok(None),
        }
    }

    pub fn query_audit_logs(
        &self,
        user_id: Option<&str>,
        start_time: Option<DateTime<Utc>>,
        end_time: Option<DateTime<Utc>>,
    ) -> Vec<AuditLog> {
        let state = self.state.lock().unwrap();
        state
            .audit_logs
            .iter()
            .filter(|log| {
                if let Some(uid) = user_id {
                    if log.user_id != uid {
                        return false;
                    }
                }
                if let Some(start) = start_time {
                    if log.timestamp < start {
                        return false;
                    }
                }
                if let Some(end) = end_time {
                    if log.timestamp > end {
                        return false;
                    }
                }
                true
            })
            .cloned()
            .collect()
    }
}
