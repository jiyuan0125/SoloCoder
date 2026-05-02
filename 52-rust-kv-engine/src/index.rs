use std::collections::HashMap;

#[derive(Debug, Clone)]
pub struct IndexEntry {
    pub offset: u64,
    pub length: usize,
}

pub struct IndexManager {
    index: HashMap<String, IndexEntry>,
}

impl IndexManager {
    pub fn new() -> Self {
        IndexManager {
            index: HashMap::new(),
        }
    }

    pub fn put(&mut self, key: String, offset: u64, length: usize) {
        self.index.insert(key, IndexEntry { offset, length });
    }

    pub fn get(&self, key: &str) -> Option<(u64, usize)> {
        self.index.get(key).map(|entry| (entry.offset, entry.length))
    }

    pub fn remove(&mut self, key: &str) {
        self.index.remove(key);
    }

    pub fn scan_prefix(&self, prefix: &str) -> Vec<String> {
        let mut result: Vec<String> = self.index
            .keys()
            .filter(|key| key.starts_with(prefix))
            .cloned()
            .collect();
        result.sort();
        result
    }

    pub fn count(&self) -> usize {
        self.index.len()
    }
}

impl Default for IndexManager {
    fn default() -> Self {
        Self::new()
    }
}
