use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SystemConfig {
    pub reminder_config: crate::ReminderConfig,
}

impl Default for SystemConfig {
    fn default() -> Self {
        SystemConfig {
            reminder_config: crate::ReminderConfig::default(),
        }
    }
}
