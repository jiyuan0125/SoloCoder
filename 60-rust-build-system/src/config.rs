use serde::Deserialize;
use std::collections::HashMap;
use std::path::Path;

#[derive(Debug, Clone, Deserialize)]
pub struct BuildConfig {
    pub targets: Vec<TargetConfig>,
}

#[derive(Debug, Clone, Deserialize)]
pub struct TargetConfig {
    pub name: String,
    pub command: String,
    pub inputs: Vec<String>,
    pub outputs: Vec<String>,
}

impl BuildConfig {
    pub fn from_file<P: AsRef<Path>>(path: P) -> Result<Self, String> {
        let content = std::fs::read_to_string(&path)
            .map_err(|e| format!("Failed to read config file: {}", e))?;
        serde_json::from_str(&content)
            .map_err(|e| format!("Failed to parse JSON: {}", e))
    }

    pub fn validate(&self) -> Result<(), String> {
        let mut names = HashMap::new();
        for target in &self.targets {
            if target.name.is_empty() {
                return Err("Target name cannot be empty".to_string());
            }
            if names.contains_key(&target.name) {
                return Err(format!("Duplicate target name: {}", target.name));
            }
            names.insert(target.name.clone(), ());
        }
        Ok(())
    }
}
