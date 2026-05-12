use std::collections::HashMap;
use uuid::Uuid;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ChannelStatus {
    Unconfigured,
    Ready,
    Transforming,
    Error,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ChannelConfig {
    pub service: String,
    pub method: String,
    pub backend_address: String,
    pub proto: String,
}

#[derive(Debug, Clone, Serialize)]
pub struct ChannelStats {
    pub total_requests: u64,
    pub success_requests: u64,
    pub failed_requests: u64,
    pub success_rate: f64,
}

#[derive(Debug, Clone, Serialize)]
pub struct Channel {
    pub id: Uuid,
    pub status: ChannelStatus,
    pub config: Option<ChannelConfig>,
    pub stats: ChannelStats,
    pub error_message: Option<String>,
}

#[derive(Debug, Clone)]
pub struct ChannelManager {
    channels: HashMap<Uuid, Channel>,
    mappings: HashMap<(String, String), Uuid>,
}

impl ChannelManager {
    pub fn new() -> Self {
        ChannelManager {
            channels: HashMap::new(),
            mappings: HashMap::new(),
        }
    }

    pub fn create(&mut self, config: Option<ChannelConfig>) -> Channel {
        let id = Uuid::new_v4();
        let status = if config.is_some() {
            ChannelStatus::Ready
        } else {
            ChannelStatus::Unconfigured
        };

        let channel = Channel {
            id,
            status,
            config: config.clone(),
            stats: ChannelStats {
                total_requests: 0,
                success_requests: 0,
                failed_requests: 0,
                success_rate: 0.0,
            },
            error_message: None,
        };

        if let Some(cfg) = &config {
            self.mappings.insert((cfg.service.clone(), cfg.method.clone()), id);
        }

        self.channels.insert(id, channel.clone());
        channel
    }

    pub fn list(&self) -> Vec<Channel> {
        self.channels.values().cloned().collect()
    }

    pub fn get(&self, id: &Uuid) -> Option<Channel> {
        self.channels.get(id).cloned()
    }

    pub fn configure(&mut self, id: &Uuid, config: ChannelConfig) -> Result<Channel, String> {
        let channel = self.channels.get_mut(id).ok_or_else(|| "Channel not found".to_string())?;

        if let Some(old_cfg) = &channel.config {
            self.mappings.remove(&(old_cfg.service.clone(), old_cfg.method.clone()));
        }

        self.mappings.insert((config.service.clone(), config.method.clone()), *id);

        channel.config = Some(config);
        channel.status = ChannelStatus::Ready;
        channel.error_message = None;

        Ok(channel.clone())
    }

    pub fn reset(&mut self, id: &Uuid) -> Result<Channel, String> {
        let channel = self.channels.get_mut(id).ok_or_else(|| "Channel not found".to_string())?;

        channel.status = if channel.config.is_some() {
            ChannelStatus::Ready
        } else {
            ChannelStatus::Unconfigured
        };
        channel.error_message = None;

        Ok(channel.clone())
    }

    pub fn get_by_service_method(&self, service: &str, method: &str) -> Option<Channel> {
        self.mappings
            .get(&(service.to_string(), method.to_string()))
            .and_then(|id| self.channels.get(id).cloned())
    }

    pub fn start_transforming(&mut self, id: &Uuid) -> Result<Channel, String> {
        let channel = self.channels.get_mut(id).ok_or_else(|| "Channel not found".to_string())?;

        if channel.status != ChannelStatus::Ready {
            return Err("Channel is not ready".to_string());
        }

        channel.status = ChannelStatus::Transforming;
        Ok(channel.clone())
    }

    pub fn complete_success(&mut self, id: &Uuid) -> Result<Channel, String> {
        let channel = self.channels.get_mut(id).ok_or_else(|| "Channel not found".to_string())?;

        channel.stats.total_requests += 1;
        channel.stats.success_requests += 1;
        channel.stats.success_rate = if channel.stats.total_requests > 0 {
            channel.stats.success_requests as f64 / channel.stats.total_requests as f64
        } else {
            0.0
        };
        channel.status = ChannelStatus::Ready;
        channel.error_message = None;

        Ok(channel.clone())
    }

    pub fn complete_error(&mut self, id: &Uuid, error_message: String) -> Result<Channel, String> {
        let channel = self.channels.get_mut(id).ok_or_else(|| "Channel not found".to_string())?;

        channel.stats.total_requests += 1;
        channel.stats.failed_requests += 1;
        channel.stats.success_rate = if channel.stats.total_requests > 0 {
            channel.stats.success_requests as f64 / channel.stats.total_requests as f64
        } else {
            0.0
        };
        channel.status = ChannelStatus::Error;
        channel.error_message = Some(error_message);

        Ok(channel.clone())
    }
}
