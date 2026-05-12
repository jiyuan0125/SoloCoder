from enum import Enum
from typing import List, Optional
from datetime import datetime
from pydantic import BaseModel, Field


class NodeStatus(str, Enum):
    NEW = "new"
    ACTIVE = "active"
    SUSPECTED = "suspected"
    REMOVED = "removed"


class Node(BaseModel):
    id: str
    address: str
    weight: int = Field(ge=0)
    status: NodeStatus = NodeStatus.NEW
    consecutive_success_count: int = 0
    consecutive_fail_count: int = 0
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)


class NodeStats(BaseModel):
    node_id: str
    total_requests: int = 0
    total_response_time_ms: int = 0
    average_response_time_ms: float = 0.0
    updated_at: datetime = Field(default_factory=datetime.now)

    def add_request(self, response_time_ms: int):
        self.total_requests += 1
        self.total_response_time_ms += response_time_ms
        self.average_response_time_ms = self.total_response_time_ms / self.total_requests
        self.updated_at = datetime.now()


class Callback(BaseModel):
    url: str
    created_at: datetime = Field(default_factory=datetime.now)


class NodeRegistrationRequest(BaseModel):
    address: str
    weight: int = Field(gt=0)


class NodeUpdateRequest(BaseModel):
    weight: Optional[int] = Field(default=None, ge=0)
    address: Optional[str] = None


class CallbackRegistrationRequest(BaseModel):
    url: str
