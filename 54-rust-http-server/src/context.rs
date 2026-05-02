use std::collections::HashMap;
use std::sync::mpsc::Sender;
use std::sync::{Arc, Mutex};

pub type BroadcastSender = Sender<Vec<u8>>;

#[derive(Clone)]
pub struct Context {
    ws_connections: Arc<Mutex<HashMap<usize, BroadcastSender>>>,
}

impl Context {
    pub fn new(ws_connections: Arc<Mutex<HashMap<usize, BroadcastSender>>>) -> Self {
        Context { ws_connections }
    }

    pub fn broadcast(&self, message: Vec<u8>) {
        let connections = self.ws_connections.lock().unwrap();
        for sender in connections.values() {
            let _ = sender.send(message.clone());
        }
    }
}
