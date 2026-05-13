use std::collections::HashMap;
use std::fs::{self, File, OpenOptions};
use std::io::{self, Read, Write};
use std::path::PathBuf;

use dashmap::DashMap;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GroupOffset {
    pub topic: String,
    pub group: String,
    pub offset: u64,
}

pub struct ConsumerGroupManager {
    base_dir: PathBuf,
    offsets: DashMap<String, HashMap<String, u64>>,
}

impl ConsumerGroupManager {
    pub fn new(base_dir: PathBuf) -> io::Result<Self> {
        fs::create_dir_all(&base_dir)?;
        
        let offsets = DashMap::new();
        let manager = ConsumerGroupManager {
            base_dir,
            offsets,
        };
        
        manager.load_all_offsets()?;
        
        Ok(manager)
    }
    
    fn load_all_offsets(&self) -> io::Result<()> {
        if let Ok(entries) = fs::read_dir(&self.base_dir) {
            for entry in entries {
                let entry = entry?;
                let path = entry.path();
                if path.extension().map_or(false, |e| e == "offset") {
                    self.load_offset_file(&path)?;
                }
            }
        }
        Ok(())
    }
    
    fn load_offset_file(&self, path: &PathBuf) -> io::Result<()> {
        let mut file = File::open(path)?;
        let mut content = String::new();
        file.read_to_string(&mut content)?;
        
        let offset: GroupOffset = serde_json::from_str(&content)
            .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))?;
        
        let mut topic_offsets = self.offsets
            .entry(offset.topic.clone())
            .or_insert_with(HashMap::new);
        
        topic_offsets.insert(offset.group.clone(), offset.offset);
        
        Ok(())
    }
    
    fn save_offset_file(&self, topic: &str, group: &str, offset: u64) -> io::Result<()> {
        let group_offset = GroupOffset {
            topic: topic.to_string(),
            group: group.to_string(),
            offset,
        };
        
        let filename = format!("{}-{}.offset", topic, group);
        let path = self.base_dir.join(filename);
        
        let mut file = OpenOptions::new()
            .create(true)
            .write(true)
            .truncate(true)
            .open(&path)?;
        
        let json = serde_json::to_string(&group_offset)
            .map_err(|e| io::Error::new(io::ErrorKind::InvalidData, e))?;
        
        file.write_all(json.as_bytes())?;
        file.flush()?;
        
        Ok(())
    }
    
    pub fn get_offset(&self, topic: &str, group: &str) -> u64 {
        self.offsets
            .get(topic)
            .and_then(|topic_offsets| topic_offsets.get(group).copied())
            .unwrap_or(0)
    }
    
    pub fn set_offset(&self, topic: &str, group: &str, offset: u64) -> io::Result<()> {
        let mut topic_offsets = self.offsets
            .entry(topic.to_string())
            .or_insert_with(HashMap::new);
        
        topic_offsets.insert(group.to_string(), offset);
        drop(topic_offsets);
        
        self.save_offset_file(topic, group, offset)?;
        
        Ok(())
    }
    
    pub fn update_offset(&self, topic: &str, group: &str, new_offset: u64) -> io::Result<()> {
        self.set_offset(topic, group, new_offset)
    }
    
    pub fn list_groups(&self, topic: &str) -> Vec<String> {
        self.offsets
            .get(topic)
            .map(|topic_offsets| topic_offsets.keys().cloned().collect())
            .unwrap_or_default()
    }
}
