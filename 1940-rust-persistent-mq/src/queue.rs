use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use std::time::{SystemTime, UNIX_EPOCH};
use twox_hash::XxHash64;
use std::hash::{Hash, Hasher};

use crate::storage::Storage;
use crate::types::{GroupMetadata, Message, PartitionMetadata, TopicConfig};

const DEFAULT_ACK_TIMEOUT_SECS: u64 = 30;

pub struct PartitionQueue {
    topic: String,
    partition: u32,
    storage: Storage,
    next_offset: u64,
    messages: Vec<Message>,
    groups: HashMap<String, GroupState>,
    round_robin_counter: Mutex<u32>,
}

#[derive(Clone)]
struct GroupState {
    last_ack_offset: u64,
    pending_acks: HashMap<u64, u64>,
}

pub struct MessageQueue {
    storage: Storage,
    partitions: Arc<Mutex<HashMap<String, Vec<Arc<Mutex<PartitionQueue>>>>>>,
}

impl PartitionQueue {
    pub fn new(topic: &str, partition: u32, storage: Storage) -> Self {
        let mut messages = storage
            .load_messages(topic, partition)
            .unwrap_or_default();

        let next_offset = if let Some(last) = messages.last() {
            last.offset + 1
        } else {
            0
        };

        let groups_metadata = storage
            .load_all_groups_for_partition(topic, partition)
            .unwrap_or_default();

        let mut groups = HashMap::new();
        for (group_name, metadata) in groups_metadata {
            groups.insert(
                group_name,
                GroupState {
                    last_ack_offset: metadata.last_ack_offset,
                    pending_acks: metadata.pending_acks,
                },
            );
        }

        PartitionQueue {
            topic: topic.to_string(),
            partition,
            storage,
            next_offset,
            messages,
            groups,
            round_robin_counter: Mutex::new(0),
        }
    }

    pub fn produce(&mut self, key: Option<&str>, payload: &str) -> u64 {
        let offset = self.next_offset;
        let timestamp = current_timestamp();

        let message = Message {
            offset,
            key: key.map(|s| s.to_string()),
            payload: payload.to_string(),
            timestamp,
            partition: self.partition,
        };

        self.storage
            .append_message(&self.topic, self.partition, &message)
            .expect("Failed to append message");

        self.messages.push(message);
        self.next_offset += 1;

        let meta = PartitionMetadata {
            topic: self.topic.clone(),
            partition: self.partition,
            next_offset: self.next_offset,
        };
        self.storage
            .save_partition_metadata(&self.topic, self.partition, &meta)
            .expect("Failed to save partition metadata");

        offset
    }

    pub fn consume(&mut self, group: &str) -> Option<Message> {
        self.cleanup_expired_acks(group);

        let group_state = self.get_or_create_group(group);
        let next_offset = self.next_offset(&group_state);

        let now = current_timestamp_secs();
        let timeout = DEFAULT_ACK_TIMEOUT_SECS;

        for message in &self.messages {
            if message.offset < next_offset {
                continue;
            }

            if let Some(&expire_at) = group_state.pending_acks.get(&message.offset) {
                if expire_at > now {
                    continue;
                }
            }

            let expire_at = now + timeout;
            let state = self.groups.get_mut(group).unwrap();
            state.pending_acks.insert(message.offset, expire_at);

            self.persist_group_state(group);

            return Some(message.clone());
        }

        None
    }

    pub fn ack(&mut self, group: &str, offset: u64) -> bool {
        let state = match self.groups.get_mut(group) {
            Some(s) => s,
            None => return false,
        };

        if !state.pending_acks.contains_key(&offset) {
            return false;
        }

        state.pending_acks.remove(&offset);

        if state.last_ack_offset == u64::MAX || offset == state.last_ack_offset + 1 {
            state.last_ack_offset = offset;
            while let Some(_) = state.pending_acks.get(&(state.last_ack_offset + 1)) {
                state.last_ack_offset += 1;
            }
        }

        self.persist_group_state(group);
        true
    }

    fn get_or_create_group(&mut self, group: &str) -> GroupState {
        self.groups
            .entry(group.to_string())
            .or_insert_with(|| GroupState {
                last_ack_offset: u64::MAX,
                pending_acks: HashMap::new(),
            })
            .clone()
    }

    fn next_offset(&self, group_state: &GroupState) -> u64 {
        if group_state.last_ack_offset == u64::MAX {
            0
        } else {
            group_state.last_ack_offset + 1
        }
    }

    fn cleanup_expired_acks(&mut self, group: &str) {
        let now = current_timestamp_secs();
        if let Some(state) = self.groups.get_mut(group) {
            state.pending_acks.retain(|_, &mut expire_at| expire_at > now);
        }
    }

    fn persist_group_state(&self, group: &str) {
        if let Some(state) = self.groups.get(group) {
            let metadata = GroupMetadata {
                last_ack_offset: state.last_ack_offset,
                pending_acks: state.pending_acks.clone(),
            };
            let _ = self
                .storage
                .save_group_metadata(&self.topic, self.partition, group, &metadata);
        }
    }
}

impl MessageQueue {
    pub fn new(storage: Storage) -> Self {
        let partitions = Arc::new(Mutex::new(HashMap::new()));

        let topics = {
            let storage_topics = storage.get_all_topics();
            storage_topics
        };

        for (topic_name, config) in topics.iter() {
            let mut partition_queues = Vec::new();
            for partition in 0..config.partitions {
                let queue = PartitionQueue::new(topic_name, partition, storage.clone());
                partition_queues.push(Arc::new(Mutex::new(queue)));
            }

            let mut p = partitions.lock().unwrap();
            p.insert(topic_name.clone(), partition_queues);
        }

        MessageQueue {
            storage,
            partitions,
        }
    }

    pub fn create_topic(&self, config: &TopicConfig) {
        self.storage
            .create_topic(config)
            .expect("Failed to create topic");

        let mut partition_queues = Vec::new();
        for partition in 0..config.partitions {
            let queue = PartitionQueue::new(&config.name, partition, self.storage.clone());
            partition_queues.push(Arc::new(Mutex::new(queue)));
        }

        let mut p = self.partitions.lock().unwrap();
        p.insert(config.name.clone(), partition_queues);
    }

    pub fn get_topic_config(&self, name: &str) -> Option<TopicConfig> {
        self.storage.get_topic_config(name)
    }

    pub fn produce(&self, topic: &str, key: Option<&str>, payload: &str) -> (u64, u32) {
        let partitions = self.partitions.lock().unwrap();
        let topic_partitions = partitions.get(topic).expect("Topic not found");

        let partition_idx = match key {
            Some(k) => {
                let mut hasher = XxHash64::default();
                k.hash(&mut hasher);
                (hasher.finish() as usize) % topic_partitions.len()
            }
            None => {
                let counter = &topic_partitions[0].lock().unwrap().round_robin_counter;
                let mut c = counter.lock().unwrap();
                let idx = *c as usize % topic_partitions.len();
                *c = c.wrapping_add(1);
                idx
            }
        };

        let mut queue = topic_partitions[partition_idx].lock().unwrap();
        let offset = queue.produce(key, payload);
        (offset, partition_idx as u32)
    }

    pub fn consume(&self, topic: &str, group: &str) -> Option<Message> {
        let partitions = self.partitions.lock().unwrap();
        let topic_partitions = partitions.get(topic).expect("Topic not found");

        for partition_queue in topic_partitions {
            let mut queue = partition_queue.lock().unwrap();
            if let Some(msg) = queue.consume(group) {
                return Some(msg);
            }
        }

        None
    }

    pub fn ack(&self, topic: &str, group: &str, offset: u64, partition: u32) -> bool {
        let partitions = self.partitions.lock().unwrap();
        let topic_partitions = partitions.get(topic).expect("Topic not found");

        if partition as usize >= topic_partitions.len() {
            return false;
        }

        let mut queue = topic_partitions[partition as usize].lock().unwrap();
        queue.ack(group, offset)
    }

    pub fn find_partition_for_offset(&self, topic: &str, offset: u64) -> Option<u32> {
        let partitions = self.partitions.lock().unwrap();
        let topic_partitions = partitions.get(topic)?;

        for partition_queue in topic_partitions {
            let queue = partition_queue.lock().unwrap();
            if queue.messages.iter().any(|m| m.offset == offset) {
                return Some(queue.partition);
            }
        }

        None
    }
}

fn current_timestamp() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("Time went backwards")
        .as_millis() as u64
}

fn current_timestamp_secs() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("Time went backwards")
        .as_secs()
}
