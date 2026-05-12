import asyncio
import heapq
import uuid
import os
import time
from collections import deque
from typing import Dict, List, Deque, Tuple, Optional
from dataclasses import dataclass, field
from fastapi import FastAPI, HTTPException, Query
from pydantic import BaseModel, Field

app = FastAPI(title="In-Process Pub/Sub Message Queue")

DEFAULT_TOPIC_CAPACITY = 10000
MAX_RETRIES = 3
RETRY_INTERVALS = [1, 2, 4]


@dataclass
class Message:
    content: str
    id: str = field(default_factory=lambda: str(uuid.uuid4()))
    created_at: float = field(default_factory=time.time)
    retry_count: int = 0


@dataclass
class ConsumerGroup:
    name: str
    pending_messages: Deque[Message] = field(default_factory=deque)
    inflight_messages: Dict[str, Message] = field(default_factory=dict)
    retry_heap: List[Tuple[float, str, Message]] = field(default_factory=list)
    retry_message_ids: set = field(default_factory=set)


@dataclass
class Topic:
    name: str
    capacity: int
    messages: Deque[Message] = field(default_factory=deque)
    consumer_groups: Dict[str, ConsumerGroup] = field(default_factory=dict)
    dead_letter_queue: Deque[Message] = field(default_factory=deque)
    lock: asyncio.Lock = field(default_factory=asyncio.Lock)


topic_manager: Dict[str, Topic] = {}
manager_lock = asyncio.Lock()


class TopicCreate(BaseModel):
    name: str
    capacity: int = Field(default=DEFAULT_TOPIC_CAPACITY, ge=1)


class MessagePublish(BaseModel):
    content: str


class AckRequest(BaseModel):
    group: str
    message_ids: List[str]


class NackRequest(BaseModel):
    group: str
    message_ids: List[str]


class RedeliverDeadLetterRequest(BaseModel):
    message_ids: List[str]


@app.post("/topics")
async def create_topic(topic_data: TopicCreate):
    async with manager_lock:
        if topic_data.name in topic_manager:
            raise HTTPException(status_code=409, detail="Topic already exists")
        topic = Topic(name=topic_data.name, capacity=topic_data.capacity)
        topic_manager[topic_data.name] = topic
        return {"name": topic.name, "capacity": topic.capacity}


async def get_or_create_topic(topic_name: str) -> Topic:
    async with manager_lock:
        if topic_name not in topic_manager:
            topic_manager[topic_name] = Topic(name=topic_name, capacity=DEFAULT_TOPIC_CAPACITY)
        return topic_manager[topic_name]


def _move_ready_retries_to_pending(group: ConsumerGroup, current_time: float):
    while group.retry_heap and group.retry_heap[0][0] <= current_time:
        retry_at, msg_id, msg = heapq.heappop(group.retry_heap)
        if msg_id in group.retry_message_ids:
            group.retry_message_ids.remove(msg_id)
            group.pending_messages.append(msg)


@app.post("/topics/{name}/publish")
async def publish_message(name: str, msg: MessagePublish):
    topic = await get_or_create_topic(name)
    message = Message(content=msg.content)
    
    async with topic.lock:
        while len(topic.messages) >= topic.capacity:
            topic.messages.popleft()
        topic.messages.append(message)
        
        for group in topic.consumer_groups.values():
            group.pending_messages.append(message)
    
    return {"message_id": message.id}


@app.get("/topics/{name}/consume")
async def consume_messages(
    name: str,
    group: str,
    limit: int = Query(default=1, ge=1, le=100)
):
    topic = await get_or_create_topic(name)
    async with topic.lock:
        if group not in topic.consumer_groups:
            consumer_group = ConsumerGroup(name=group)
            for msg in topic.messages:
                consumer_group.pending_messages.append(msg)
            topic.consumer_groups[group] = consumer_group
        else:
            consumer_group = topic.consumer_groups[group]
        
        current_time = time.time()
        _move_ready_retries_to_pending(consumer_group, current_time)
        
        messages = []
        max_messages = min(limit, len(consumer_group.pending_messages))
        
        for _ in range(max_messages):
            msg = consumer_group.pending_messages.popleft()
            consumer_group.inflight_messages[msg.id] = msg
            messages.append({
                "id": msg.id,
                "content": msg.content,
                "created_at": msg.created_at,
                "retry_count": msg.retry_count
            })
        
        return {"messages": messages}


@app.post("/topics/{name}/ack")
async def acknowledge_messages(name: str, ack: AckRequest):
    topic = await get_or_create_topic(name)
    async with topic.lock:
        if ack.group not in topic.consumer_groups:
            raise HTTPException(status_code=404, detail="Consumer group not found")
        
        consumer_group = topic.consumer_groups[ack.group]
        results = []
        
        for msg_id in ack.message_ids:
            if msg_id in consumer_group.inflight_messages:
                del consumer_group.inflight_messages[msg_id]
                results.append({"message_id": msg_id, "status": "acked"})
            else:
                results.append({"message_id": msg_id, "status": "not_found"})
        
        return {"results": results}


@app.post("/topics/{name}/nack")
async def nacknowledge_messages(name: str, nack: NackRequest):
    topic = await get_or_create_topic(name)
    async with topic.lock:
        if nack.group not in topic.consumer_groups:
            raise HTTPException(status_code=404, detail="Consumer group not found")
        
        consumer_group = topic.consumer_groups[nack.group]
        current_time = time.time()
        results = []
        
        for msg_id in nack.message_ids:
            if msg_id in consumer_group.inflight_messages:
                msg = consumer_group.inflight_messages.pop(msg_id)
                msg.retry_count += 1
                
                if msg.retry_count >= MAX_RETRIES:
                    topic.dead_letter_queue.append(msg)
                    results.append({"message_id": msg_id, "status": "dead_letter"})
                else:
                    retry_idx = min(msg.retry_count - 1, len(RETRY_INTERVALS) - 1)
                    retry_interval = RETRY_INTERVALS[retry_idx]
                    retry_at = current_time + retry_interval
                    
                    heapq.heappush(consumer_group.retry_heap, (retry_at, msg.id, msg))
                    consumer_group.retry_message_ids.add(msg.id)
                    
                    results.append({
                        "message_id": msg_id, 
                        "status": "retried",
                        "retry_count": msg.retry_count,
                        "next_retry_at": retry_at
                    })
            else:
                results.append({"message_id": msg_id, "status": "not_found"})
        
        return {"results": results}


@app.get("/topics")
async def list_topics():
    async with manager_lock:
        topics_info = []
        for name, topic in topic_manager.items():
            async with topic.lock:
                consumer_groups_info = []
                for group_name, group in topic.consumer_groups.items():
                    consumer_groups_info.append({
                        "name": group_name,
                        "pending": len(group.pending_messages),
                        "inflight": len(group.inflight_messages),
                        "retries": len(group.retry_heap)
                    })
                
                topics_info.append({
                    "name": name,
                    "capacity": topic.capacity,
                    "total_messages": len(topic.messages),
                    "consumer_groups": consumer_groups_info,
                    "dead_letter_count": len(topic.dead_letter_queue)
                })
        
        return {"topics": topics_info}


@app.get("/topics/{name}/deadletter")
async def list_dead_letters(name: str):
    topic = await get_or_create_topic(name)
    async with topic.lock:
        dead_letters = [{
            "id": msg.id,
            "content": msg.content,
            "created_at": msg.created_at,
            "retry_count": msg.retry_count
        } for msg in topic.dead_letter_queue]
        
        return {"dead_letters": dead_letters}


@app.post("/topics/{name}/deadletter/redeliver")
async def redeliver_dead_letters(name: str, request: RedeliverDeadLetterRequest):
    topic = await get_or_create_topic(name)
    async with topic.lock:
        results = []
        remaining = []
        redelivered_ids = set(request.message_ids)
        
        for msg in topic.dead_letter_queue:
            if msg.id in redelivered_ids:
                msg.retry_count = 0
                topic.messages.append(msg)
                
                for group in topic.consumer_groups.values():
                    group.pending_messages.append(msg)
                
                results.append({"message_id": msg.id, "status": "redelivered"})
            else:
                remaining.append(msg)
        
        topic.dead_letter_queue = deque(remaining)
        
        for msg_id in redelivered_ids:
            if not any(r["message_id"] == msg_id for r in results):
                results.append({"message_id": msg_id, "status": "not_found"})
        
        return {"results": results}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
