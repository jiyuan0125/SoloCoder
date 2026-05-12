from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Optional
from uuid import UUID, uuid4

from pydantic import BaseModel, Field


class ProbeType(str, Enum):
    http = "http"
    tcp = "tcp"


class HealthStatus(str, Enum):
    unknown = "unknown"
    healthy = "healthy"
    unhealthy = "unhealthy"


class TargetCreate(BaseModel):
    type: ProbeType
    address: str
    interval: int = Field(..., gt=0)
    timeout: int = Field(..., gt=0)
    failure_threshold: int = Field(..., gt=0)
    recovery_threshold: int = Field(..., gt=0)
    group_id: Optional[UUID] = None


class Target(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    type: ProbeType
    address: str
    interval: int
    timeout: int
    failure_threshold: int
    recovery_threshold: int
    group_id: Optional[UUID] = None
    status: HealthStatus = HealthStatus.unknown
    consecutive_failures: int = 0
    consecutive_successes: int = 0
    created_at: datetime = Field(default_factory=datetime.utcnow)


class GroupCreate(BaseModel):
    name: str
    parent_id: Optional[UUID] = None


class Group(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    name: str
    parent_id: Optional[UUID] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)


class GroupStatus(Group):
    status: HealthStatus
    sub_groups: list[GroupStatus] = []


class CallbackCreate(BaseModel):
    url: str


class Callback(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    url: str
    created_at: datetime = Field(default_factory=datetime.utcnow)


class StateChangeLog(BaseModel):
    timestamp: datetime = Field(default_factory=datetime.utcnow)
    target_id: UUID
    target_address: str
    old_status: HealthStatus
    new_status: HealthStatus
