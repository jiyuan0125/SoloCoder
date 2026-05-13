use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs::{File, OpenOptions};
use std::io::{BufRead, BufReader, Write};
use std::path::{Path, PathBuf};
use std::sync::{Arc, Mutex};

use crate::types::{GroupMetadata, Message, PartitionMetadata, TopicConfig};

#[derive(Debug, Clone)]
pub struct Storage {
    base_dir: PathBuf,
    topics: Arc<Mutex<HashMap<String, TopicConfig>>>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct LogEntry {
    #[serde(flatten)]
    message: Message,
}

impl Storage {
    pub fn new(base_dir: &str) -> Self {
        let base_path = PathBuf::from(base_dir);
        if !base_path.exists() {
            std::fs::create_dir_all(&base_path).unwrap();
        }

        let topics = load_topics(&base_path).unwrap_or_default();

        Storage {
            base_dir: base_path,
            topics: Arc::new(Mutex::new(topics)),
        }
    }

    pub fn create_topic(&self, config: &TopicConfig) -> Result<()> {
        let mut topics = self.topics.lock().unwrap();
        
        if topics.contains_key(&config.name) {
            return Ok(());
        }

        let topic_dir = self.get_topic_dir(&config.name);
        std::fs::create_dir_all(&topic_dir)?;

        for partition in 0..config.partitions {
            let partition_dir = topic_dir.join(format!("partition-{}", partition));
            std::fs::create_dir_all(&partition_dir)?;

            let log_file = partition_dir.join("messages.log");
            OpenOptions::new()
                .create(true)
                .write(true)
                .append(true)
                .open(&log_file)?;

            let groups_dir = partition_dir.join("groups");
            std::fs::create_dir_all(&groups_dir)?;
        }

        let config_file = topic_dir.join("topic.json");
        let config_json = serde_json::to_string_pretty(config)?;
        let mut file = File::create(&config_file)?;
        file.write_all(config_json.as_bytes())?;

        topics.insert(config.name.clone(), config.clone());

        Ok(())
    }

    pub fn get_topic_config(&self, name: &str) -> Option<TopicConfig> {
        let topics = self.topics.lock().unwrap();
        topics.get(name).cloned()
    }

    pub fn get_all_topics(&self) -> HashMap<String, TopicConfig> {
        let topics = self.topics.lock().unwrap();
        topics.clone()
    }

    pub fn append_message(&self, topic: &str, partition: u32, message: &Message) -> Result<()> {
        let log_path = self.get_log_path(topic, partition);
        let mut file = OpenOptions::new()
            .create(true)
            .write(true)
            .append(true)
            .open(&log_path)?;

        let entry = LogEntry {
            message: message.clone(),
        };
        let line = serde_json::to_string(&entry)? + "\n";
        file.write_all(line.as_bytes())?;
        file.flush()?;

        Ok(())
    }

    pub fn load_messages(&self, topic: &str, partition: u32) -> Result<Vec<Message>> {
        let log_path = self.get_log_path(topic, partition);
        if !log_path.exists() {
            return Ok(Vec::new());
        }

        let file = File::open(&log_path)?;
        let reader = BufReader::new(file);
        let mut messages = Vec::new();

        for line in reader.lines() {
            let line = line?;
            if line.is_empty() {
                continue;
            }
            let entry: LogEntry = serde_json::from_str(&line)?;
            messages.push(entry.message);
        }

        Ok(messages)
    }

    pub fn save_group_metadata(
        &self,
        topic: &str,
        partition: u32,
        group: &str,
        metadata: &GroupMetadata,
    ) -> Result<()> {
        let groups_dir = self.get_groups_dir(topic, partition);
        std::fs::create_dir_all(&groups_dir)?;

        let group_file = groups_dir.join(format!("{}.json", group));
        let json = serde_json::to_string_pretty(metadata)?;
        let mut file = File::create(&group_file)?;
        file.write_all(json.as_bytes())?;

        Ok(())
    }

    pub fn load_group_metadata(
        &self,
        topic: &str,
        partition: u32,
        group: &str,
    ) -> Result<Option<GroupMetadata>> {
        let group_file = self.get_groups_dir(topic, partition).join(format!("{}.json", group));
        if !group_file.exists() {
            return Ok(None);
        }

        let file = File::open(&group_file)?;
        let metadata: GroupMetadata = serde_json::from_reader(file)?;

        Ok(Some(metadata))
    }

    pub fn load_all_groups_for_partition(
        &self,
        topic: &str,
        partition: u32,
    ) -> Result<HashMap<String, GroupMetadata>> {
        let groups_dir = self.get_groups_dir(topic, partition);
        if !groups_dir.exists() {
            return Ok(HashMap::new());
        }

        let mut result = HashMap::new();

        for entry in std::fs::read_dir(&groups_dir)? {
            let entry = entry?;
            let path = entry.path();
            if path.extension().map(|s| s == "json").unwrap_or(false) {
                if let Some(group_name) = path.file_stem().map(|s| s.to_string_lossy().to_string()) {
                    let file = File::open(&path)?;
                    let metadata: GroupMetadata = serde_json::from_reader(file)?;
                    result.insert(group_name, metadata);
                }
            }
        }

        Ok(result)
    }

    pub fn save_partition_metadata(
        &self,
        topic: &str,
        partition: u32,
        metadata: &PartitionMetadata,
    ) -> Result<()> {
        let partition_dir = self.get_partition_dir(topic, partition);
        std::fs::create_dir_all(&partition_dir)?;

        let meta_file = partition_dir.join("metadata.json");
        let json = serde_json::to_string_pretty(metadata)?;
        let mut file = File::create(&meta_file)?;
        file.write_all(json.as_bytes())?;

        Ok(())
    }

    pub fn load_partition_metadata(
        &self,
        topic: &str,
        partition: u32,
    ) -> Result<Option<PartitionMetadata>> {
        let meta_file = self.get_partition_dir(topic, partition).join("metadata.json");
        if !meta_file.exists() {
            return Ok(None);
        }

        let file = File::open(&meta_file)?;
        let metadata: PartitionMetadata = serde_json::from_reader(file)?;

        Ok(Some(metadata))
    }

    fn get_topic_dir(&self, topic: &str) -> PathBuf {
        self.base_dir.join(topic)
    }

    fn get_partition_dir(&self, topic: &str, partition: u32) -> PathBuf {
        self.get_topic_dir(topic).join(format!("partition-{}", partition))
    }

    fn get_log_path(&self, topic: &str, partition: u32) -> PathBuf {
        self.get_partition_dir(topic, partition).join("messages.log")
    }

    fn get_groups_dir(&self, topic: &str, partition: u32) -> PathBuf {
        self.get_partition_dir(topic, partition).join("groups")
    }
}

fn load_topics(base_dir: &Path) -> Result<HashMap<String, TopicConfig>> {
    let mut topics = HashMap::new();

    if !base_dir.exists() {
        return Ok(topics);
    }

    for entry in std::fs::read_dir(base_dir)? {
        let entry = entry?;
        let path = entry.path();
        if path.is_dir() {
            let config_file = path.join("topic.json");
            if config_file.exists() {
                let file = File::open(&config_file)?;
                let config: TopicConfig = serde_json::from_reader(file)?;
                topics.insert(config.name.clone(), config);
            }
        }
    }

    Ok(topics)
}
