import asyncio
import time
import json
from typing import Dict, List, Optional, Any
from models import Topic, Message, ConsumerGroup
from filter import evaluate_filter


class MessageBroker:
    def __init__(self):
        self.topics: Dict[str, Topic] = {}
        self.retry_interval = 10
        self.max_retries = 3
        self._retry_task = None

    async def start(self):
        self._retry_task = asyncio.create_task(self._process_retries())

    async def stop(self):
        if self._retry_task:
            self._retry_task.cancel()
            try:
                await self._retry_task
            except asyncio.CancelledError:
                pass

    def create_topic(self, name: str, creator: str, max_size: int = 1000) -> Topic:
        if name in self.topics:
            raise ValueError(f"Topic '{name}' already exists")
        topic = Topic(name=name, max_size=max_size, creator=creator)
        self.topics[name] = topic
        return topic

    def get_topic(self, name: str) -> Optional[Topic]:
        return self.topics.get(name)

    def list_topics(self) -> List[str]:
        return list(self.topics.keys())

    async def publish_message(self, topic_name: str, content: str, priority: int = 0) -> Message:
        topic = self.get_topic(topic_name)
        if not topic:
            raise ValueError(f"Topic '{topic_name}' does not exist")
        
        async with topic.lock:
            message_id = topic.next_message_id
            topic.next_message_id += 1
            
            message = Message(
                id=message_id,
                content=content,
                priority=priority,
                created_at=time.time()
            )
            
            topic.messages.append(message)
            
            if len(topic.messages) > topic.max_size:
                removed = topic.messages.pop(0)
                
                for group in topic.groups.values():
                    if removed in group.messages:
                        group.messages.remove(removed)
                        if removed.id > group.offset:
                            group.offset = removed.id
            
            for group in topic.groups.values():
                if self._should_include_message(group, message):
                    group.messages.append(message)
            
            return message

    def _should_include_message(self, group: ConsumerGroup, message: Message) -> bool:
        if group.filter_expr:
            return evaluate_filter(group.filter_expr, message.content)
        return True

    async def subscribe(self, topic_name: str, group_id: str, filter_expr: Optional[str] = None) -> ConsumerGroup:
        topic = self.get_topic(topic_name)
        if not topic:
            raise ValueError(f"Topic '{topic_name}' does not exist")
        
        async with topic.lock:
            if group_id in topic.groups:
                if filter_expr:
                    topic.groups[group_id].filter_expr = filter_expr
                return topic.groups[group_id]
            
            group = ConsumerGroup(
                id=group_id,
                offset=0,
                filter_expr=filter_expr
            )
            
            for message in topic.messages:
                if self._should_include_message(group, message):
                    group.messages.append(message)
            
            topic.groups[group_id] = group
            return group

    async def consume_message(self, topic_name: str, group_id: str) -> Optional[Message]:
        topic = self.get_topic(topic_name)
        if not topic:
            raise ValueError(f"Topic '{topic_name}' does not exist")
        
        if group_id not in topic.groups:
            raise ValueError(f"Consumer group '{group_id}' not subscribed to topic '{topic_name}'")
        
        group = topic.groups[group_id]
        
        async with topic.lock:
            pending_messages = [
                m for m in group.messages 
                if m.id not in group.consumed and m.retry_count < self.max_retries
            ]
            
            if not pending_messages:
                return None
            
            pending_messages.sort(key=lambda m: (-m.priority, m.id))
            message = pending_messages[0]
            
            group.consumed.add(message.id)
            if message.id > group.offset:
                group.offset = message.id
            
            return message

    async def get_offset(self, topic_name: str, group_id: str) -> int:
        topic = self.get_topic(topic_name)
        if not topic:
            raise ValueError(f"Topic '{topic_name}' does not exist")
        
        if group_id not in topic.groups:
            raise ValueError(f"Consumer group '{group_id}' not subscribed to topic '{topic_name}'")
        
        return topic.groups[group_id].offset

    async def mark_message_failed(self, topic_name: str, message_id: int) -> bool:
        topic = self.get_topic(topic_name)
        if not topic:
            return False
        
        async with topic.lock:
            for message in topic.messages:
                if message.id == message_id:
                    message.retry_count += 1
                    message.last_retry_at = time.time()
                    
                    if message.retry_count >= self.max_retries:
                        topic.dead_letter.append(message)
                        
                        for group in topic.groups.values():
                            if message in group.messages:
                                group.messages.remove(message)
                        
                        if message in topic.messages:
                            topic.messages.remove(message)
                    
                    return True
        return False

    async def get_dead_letter(self, topic_name: str) -> List[Message]:
        topic = self.get_topic(topic_name)
        if not topic:
            raise ValueError(f"Topic '{topic_name}' does not exist")
        return topic.dead_letter

    async def _process_retries(self):
        while True:
            try:
                await asyncio.sleep(self.retry_interval)
                
                for topic in self.topics.values():
                    async with topic.lock:
                        for group in topic.groups.values():
                            for message in group.messages:
                                if message.retry_count > 0 and message.retry_count < self.max_retries:
                                    if time.time() - message.last_retry_at >= self.retry_interval:
                                        pass
            except asyncio.CancelledError:
                break
            except Exception:
                pass

    def get_message(self, topic_name: str, message_id: int) -> Optional[Message]:
        topic = self.get_topic(topic_name)
        if not topic:
            return None
        
        for message in topic.messages + topic.dead_letter:
            if message.id == message_id:
                return message
        return None

    def is_authorized(self, topic_name: str, user: str, is_admin: bool = False) -> bool:
        topic = self.get_topic(topic_name)
        if not topic:
            return False
        return is_admin or topic.creator == user
