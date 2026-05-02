use std::collections::HashMap;
use std::hash::Hash;

#[derive(Debug)]
struct Node<K> {
    key: K,
    prev: Option<usize>,
    next: Option<usize>,
}

pub struct LruList<K: Eq + Hash + Clone> {
    nodes: Vec<Node<K>>,
    key_to_index: HashMap<K, usize>,
    head: Option<usize>,
    tail: Option<usize>,
    free_indices: Vec<usize>,
}

impl<K: Eq + Hash + Clone> LruList<K> {
    pub fn new() -> Self {
        Self {
            nodes: Vec::new(),
            key_to_index: HashMap::new(),
            head: None,
            tail: None,
            free_indices: Vec::new(),
        }
    }

    pub fn with_capacity(capacity: usize) -> Self {
        Self {
            nodes: Vec::with_capacity(capacity),
            key_to_index: HashMap::with_capacity(capacity),
            head: None,
            tail: None,
            free_indices: Vec::new(),
        }
    }

    pub fn access(&mut self, key: &K) {
        if let Some(&idx) = self.key_to_index.get(key) {
            self.move_to_front(idx);
        }
    }

    pub fn insert(&mut self, key: K) {
        let key_clone = key.clone();
        let idx = if let Some(free_idx) = self.free_indices.pop() {
            self.nodes[free_idx] = Node {
                key,
                prev: None,
                next: self.head,
            };
            free_idx
        } else {
            self.nodes.push(Node {
                key,
                prev: None,
                next: self.head,
            });
            self.nodes.len() - 1
        };

        if let Some(old_head) = self.head {
            self.nodes[old_head].prev = Some(idx);
        }
        self.head = Some(idx);

        if self.tail.is_none() {
            self.tail = Some(idx);
        }

        self.key_to_index.insert(key_clone, idx);
    }

    pub fn remove(&mut self, key: &K) -> bool {
        if let Some(&idx) = self.key_to_index.get(key) {
            self.remove_node(idx);
            self.key_to_index.remove(key);
            self.free_indices.push(idx);
            true
        } else {
            false
        }
    }

    pub fn pop_lru(&mut self) -> Option<K> {
        let tail_idx = self.tail?;
        let node = &self.nodes[tail_idx];
        let key = node.key.clone();
        
        self.remove_node(tail_idx);
        self.key_to_index.remove(&key);
        self.free_indices.push(tail_idx);
        
        Some(key)
    }

    pub fn contains(&self, key: &K) -> bool {
        self.key_to_index.contains_key(key)
    }

    pub fn len(&self) -> usize {
        self.key_to_index.len()
    }

    pub fn is_empty(&self) -> bool {
        self.key_to_index.is_empty()
    }

    fn move_to_front(&mut self, idx: usize) {
        if self.head == Some(idx) {
            return;
        }

        self.detach(idx);

        let node = &mut self.nodes[idx];
        node.prev = None;
        node.next = self.head;

        if let Some(old_head) = self.head {
            self.nodes[old_head].prev = Some(idx);
        }
        self.head = Some(idx);

        if self.tail.is_none() {
            self.tail = Some(idx);
        }
    }

    fn remove_node(&mut self, idx: usize) {
        self.detach(idx);
        self.nodes[idx].prev = None;
        self.nodes[idx].next = None;
    }

    fn detach(&mut self, idx: usize) {
        let node = &self.nodes[idx];
        let prev_idx = node.prev;
        let next_idx = node.next;

        if let Some(p) = prev_idx {
            self.nodes[p].next = next_idx;
        } else {
            self.head = next_idx;
        }

        if let Some(n) = next_idx {
            self.nodes[n].prev = prev_idx;
        } else {
            self.tail = prev_idx;
        }
    }
}

impl<K: Eq + Hash + Clone> Default for LruList<K> {
    fn default() -> Self {
        Self::new()
    }
}
