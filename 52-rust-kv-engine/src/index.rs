use crate::storage::{read_record, Record, RecordType};
use std::collections::HashMap;
use std::fs::File;
use std::io;

#[derive(Debug, Clone)]
pub struct IndexEntry {
    pub offset: u64,
    pub length: u64,
    pub is_alive: bool,
}

#[derive(Debug, Default)]
pub struct Index {
    entries: HashMap<String, IndexEntry>,
}

impl Index {
    pub fn new() -> Self {
        Index {
            entries: HashMap::new(),
        }
    }

    pub fn build_from_file(file: &mut File) -> io::Result<Self> {
        let mut index = Index::new();
        let mut offset = 0u64;

        loop {
            match read_record(file, offset) {
                Ok(Some((record, record_size))) => {
                    index.update(&record, offset, record_size);
                    offset += record_size;
                }
                Ok(None) => break,
                Err(e) => {
                    if e.kind() == io::ErrorKind::InvalidData {
                        eprintln!("Warning: Corrupted record at offset {}, skipping (simulating crash recovery)", offset);
                        break;
                    }
                    return Err(e);
                }
            }
        }

        Ok(index)
    }

    pub fn update(&mut self, record: &Record, offset: u64, length: u64) {
        let is_alive = match record.op_type {
            RecordType::Put => true,
            RecordType::Delete => false,
        };

        let entry = IndexEntry {
            offset,
            length,
            is_alive,
        };

        self.entries.insert(record.key.clone(), entry);
    }

    pub fn get(&self, key: &str) -> Option<&IndexEntry> {
        self.entries.get(key)
    }

    pub fn scan_prefix(&self, prefix: &str) -> Vec<String> {
        let mut keys: Vec<String> = self
            .entries
            .iter()
            .filter(|(key, entry)| key.starts_with(prefix) && entry.is_alive)
            .map(|(key, _)| key.clone())
            .collect();
        
        keys.sort();
        keys
    }

    pub fn count(&self) -> usize {
        self.entries.values().filter(|e| e.is_alive).count()
    }

    pub fn keys(&self) -> Vec<(String, IndexEntry)> {
        self.entries
            .iter()
            .map(|(k, v)| (k.clone(), v.clone()))
            .collect()
    }

    pub fn clear(&mut self) {
        self.entries.clear();
    }
}
