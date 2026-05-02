use std::net::TcpStream;
use std::io::{self, Write, ErrorKind};
use std::sync::{Arc, Mutex};
use std::collections::HashMap;
use std::sync::mpsc::{self, Sender, Receiver};
use std::thread;
use std::time::Duration;

use crate::packet::{
    self, Packet, FixedHeader, PacketType, QoS,
    ConnackPacket, ConnectPacket, PublishPacket, PubackPacket,
    SubscribePacket, SubackPacket, UnsubscribePacket,
};
use crate::session::{SessionManager, WillMessage};

pub enum BrokerMessage {
    Publish {
        topic: String,
        payload: Vec<u8>,
        qos: QoS,
        retain: bool,
    },
    PublishToClient {
        client_id: String,
        topic: String,
        payload: Vec<u8>,
        qos: QoS,
        retain: bool,
    },
    Disconnect {
        client_id: String,
        send_will: bool,
    },
}

pub struct Broker {
    session_manager: SessionManager,
    client_senders: HashMap<String, Sender<BrokerMessage>>,
    next_packet_id: u16,
}

impl Broker {
    pub fn new() -> Self {
        Broker {
            session_manager: SessionManager::new(),
            client_senders: HashMap::new(),
            next_packet_id: 1,
        }
    }

    pub fn get_next_packet_id(&mut self) -> u16 {
        let id = self.next_packet_id;
        self.next_packet_id = self.next_packet_id.wrapping_add(1);
        if self.next_packet_id == 0 {
            self.next_packet_id = 1;
        }
        id
    }

    pub fn register_client(&mut self, client_id: String, sender: Sender<BrokerMessage>) {
        self.client_senders.insert(client_id, sender);
    }

    pub fn unregister_client(&mut self, client_id: &str) {
        self.client_senders.remove(client_id);
    }

    pub fn handle_publish(
        &mut self,
        topic: &str,
        payload: &[u8],
        qos: QoS,
        retain: bool,
    ) {
        if retain {
            self.session_manager.set_retained_message(
                topic.to_string(),
                payload.to_vec(),
                qos,
            );
        }

        let subscribers = self.session_manager.get_matching_subscribers(topic);
        
        for (client_id, sub_qos) in subscribers {
            let effective_qos = if qos as u8 > sub_qos as u8 {
                sub_qos
            } else {
                qos
            };
            
            if let Some(sender) = self.client_senders.get(client_id) {
                let _ = sender.send(BrokerMessage::PublishToClient {
                    client_id: client_id.to_string(),
                    topic: topic.to_string(),
                    payload: payload.to_vec(),
                    qos: effective_qos,
                    retain: false,
                });
            }
        }
    }

    pub fn handle_will(&mut self, will: &WillMessage) {
        self.handle_publish(
            &will.topic,
            &will.payload,
            will.qos,
            will.retain,
        );
    }

    pub fn send_retained_messages(
        &mut self,
        client_id: &str,
        topic_filter: &str,
    ) {
        let retained = self.session_manager.get_retained_messages_for_filter(topic_filter);
        
        for msg in retained {
            if let Some(sender) = self.client_senders.get(client_id) {
                let _ = sender.send(BrokerMessage::PublishToClient {
                    client_id: client_id.to_string(),
                    topic: msg.topic.clone(),
                    payload: msg.payload.clone(),
                    qos: msg.qos,
                    retain: true,
                });
            }
        }
    }
}

struct ClientHandler {
    stream: TcpStream,
    broker: Arc<Mutex<Broker>>,
    client_id: Option<String>,
    rx: Receiver<BrokerMessage>,
    connected: bool,
}

impl ClientHandler {
    fn new(stream: TcpStream, broker: Arc<Mutex<Broker>>, rx: Receiver<BrokerMessage>) -> Self {
        ClientHandler {
            stream,
            broker,
            client_id: None,
            rx,
            connected: true,
        }
    }

    fn run(&mut self) {
        self.stream.set_read_timeout(Some(Duration::from_secs(1))).ok();
        
        while self.connected {
            match self.read_and_handle_packet() {
                Ok(()) => {}
                Err(e) if e.kind() == ErrorKind::WouldBlock || e.kind() == ErrorKind::TimedOut => {
                    self.check_timeout();
                    self.process_outgoing_messages();
                    continue;
                }
                Err(_) => {
                    break;
                }
            }
            self.process_outgoing_messages();
        }

        if let Some(ref client_id) = self.client_id {
            let will = {
                let mut broker = self.broker.lock().unwrap();
                broker.unregister_client(client_id);
                broker.session_manager.disconnect_client(client_id, true)
            };
            
            if let Some(will) = will {
                let mut broker = self.broker.lock().unwrap();
                broker.handle_will(&will);
            }
        }
    }

    fn check_timeout(&mut self) {
        if let Some(ref client_id) = self.client_id {
            let is_timeout = {
                let broker = self.broker.lock().unwrap();
                if let Some(session) = broker.session_manager.get_session_ref(client_id) {
                    session.is_timeout()
                } else {
                    false
                }
            };
            
            if is_timeout {
                self.connected = false;
            }
        }
    }

    fn process_outgoing_messages(&mut self) {
        loop {
            match self.rx.try_recv() {
                Ok(msg) => {
                    match msg {
                        BrokerMessage::PublishToClient { topic, payload, qos, retain, .. } => {
                            let packet_id = if qos != QoS::AtMostOnce {
                                let mut broker = self.broker.lock().unwrap();
                                Some(broker.get_next_packet_id())
                            } else {
                                None
                            };
                            
                            if let Ok(pkt) = packet::encode_publish_packet(
                                &topic,
                                &payload,
                                qos,
                                retain,
                                false,
                                packet_id,
                            ) {
                                let _ = self.stream.write_all(&pkt);
                                let _ = self.stream.flush();
                            }
                        }
                        BrokerMessage::Disconnect { .. } => {
                            self.connected = false;
                        }
                        _ => {}
                    }
                }
                Err(mpsc::TryRecvError::Empty) => break,
                Err(mpsc::TryRecvError::Disconnected) => {
                    self.connected = false;
                    break;
                }
            }
        }
    }

    fn read_and_handle_packet(&mut self) -> io::Result<()> {
        let (fixed_header, packet) = packet::read_packet(&mut self.stream)?;
        
        if let Some(ref client_id) = self.client_id {
            let mut broker = self.broker.lock().unwrap();
            broker.session_manager.update_activity(client_id);
        }

        match packet {
            Packet::Connect(connect) => {
                self.handle_connect(connect)?;
            }
            Packet::Publish(publish) => {
                self.handle_publish(fixed_header, publish)?;
            }
            Packet::Puback(puback) => {
                self.handle_puback(puback);
            }
            Packet::Subscribe(subscribe) => {
                self.handle_subscribe(subscribe)?;
            }
            Packet::Unsubscribe(unsubscribe) => {
                self.handle_unsubscribe(unsubscribe)?;
            }
            Packet::Pingreq => {
                self.handle_pingreq()?;
            }
            Packet::Disconnect => {
                self.handle_disconnect()?;
            }
            _ => {}
        }

        Ok(())
    }

    fn handle_connect(&mut self, connect: ConnectPacket) -> io::Result<()> {
        if self.client_id.is_some() {
            return Err(io::Error::new(ErrorKind::InvalidData, "Already connected"));
        }

        if connect.protocol_name != "MQTT" || connect.protocol_level != 4 {
            let connack = ConnackPacket {
                session_present: false,
                return_code: 1,
            };
            let pkt = packet::encode_connack_packet(&connack)?;
            self.stream.write_all(&pkt)?;
            self.stream.flush()?;
            self.connected = false;
            return Ok(());
        }

        let client_id = if connect.client_id.is_empty() {
            if connect.connect_flags.clean_session {
                use std::time::{SystemTime, UNIX_EPOCH};
                let nanos = SystemTime::now()
                    .duration_since(UNIX_EPOCH)
                    .map(|d| d.as_nanos())
                    .unwrap_or(0);
                format!("anon_{}", nanos)
            } else {
                let connack = ConnackPacket {
                    session_present: false,
                    return_code: 2,
                };
                let pkt = packet::encode_connack_packet(&connack)?;
                self.stream.write_all(&pkt)?;
                self.stream.flush()?;
                self.connected = false;
                return Ok(());
            }
        } else {
            connect.client_id
        };

        let (tx, rx) = mpsc::channel();
        self.rx = rx;

        let (session_present, subscriptions) = {
            let mut broker = self.broker.lock().unwrap();
            
            let old_subs = if !connect.connect_flags.clean_session {
                if let Some(existing) = broker.session_manager.get_session(&client_id) {
                    if !existing.clean_session {
                        existing.subscriptions.clone()
                    } else {
                        Vec::new()
                    }
                } else {
                    Vec::new()
                }
            } else {
                Vec::new()
            };

            let (present, _) = broker.session_manager.get_or_create_session(
                client_id.clone(),
                connect.connect_flags.clean_session,
                connect.keep_alive,
            );

            if connect.connect_flags.will_flag {
                if let (Some(topic), Some(msg)) = (connect.will_topic, connect.will_message) {
                    let will = WillMessage {
                        topic,
                        payload: msg,
                        qos: connect.connect_flags.will_qos,
                        retain: connect.connect_flags.will_retain,
                    };
                    broker.session_manager.set_will_message(&client_id, will);
                }
            }

            broker.register_client(client_id.clone(), tx);

            (present, old_subs)
        };

        self.client_id = Some(client_id.clone());

        let connack = ConnackPacket {
            session_present,
            return_code: 0,
        };
        let pkt = packet::encode_connack_packet(&connack)?;
        self.stream.write_all(&pkt)?;
        self.stream.flush()?;

        for sub in subscriptions {
            let mut broker = self.broker.lock().unwrap();
            broker.send_retained_messages(&client_id, &sub.topic_filter);
        }

        Ok(())
    }

    fn handle_publish(&mut self, fixed_header: FixedHeader, publish: PublishPacket) -> io::Result<()> {
        if self.client_id.is_none() {
            return Err(io::Error::new(ErrorKind::InvalidData, "Not connected"));
        }

        if fixed_header.qos != QoS::AtMostOnce {
            if let Some(packet_id) = publish.packet_identifier {
                let puback = packet::encode_puback_packet(packet_id)?;
                self.stream.write_all(&puback)?;
                self.stream.flush()?;
            }
        }

        {
            let mut broker = self.broker.lock().unwrap();
            broker.handle_publish(
                &publish.topic_name,
                &publish.payload,
                fixed_header.qos,
                fixed_header.retain,
            );
        }

        Ok(())
    }

    fn handle_puback(&mut self, _puback: PubackPacket) {
    }

    fn handle_subscribe(&mut self, subscribe: SubscribePacket) -> io::Result<()> {
        if let Some(ref client_id) = self.client_id {
            let mut return_codes = Vec::new();
            let topics_to_send_retained: Vec<String> = {
                let mut broker = self.broker.lock().unwrap();
                let mut topics = Vec::new();
                
                for topic in &subscribe.topics {
                    if crate::topic::validate_topic_filter(&topic.topic_filter) {
                        broker.session_manager.add_subscription(
                            client_id,
                            topic.topic_filter.clone(),
                            topic.qos,
                        );
                        return_codes.push(topic.qos as u8);
                        topics.push(topic.topic_filter.clone());
                    } else {
                        return_codes.push(0x80);
                    }
                }
                
                topics
            };

            let suback = packet::encode_suback_packet(subscribe.packet_identifier, &return_codes)?;
            self.stream.write_all(&suback)?;
            self.stream.flush()?;

            for topic_filter in topics_to_send_retained {
                let mut broker = self.broker.lock().unwrap();
                broker.send_retained_messages(client_id, &topic_filter);
            }
        }

        Ok(())
    }

    fn handle_unsubscribe(&mut self, unsubscribe: UnsubscribePacket) -> io::Result<()> {
        if let Some(ref client_id) = self.client_id {
            let mut broker = self.broker.lock().unwrap();
            
            for topic in &unsubscribe.topics {
                broker.session_manager.remove_subscription(client_id, topic);
            }
        }

        let unsuback = packet::encode_unsuback_packet(unsubscribe.packet_identifier)?;
        self.stream.write_all(&unsuback)?;
        self.stream.flush()?;

        Ok(())
    }

    fn handle_pingreq(&mut self) -> io::Result<()> {
        let pingresp = packet::encode_pingresp_packet()?;
        self.stream.write_all(&pingresp)?;
        self.stream.flush()
    }

    fn handle_disconnect(&mut self) -> io::Result<()> {
        if let Some(ref client_id) = self.client_id {
            let mut broker = self.broker.lock().unwrap();
            broker.unregister_client(client_id);
            broker.session_manager.disconnect_client(client_id, false);
        }
        self.connected = false;
        Ok(())
    }
}

pub fn handle_client(stream: TcpStream, broker: Arc<Mutex<Broker>>) {
    let (_tx, rx) = mpsc::channel();
    
    let mut handler = ClientHandler::new(stream, broker, rx);
    handler.run();
}
