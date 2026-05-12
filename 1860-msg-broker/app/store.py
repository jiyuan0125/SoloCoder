import time
import threading
import logging
from typing import Dict, List, Optional, Set, Any, Tuple
from collections import defaultdict
from dataclasses import dataclass, field
from .snowflake import generate_id

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

ACK_TIMEOUT_SECONDS = 30
MAX_RETRY = 3
RETRY_INTERVALS = [1, 2, 4]
CONSUMER_HEARTBEAT_INTERVAL = 10
CONSUMER_TIMEOUT = 20


@dataclass
class Topic:
    name: str
    capacity: int
    messages: List[Dict[str, Any]] = field(default_factory=list)
    dead_letters: List[Dict[str, Any]] = field(default_factory=list)
    group_offsets: Dict[str, Dict[str, Any]] = field(default_factory=dict)
    in_flight: Dict[int, Dict[str, Any]] = field(default_factory=dict)
    group_consumers: Dict[str, Dict[str, float]] = field(default_factory=lambda: defaultdict(dict))
    lock: threading.RLock = field(default_factory=threading.RLock)
    
    def get_group(self, group: str) -> Dict[str, Any]:
        if group not in self.group_offsets:
            self.group_offsets[group] = {
                "last_ack_id": 0,
                "total_consumed": 0,
                "message_index": 0
            }
        return self.group_offsets[group]
    
    def register_consumer(self, group: str, consumer_id: str) -> None:
        with self.lock:
            if group not in self.group_consumers:
                self.group_consumers[group] = {}
            self.group_consumers[group][consumer_id] = time.time()
    
    def heartbeat(self, group: str, consumer_id: str) -> None:
        with self.lock:
            if group in self.group_consumers and consumer_id in self.group_consumers[group]:
                self.group_consumers[group][consumer_id] = time.time()
    
    def get_active_consumers(self, group: str) -> List[str]:
        now = time.time()
        active = []
        if group not in self.group_consumers:
            return active
        for consumer_id, last_heartbeat in list(self.group_consumers[group].items()):
            if now - last_heartbeat < CONSUMER_TIMEOUT:
                active.append(consumer_id)
            else:
                self.handle_consumer_timeout(group, consumer_id)
        return active
    
    def handle_consumer_timeout(self, group: str, consumer_id: str) -> None:
        with self.lock:
            if group in self.group_consumers and consumer_id in self.group_consumers[group]:
                del self.group_consumers[group][consumer_id]
                
                for msg_id, inflight in list(self.in_flight.items()):
                    if inflight["group"] == group and inflight["consumer_id"] == consumer_id:
                        del self.in_flight[msg_id]
                        logger.info(f"Consumer {consumer_id} in group {group} timed out. Releasing message {msg_id}")


class MessageStore:
    def __init__(self):
        self.topics: Dict[str, Topic] = {}
        self.global_lock: threading.RLock = threading.RLock()
        self._message_id_counter: int = 0
        self._counter_lock: threading.Lock = threading.Lock()
    
    def create_topic(self, name: str, capacity: int) -> Topic:
        with self.global_lock:
            if name in self.topics:
                raise ValueError(f"Topic '{name}' already exists")
            topic = Topic(name=name, capacity=capacity)
            self.topics[name] = topic
            return topic
    
    def get_topic(self, name: str) -> Optional[Topic]:
        with self.global_lock:
            return self.topics.get(name)
    
    def list_topics(self) -> List[Dict[str, Any]]:
        with self.global_lock:
            return [
                {
                    "name": t.name,
                    "capacity": t.capacity,
                    "current_count": len(t.messages),
                    "deadletter_count": len(t.dead_letters)
                }
                for t in self.topics.values()
            ]
    
    def produce(self, topic_name: str, content: Any) -> int:
        topic = self.get_topic(topic_name)
        if not topic:
            raise ValueError(f"Topic '{topic_name}' does not exist")
        
        with topic.lock:
            msg_id = generate_id()
            message = {
                "id": msg_id,
                "topic": topic_name,
                "content": content,
                "created_at": time.time(),
                "retry_count": 0,
                "next_retry_at": None
            }
            
            while len(topic.messages) >= topic.capacity:
                if topic.messages:
                    dropped = topic.messages.pop(0)
                    logger.warning(f"Topic '{topic_name}' capacity exceeded. Dropping oldest message: {dropped['id']}")
            
            topic.messages.append(message)
            return msg_id
    
    def consume(self, topic_name: str, group: str, consumer_id: str, limit: int) -> List[Dict[str, Any]]:
        topic = self.get_topic(topic_name)
        if not topic:
            raise ValueError(f"Topic '{topic_name}' does not exist")
        
        limit = max(1, min(100, limit))
        
        with topic.lock:
            topic.register_consumer(group, consumer_id)
            
            group_info = topic.get_group(group)
            active_consumers = topic.get_active_consumers(group)
            if not active_consumers:
                return []
            
            try:
                consumer_idx = sorted(active_consumers).index(consumer_id)
            except ValueError:
                return []
            
            consumer_count = len(active_consumers)
            consumed = []
            now = time.time()
            in_flight_ids = set(topic.in_flight.keys())
            
            for msg in topic.messages:
                if len(consumed) >= limit:
                    break
                
                msg_id = msg["id"]
                if msg_id in in_flight_ids:
                    continue
                
                if msg["next_retry_at"] and now < msg["next_retry_at"]:
                    continue
                
                if msg_id % consumer_count != consumer_idx:
                    continue
                
                consumed.append(msg)
                
                inflight_record = {
                    "message_id": msg_id,
                    "group": group,
                    "consumer_id": consumer_id,
                    "delivered_at": now,
                    "expires_at": now + ACK_TIMEOUT_SECONDS
                }
                topic.in_flight[msg_id] = inflight_record
            
            return consumed
    
    def ack(self, topic_name: str, group: str, message_ids: List[int]) -> Tuple[int, int]:
        topic = self.get_topic(topic_name)
        if not topic:
            raise ValueError(f"Topic '{topic_name}' does not exist")
        
        acked_count = 0
        not_found_count = 0
        
        with topic.lock:
            group_info = topic.get_group(group)
            
            for msg_id in message_ids:
                if msg_id not in topic.in_flight:
                    not_found_count += 1
                    continue
                
                inflight = topic.in_flight[msg_id]
                if inflight["group"] != group:
                    not_found_count += 1
                    continue
                
                del topic.in_flight[msg_id]
                
                msg_index = None
                for i, m in enumerate(topic.messages):
                    if m["id"] == msg_id:
                        msg_index = i
                        break
                
                if msg_index is not None:
                    topic.messages.pop(msg_index)
                    group_info["last_ack_id"] = max(group_info["last_ack_id"], msg_id)
                    group_info["total_consumed"] += 1
                    acked_count += 1
                else:
                    not_found_count += 1
        
        return acked_count, not_found_count
    
    def nack(self, topic_name: str, message_id: int, reason: str = "nack") -> None:
        topic = self.get_topic(topic_name)
        if not topic:
            return
        
        with topic.lock:
            if message_id not in topic.in_flight:
                return
            
            inflight = topic.in_flight[message_id]
            del topic.in_flight[message_id]
            
            msg = None
            for m in topic.messages:
                if m["id"] == message_id:
                    msg = m
                    break
            
            if not msg:
                return
            
            msg["retry_count"] += 1
            
            if msg["retry_count"] >= MAX_RETRY:
                dead_letter = {
                    "id": msg["id"],
                    "topic": topic_name,
                    "content": msg["content"],
                    "created_at": msg["created_at"],
                    "retry_count": msg["retry_count"],
                    "moved_to_deadletter_at": time.time(),
                    "reason": reason
                }
                topic.dead_letters.append(dead_letter)
                
                for i, m in enumerate(topic.messages):
                    if m["id"] == message_id:
                        topic.messages.pop(i)
                        break
                
                logger.warning(f"Message {message_id} moved to dead letter queue after {MAX_RETRY} retries")
            else:
                retry_interval = RETRY_INTERVALS[min(msg["retry_count"] - 1, len(RETRY_INTERVALS) - 1)]
                msg["next_retry_at"] = time.time() + retry_interval
    
    def process_timeouts(self) -> None:
        for topic in list(self.topics.values()):
            with topic.lock:
                now = time.time()
                
                for group in list(topic.group_consumers.keys()):
                    topic.get_active_consumers(group)
                
                expired = []
                for msg_id, inflight in list(topic.in_flight.items()):
                    if now >= inflight["expires_at"]:
                        expired.append(msg_id)
                
                for msg_id in expired:
                    self.nack(topic.name, msg_id, reason="ack_timeout")
    
    def get_group_progress(self, topic_name: str) -> List[Dict[str, Any]]:
        topic = self.get_topic(topic_name)
        if not topic:
            return []
        
        with topic.lock:
            progress = []
            for group, offset in topic.group_offsets.items():
                in_flight_count = sum(
                    1 for inflight in topic.in_flight.values()
                    if inflight["group"] == group
                )
                progress.append({
                    "group": group,
                    "last_ack_id": offset["last_ack_id"],
                    "total_consumed": offset["total_consumed"],
                    "in_flight": in_flight_count
                })
            return progress
    
    def get_dead_letters(self, topic_name: str) -> List[Dict[str, Any]]:
        topic = self.get_topic(topic_name)
        if not topic:
            return []
        
        with topic.lock:
            return list(topic.dead_letters)
    
    def resend_dead_letter(self, topic_name: str, message_id: int) -> bool:
        topic = self.get_topic(topic_name)
        if not topic:
            return False
        
        with topic.lock:
            dead_index = None
            dead_msg = None
            for i, dl in enumerate(topic.dead_letters):
                if dl["id"] == message_id:
                    dead_index = i
                    dead_msg = dl
                    break
            
            if dead_index is None or not dead_msg:
                return False
            
            while len(topic.messages) >= topic.capacity:
                if topic.messages:
                    topic.messages.pop(0)
            
            new_msg = {
                "id": dead_msg["id"],
                "topic": topic_name,
                "content": dead_msg["content"],
                "created_at": dead_msg["created_at"],
                "retry_count": 0,
                "next_retry_at": None
            }
            topic.messages.append(new_msg)
            topic.dead_letters.pop(dead_index)
            
            logger.info(f"Dead letter message {message_id} resent to topic '{topic_name}'")
            return True


store = MessageStore()
