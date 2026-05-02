use std::collections::HashMap;
use std::time::Instant;
use crate::packet::QoS;

#[derive(Debug, Clone)]
pub struct WillMessage {
    pub topic: String,
    pub payload: Vec<u8>,
    pub qos: QoS,
    pub retain: bool,
}

#[derive(Debug, Clone)]
pub struct Subscription {
    pub topic_filter: String,
    pub qos: QoS,
}

#[derive(Debug, Clone)]
pub struct Session {
    pub client_id: String,
    pub clean_session: bool,
    pub subscriptions: Vec<Subscription>,
    pub will_message: Option<WillMessage>,
    pub last_activity: Instant,
    pub keep_alive: u16,
    pub connected: bool,
}

impl Session {
    pub fn new(client_id: String, clean_session: bool, keep_alive: u16) -> Self {
        Session {
            client_id,
            clean_session,
            subscriptions: Vec::new(),
            will_message: None,
            last_activity: Instant::now(),
            keep_alive,
            connected: true,
        }
    }

    pub fn add_subscription(&mut self, topic_filter: String, qos: QoS) {
        for sub in &mut self.subscriptions {
            if sub.topic_filter == topic_filter {
                sub.qos = qos;
                return;
            }
        }
        self.subscriptions.push(Subscription { topic_filter, qos });
    }

    pub fn remove_subscription(&mut self, topic_filter: &str) {
        self.subscriptions.retain(|s| s.topic_filter != topic_filter);
    }

    pub fn update_activity(&mut self) {
        self.last_activity = Instant::now();
    }

    pub fn is_timeout(&self) -> bool {
        let timeout_secs = if self.keep_alive > 0 {
            (self.keep_alive as f64 * 1.5) as u64
        } else {
            90
        };
        self.last_activity.elapsed().as_secs() > timeout_secs
    }
}

#[derive(Debug, Clone)]
pub struct RetainedMessage {
    pub topic: String,
    pub payload: Vec<u8>,
    pub qos: QoS,
}

#[derive(Debug, Default)]
pub struct SessionManager {
    sessions: HashMap<String, Session>,
    retained_messages: HashMap<String, RetainedMessage>,
}

impl SessionManager {
    pub fn new() -> Self {
        SessionManager::default()
    }

    pub fn get_or_create_session(
        &mut self,
        client_id: String,
        clean_session: bool,
        keep_alive: u16,
    ) -> (bool, &mut Session) {
        let session_present = !clean_session && self.sessions.contains_key(&client_id);
        
        if session_present {
            let session = self.sessions.get_mut(&client_id).unwrap();
            session.connected = true;
            session.keep_alive = keep_alive;
            session.update_activity();
            (true, session)
        } else {
            let session = Session::new(client_id.clone(), clean_session, keep_alive);
            self.sessions.insert(client_id.clone(), session);
            let session = self.sessions.get_mut(&client_id).unwrap();
            (false, session)
        }
    }

    pub fn get_session(&mut self, client_id: &str) -> Option<&mut Session> {
        self.sessions.get_mut(client_id)
    }

    pub fn get_session_ref(&self, client_id: &str) -> Option<&Session> {
        self.sessions.get(client_id)
    }

    pub fn disconnect_client(&mut self, client_id: &str, send_will: bool) -> Option<WillMessage> {
        let will = if send_will {
            self.sessions.get(client_id).and_then(|s| s.will_message.clone())
        } else {
            None
        };

        if let Some(session) = self.sessions.get_mut(client_id) {
            session.connected = false;
            if session.clean_session {
                self.sessions.remove(client_id);
            } else {
                session.will_message = None;
            }
        }

        will
    }

    pub fn set_will_message(&mut self, client_id: &str, will: WillMessage) {
        if let Some(session) = self.sessions.get_mut(client_id) {
            session.will_message = Some(will);
        }
    }

    pub fn add_subscription(&mut self, client_id: &str, topic_filter: String, qos: QoS) {
        if let Some(session) = self.sessions.get_mut(client_id) {
            session.add_subscription(topic_filter, qos);
        }
    }

    pub fn remove_subscription(&mut self, client_id: &str, topic_filter: &str) {
        if let Some(session) = self.sessions.get_mut(client_id) {
            session.remove_subscription(topic_filter);
        }
    }

    pub fn get_matching_subscribers(
        &self,
        topic: &str,
    ) -> Vec<(&str, QoS)> {
        let mut result = Vec::new();
        
        for (client_id, session) in &self.sessions {
            if !session.connected {
                continue;
            }
            for sub in &session.subscriptions {
                if crate::topic::matches(&sub.topic_filter, topic) {
                    result.push((client_id.as_str(), sub.qos));
                    break;
                }
            }
        }
        
        result
    }

    pub fn set_retained_message(
        &mut self,
        topic: String,
        payload: Vec<u8>,
        qos: QoS,
    ) {
        if payload.is_empty() {
            self.retained_messages.remove(&topic);
        } else {
            self.retained_messages.insert(
                topic.clone(),
                RetainedMessage { topic, payload, qos },
            );
        }
    }

    pub fn get_retained_messages_for_filter(
        &self,
        topic_filter: &str,
    ) -> Vec<&RetainedMessage> {
        self.retained_messages
            .values()
            .filter(|msg| crate::topic::matches(topic_filter, &msg.topic))
            .collect()
    }

    pub fn update_activity(&mut self, client_id: &str) {
        if let Some(session) = self.sessions.get_mut(client_id) {
            session.update_activity();
        }
    }

    pub fn get_timeout_clients(&self) -> Vec<String> {
        self.sessions
            .iter()
            .filter(|(_, session)| session.connected && session.is_timeout())
            .map(|(client_id, _)| client_id.clone())
            .collect()
    }
}
