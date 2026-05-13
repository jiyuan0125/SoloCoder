use crate::errors::ConfigError;
use crate::models::{ConfigChange, ConfigValue, ValueType};

pub fn validate_type(value_type: &ValueType, value: &serde_json::Value) -> Result<(), ConfigError> {
    match value_type {
        ValueType::String => {
            if !value.is_string() {
                return Err(ConfigError::TypeValidationError(
                    format!("Expected string type, got {:?}", value)
                ));
            }
        }
        ValueType::Number => {
            if !value.is_number() {
                return Err(ConfigError::TypeValidationError(
                    format!("Expected number type, got {:?}", value)
                ));
            }
        }
        ValueType::Json => {
            if value.is_null() || value.is_boolean() || value.is_string() || value.is_number() {
                return Err(ConfigError::TypeValidationError(
                    format!("Expected JSON object/array type, got {:?}", value)
                ));
            }
        }
    }
    Ok(())
}

pub fn validate_config_changes(changes: &[ConfigChange]) -> Result<(), ConfigError> {
    for change in changes {
        validate_type(&change.value_type, &change.value)?;
    }
    Ok(())
}

pub fn to_config_value(change: &ConfigChange) -> ConfigValue {
    ConfigValue {
        value_type: change.value_type.clone(),
        value: change.value.clone(),
    }
}
