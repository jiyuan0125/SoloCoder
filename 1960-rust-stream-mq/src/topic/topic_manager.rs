use std::fs;
use std::io;
use std::path::PathBuf;
use std::sync::Arc;

use dashmap::DashMap;

use crate::consumer::ConsumerGroupManager;
use crate::storage::LogStore;
use crate::storage::Message;

pub struct Topic {
    pub name: String,
    pub store: LogStore,
}

pub struct TopicManager {
    base_dir: PathBuf,
    topics: DashMap<String, Topic>,
    group_manager: Arc<ConsumerGroupManager>,
}

impl TopicManager {
    pub fn new(base_dir: PathBuf) -> io::Result<Self> {
        fs::create_dir_all(&base_dir)?;
        
        let group_dir = base_dir.join("offsets");
        let group_manager = Arc::new(ConsumerGroupManager::new(group_dir)?);
        
        let topics = DashMap::new();
        let manager = TopicManager {
            base_dir,
            topics,
            group_manager,
        };
        
        manager.load_existing_topics()?;
        
        Ok(manager)
    }
    
    fn load_existing_topics(&self) -> io::Result<()> {
        if let Ok(entries) = fs::read_dir(&self.base_dir) {
            for entry in entries {
                let entry = entry?;
                let path = entry.path();
                if path.is_dir() {
                    if let Some(name) = path.file_name() {
                        let name_str = name.to_string_lossy().to_string();
                        if name_str != "offsets" {
                            let topic_dir = self.base_dir.join(&name_str);
                            let store = LogStore::new(topic_dir)?;
                            let topic = Topic {
                                name: name_str.clone(),
                                store,
                            };
                            self.topics.insert(name_str, topic);
                        }
                    }
                }
            }
        }
        Ok(())
    }
    
    fn get_store(&self, topic_name: &str) -> io::Result<LogStore> {
        if self.topics.contains_key(topic_name) {
            let topic_dir = self.base_dir.join(topic_name);
            return LogStore::new(topic_dir);
        }
        
        let topic_dir = self.base_dir.join(topic_name);
        fs::create_dir_all(&topic_dir)?;
        
        let store = LogStore::new(topic_dir)?;
        let topic = Topic {
            name: topic_name.to_string(),
            store: LogStore::new(self.base_dir.join(topic_name))?,
        };
        
        self.topics.insert(topic_name.to_string(), topic);
        
        Ok(store)
    }
    
    pub fn produce(&self, topic_name: &str, payload: &[u8]) -> io::Result<u64> {
        let store = self.get_store(topic_name)?;
        store.append(payload)
    }
    
    pub fn consume(
        &self,
        topic_name: &str,
        group: &str,
        limit: usize,
    ) -> io::Result<Vec<Message>> {
        let store = self.get_store(topic_name)?;
        let offset = self.group_manager.get_offset(topic_name, group);
        let messages = store.read(offset, limit)?;
        Ok(messages)
    }
    
    pub fn ack(&self, topic_name: &str, group: &str, offset: u64) -> io::Result<()> {
        self.group_manager.update_offset(topic_name, group, offset)
    }
    
    pub fn set_consumer_offset(
        &self,
        topic_name: &str,
        group: &str,
        offset: u64,
    ) -> io::Result<()> {
        self.group_manager.set_offset(topic_name, group, offset)
    }
    
    pub fn get_consumer_offset(&self, topic_name: &str, group: &str) -> u64 {
        self.group_manager.get_offset(topic_name, group)
    }
    
    pub fn list_topics(&self) -> Vec<String> {
        self.topics.iter().map(|entry| entry.key().clone()).collect()
    }
    
    pub fn get_last_offset(&self, topic_name: &str) -> io::Result<u64> {
        let store = self.get_store(topic_name)?;
        Ok(store.get_last_offset())
    }
}


