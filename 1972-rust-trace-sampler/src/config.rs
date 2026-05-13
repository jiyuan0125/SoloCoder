use crate::models::SamplerConfig;
use parking_lot::RwLock;

pub struct ConfigManager {
    config: RwLock<SamplerConfig>,
}

impl ConfigManager {
    pub fn new(initial: SamplerConfig) -> Self {
        ConfigManager {
            config: RwLock::new(initial),
        }
    }

    pub fn get_config(&self) -> SamplerConfig {
        self.config.read().clone()
    }

    pub fn set_config(&self, new_config: SamplerConfig) -> Result<(), String> {
        self.validate(&new_config)?;
        *self.config.write() = new_config;
        Ok(())
    }

    fn validate(&self, config: &SamplerConfig) -> Result<(), String> {
        if config.fixed_rate.rate < 0.0 || config.fixed_rate.rate > 1.0 {
            return Err("fixed_rate.rate must be between 0 and 1".to_string());
        }
        if config.adaptive.min_rate < 0.0 || config.adaptive.min_rate > 1.0 {
            return Err("adaptive.min_rate must be between 0 and 1".to_string());
        }
        if config.adaptive.max_rate < 0.0 || config.adaptive.max_rate > 1.0 {
            return Err("adaptive.max_rate must be between 0 and 1".to_string());
        }
        if config.adaptive.min_rate > config.adaptive.max_rate {
            return Err("adaptive.min_rate must be <= adaptive.max_rate".to_string());
        }
        if config.error_based.base_rate < 0.0 || config.error_based.base_rate > 1.0 {
            return Err("error_based.base_rate must be between 0 and 1".to_string());
        }
        if config.error_based.error_rate_threshold < 0.0 || config.error_based.error_rate_threshold > 1.0 {
            return Err("error_based.error_rate_threshold must be between 0 and 1".to_string());
        }
        if config.error_based.max_multiplier < 1.0 {
            return Err("error_based.max_multiplier must be >= 1".to_string());
        }
        Ok(())
    }
}
