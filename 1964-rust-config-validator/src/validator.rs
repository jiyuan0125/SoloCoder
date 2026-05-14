use crate::models::{ConfigType, Schema, ValidationError, ValidationResult};
use regex::Regex;
use serde_json::Value;

pub fn validate(value: &Value, schema: &Schema, field_path: &str) -> ValidationResult {
    let mut errors = Vec::new();
    
    if value.is_null() {
        if schema.required.unwrap_or(false) {
            errors.push(ValidationError {
                field: field_path.to_string(),
                expected: "non-null value".to_string(),
                actual: "null".to_string(),
                message: format!("Field '{}' is required but got null", field_path),
            });
        }
        return ValidationResult {
            valid: errors.is_empty(),
            errors,
        };
    }
    
    let type_result = check_type(value, schema, field_path);
    if !type_result.valid {
        errors.extend(type_result.errors);
        return ValidationResult {
            valid: false,
            errors,
        };
    }
    
    match schema.config_type {
        ConfigType::String => validate_string(value, schema, field_path, &mut errors),
        ConfigType::Number => validate_number(value, schema, field_path, &mut errors),
        ConfigType::Boolean => validate_boolean(value, schema, field_path, &mut errors),
        ConfigType::Array => validate_array(value, schema, field_path, &mut errors),
        ConfigType::Object => validate_object(value, schema, field_path, &mut errors),
    }
    
    check_enum(value, schema, field_path, &mut errors);
    
    ValidationResult {
        valid: errors.is_empty(),
        errors,
    }
}

fn check_type(value: &Value, schema: &Schema, field_path: &str) -> ValidationResult {
    let is_type_match = match schema.config_type {
        ConfigType::String => value.is_string(),
        ConfigType::Number => value.is_number(),
        ConfigType::Boolean => value.is_boolean(),
        ConfigType::Array => value.is_array(),
        ConfigType::Object => value.is_object(),
    };
    
    if !is_type_match {
        let actual_type = get_json_type(value);
        let expected_type = format!("{:?}", schema.config_type).to_lowercase();
        return ValidationResult {
            valid: false,
            errors: vec![ValidationError {
                field: field_path.to_string(),
                expected: expected_type.clone(),
                actual: actual_type.clone(),
                message: format!("Field '{}' expected type '{}' but got '{}'", field_path, expected_type, actual_type),
            }],
        };
    }
    
    ValidationResult {
        valid: true,
        errors: Vec::new(),
    }
}

fn validate_string(value: &Value, schema: &Schema, field_path: &str, errors: &mut Vec<ValidationError>) {
    if let Some(s) = value.as_str() {
        if let Some(min_length) = schema.min_length {
            if s.len() < min_length {
                errors.push(ValidationError {
                    field: field_path.to_string(),
                    expected: format!("min_length={}", min_length),
                    actual: format!("length={}", s.len()),
                    message: format!("Field '{}' must be at least {} characters, got {}", field_path, min_length, s.len()),
                });
            }
        }
        
        if let Some(max_length) = schema.max_length {
            if s.len() > max_length {
                errors.push(ValidationError {
                    field: field_path.to_string(),
                    expected: format!("max_length={}", max_length),
                    actual: format!("length={}", s.len()),
                    message: format!("Field '{}' must be at most {} characters, got {}", field_path, max_length, s.len()),
                });
            }
        }
        
        if let Some(pattern) = &schema.pattern {
            if let Ok(re) = Regex::new(pattern) {
                if !re.is_match(s) {
                    errors.push(ValidationError {
                        field: field_path.to_string(),
                        expected: format!("pattern='{}'", pattern),
                        actual: format!("'{}'", s),
                        message: format!("Field '{}' does not match pattern '{}'", field_path, pattern),
                    });
                }
            }
        }
    }
}

fn validate_number(value: &Value, schema: &Schema, field_path: &str, errors: &mut Vec<ValidationError>) {
    if let Some(n) = value.as_f64() {
        if let Some(min) = schema.minimum {
            if n < min {
                errors.push(ValidationError {
                    field: field_path.to_string(),
                    expected: format!(">={}", min),
                    actual: n.to_string(),
                    message: format!("Field '{}' must be >= {}, got {}", field_path, min, n),
                });
            }
        }
        
        if let Some(max) = schema.maximum {
            if n > max {
                errors.push(ValidationError {
                    field: field_path.to_string(),
                    expected: format!("<={}", max),
                    actual: n.to_string(),
                    message: format!("Field '{}' must be <= {}, got {}", field_path, max, n),
                });
            }
        }
    }
}

fn validate_boolean(_value: &Value, _schema: &Schema, _field_path: &str, _errors: &mut Vec<ValidationError>) {
}

fn validate_array(value: &Value, schema: &Schema, field_path: &str, errors: &mut Vec<ValidationError>) {
    if let Some(arr) = value.as_array() {
        if let Some(items_schema) = &schema.items {
            for (index, item) in arr.iter().enumerate() {
                let item_path = format!("{}[{}]", field_path, index);
                let item_result = validate(item, items_schema, &item_path);
                if !item_result.valid {
                    errors.extend(item_result.errors);
                }
            }
        }
    }
}

fn validate_object(value: &Value, schema: &Schema, field_path: &str, errors: &mut Vec<ValidationError>) {
    if let Some(obj) = value.as_object() {
        if let Some(properties) = &schema.properties {
            for (prop_name, prop_schema) in properties {
                let prop_path = if field_path.is_empty() {
                    prop_name.clone()
                } else {
                    format!("{}.{}", field_path, prop_name)
                };
                
                if let Some(prop_value) = obj.get(prop_name) {
                    let prop_result = validate(prop_value, prop_schema, &prop_path);
                    if !prop_result.valid {
                        errors.extend(prop_result.errors);
                    }
                } else if prop_schema.required.unwrap_or(false) {
                    errors.push(ValidationError {
                        field: prop_path.clone(),
                        expected: "present".to_string(),
                        actual: "missing".to_string(),
                        message: format!("Required field '{}' is missing", prop_path),
                    });
                }
            }
        }
    }
}

fn check_enum(value: &Value, schema: &Schema, field_path: &str, errors: &mut Vec<ValidationError>) {
    if let Some(enum_values) = &schema.enum_values {
        if !enum_values.contains(value) {
            let expected = format!("one of {:?}", enum_values);
            errors.push(ValidationError {
                field: field_path.to_string(),
                expected,
                actual: value.to_string(),
                message: format!("Field '{}' value not in allowed enum", field_path),
            });
        }
    }
}

fn get_json_type(value: &Value) -> String {
    match value {
        Value::String(_) => "string",
        Value::Number(_) => "number",
        Value::Bool(_) => "boolean",
        Value::Array(_) => "array",
        Value::Object(_) => "object",
        Value::Null => "null",
    }.to_string()
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;

    #[test]
    fn test_validate_string_type() {
        let schema = Schema {
            config_type: ConfigType::String,
            required: None,
            minimum: None,
            maximum: None,
            min_length: None,
            max_length: None,
            pattern: None,
            items: None,
            properties: None,
            enum_values: None,
        };
        
        let result = validate(&json!("hello"), &schema, "test");
        assert!(result.valid);
        
        let result = validate(&json!(123), &schema, "test");
        assert!(!result.valid);
        assert_eq!(result.errors[0].expected, "string");
        assert_eq!(result.errors[0].actual, "number");
    }

    #[test]
    fn test_validate_required_field() {
        let schema = Schema {
            config_type: ConfigType::String,
            required: Some(true),
            minimum: None,
            maximum: None,
            min_length: None,
            max_length: None,
            pattern: None,
            items: None,
            properties: None,
            enum_values: None,
        };
        
        let result = validate(&json!(null), &schema, "test");
        assert!(!result.valid);
        assert_eq!(result.errors[0].field, "test");
    }

    #[test]
    fn test_validate_number_range() {
        let schema = Schema {
            config_type: ConfigType::Number,
            required: None,
            minimum: Some(0.0),
            maximum: Some(100.0),
            min_length: None,
            max_length: None,
            pattern: None,
            items: None,
            properties: None,
            enum_values: None,
        };
        
        let result = validate(&json!(50), &schema, "test");
        assert!(result.valid);
        
        let result = validate(&json!(-1), &schema, "test");
        assert!(!result.valid);
        
        let result = validate(&json!(101), &schema, "test");
        assert!(!result.valid);
    }

    #[test]
    fn test_validate_string_pattern() {
        let schema = Schema {
            config_type: ConfigType::String,
            required: None,
            minimum: None,
            maximum: None,
            min_length: None,
            max_length: None,
            pattern: Some("^[a-z]+$".to_string()),
            items: None,
            properties: None,
            enum_values: None,
        };
        
        let result = validate(&json!("hello"), &schema, "test");
        assert!(result.valid);
        
        let result = validate(&json!("HELLO"), &schema, "test");
        assert!(!result.valid);
    }

    #[test]
    fn test_schema_deserialization_with_aliases() {
        let json_schema = json!({
            "type": "number",
            "min": 1.0,
            "max": 300.0
        });
        
        let schema: Schema = serde_json::from_value(json_schema).unwrap();
        assert_eq!(schema.config_type, ConfigType::Number);
        assert_eq!(schema.minimum, Some(1.0));
        assert_eq!(schema.maximum, Some(300.0));
        
        let result = validate(&json!(500.0), &schema, "test");
        assert!(!result.valid);
        assert_eq!(result.errors[0].expected, "<=300");
        
        let result = validate(&json!(150.0), &schema, "test");
        assert!(result.valid);
    }

    #[test]
    fn test_schema_deserialization_minlength_alias() {
        let json_schema = json!({
            "type": "string",
            "minLength": 3,
            "maxLength": 10
        });
        
        let schema: Schema = serde_json::from_value(json_schema).unwrap();
        assert_eq!(schema.min_length, Some(3));
        assert_eq!(schema.max_length, Some(10));
        
        let result = validate(&json!("ab"), &schema, "test");
        assert!(!result.valid);
        
        let result = validate(&json!("hello"), &schema, "test");
        assert!(result.valid);
    }

    #[test]
    fn test_schema_deserialization_enum_alias() {
        let json_schema = json!({
            "type": "string",
            "enum": ["a", "b", "c"]
        });
        
        let schema: Schema = serde_json::from_value(json_schema).unwrap();
        assert!(schema.enum_values.is_some());
        
        let result = validate(&json!("a"), &schema, "test");
        assert!(result.valid);
        
        let result = validate(&json!("d"), &schema, "test");
        assert!(!result.valid);
    }
}
