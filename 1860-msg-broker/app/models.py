from pydantic import BaseModel, Field
from typing import List, Optional, Any
from datetime import datetime


class Message(BaseModel):
    id: int
    topic: str
    content: Any
    created_at: float
    retry_count: int = 0
    next_retry_at: Optional[float] = None


class DeadLetterMessage(BaseModel):
    id: int
    topic: str
    content: Any
    created_at: float
    retry_count: int
    moved_to_deadletter_at: float
    reason: str


class InFlightMessage(BaseModel):
    message_id: int
    group: str
    consumer_id: str
    delivered_at: float
    expires_at: float


class GroupOffset(BaseModel):
    group: str
    last_ack_id: int = 0
    total_consumed: int = 0


class TopicInfo(BaseModel):
    name: str
    capacity: int
    current_count: int
    deadletter_count: int


class TopicDetail(BaseModel):
    name: str
    capacity: int
    current_count: int
    groups: List[str]


class ConsumerProgress(BaseModel):
    group: str
    last_ack_id: int
    total_consumed: int
    in_flight: int


class CreateTopicRequest(BaseModel):
    capacity: int = Field(..., ge=1, le=1000000)


class ProduceMessageRequest(BaseModel):
    content: Any


class AckRequest(BaseModel):
    message_ids: List[int]


class ProduceResponse(BaseModel):
    id: int


class ConsumeResponse(BaseModel):
    messages: List[dict]


class TopicListResponse(BaseModel):
    topics: List[TopicInfo]


class GroupsResponse(BaseModel):
    groups: List[ConsumerProgress]


class DeadLetterResponse(BaseModel):
    messages: List[dict]
