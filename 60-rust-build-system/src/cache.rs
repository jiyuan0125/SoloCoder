use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs::File;
use std::io::{Read, Write};
use std::path::Path;

#[derive(Debug, Clone, Serialize, Deserialize, Default)]
pub struct BuildCache {
    #[serde(rename = "targets")]
    pub target_hashes: HashMap<String, TargetCacheEntry>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TargetCacheEntry {
    pub input_hashes: HashMap<String, String>,
    pub output_hashes: HashMap<String, String>,
}

impl BuildCache {
    const CACHE_FILE: &'static str = ".build_cache";

    pub fn load() -> Self {
        let path = Path::new(Self::CACHE_FILE);
        if !path.exists() {
            return BuildCache::default();
        }

        let mut file = match File::open(path) {
            Ok(f) => f,
            Err(_) => return BuildCache::default(),
        };

        let mut content = String::new();
        if file.read_to_string(&mut content).is_err() {
            return BuildCache::default();
        }

        serde_json::from_str(&content).unwrap_or_default()
    }

    pub fn save(&self) -> Result<(), String> {
        let path = Path::new(Self::CACHE_FILE);
        let content = serde_json::to_string_pretty(self)
            .map_err(|e| format!("Failed to serialize cache: {}", e))?;
        
        let mut file = File::create(path)
            .map_err(|e| format!("Failed to create cache file: {}", e))?;
        
        file.write_all(content.as_bytes())
            .map_err(|e| format!("Failed to write cache file: {}", e))?;
        
        Ok(())
    }

    pub fn get_target_entry(&self, target_name: &str) -> Option<&TargetCacheEntry> {
        self.target_hashes.get(target_name)
    }

    pub fn set_target_entry(&mut self, target_name: String, entry: TargetCacheEntry) {
        self.target_hashes.insert(target_name, entry);
    }

    pub fn is_target_cached(
        &self,
        target_name: &str,
        current_input_hashes: &HashMap<String, String>,
    ) -> bool {
        let Some(entry) = self.get_target_entry(target_name) else {
            return false;
        };

        if entry.input_hashes.len() != current_input_hashes.len() {
            return false;
        }

        for (path, hash) in current_input_hashes {
            match entry.input_hashes.get(path) {
                Some(cached_hash) if cached_hash == hash => continue,
                _ => return false,
            }
        }

        true
    }
}
