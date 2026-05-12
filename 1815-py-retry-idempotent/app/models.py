from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Any, Optional
from uuid import UUID, uuid4

from pydantic import BaseModel, Field, field_validator


class TaskStatus(str, Enum):
    PENDING = "PENDING"
    RUNNING = "RUNNING"
    RETRYING = "RETRYING"
    SUCCESS = "SUCCESS"
    FAILED = "FAILED"


class BackoffStrategy(str, Enum):
    EXPONENTIAL = "exponential"


class RetryConfig(BaseModel):
    backoff_strategy: BackoffStrategy = BackoffStrategy.EXPONENTIAL
    max_retries: int = Field(default=5, ge=0, le=50)


class CreateTaskRequest(BaseModel):
    url: str = Field(..., min_length=1)
    method: str = Field(default="GET")
    body: Optional[dict[str, Any]] = None
    idempotency_key: str = Field(..., min_length=1, max_length=128)
    retry_config: RetryConfig = Field(default_factory=RetryConfig)

    @field_validator("method")
    @classmethod
    def validate_method(cls, v: str) -> str:
        allowed = {"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
        if v.upper() not in allowed:
            raise ValueError(f"Unsupported HTTP method: {v}")
        return v.upper()

    @field_validator("idempotency_key")
    @classmethod
    def validate_idempotency_key(cls, v: str) -> str:
        if not v or not v.strip():
            raise ValueError("idempotency_key must not be empty")
        if len(v) > 128:
            raise ValueError("idempotency_key exceeds 128 characters")
        return v.strip()


class TaskResult(BaseModel):
    status_code: Optional[int] = None
    body: Optional[Any] = None
    error: Optional[str] = None


class Task(BaseModel):
    id: UUID = Field(default_factory=uuid4)
    status: TaskStatus = TaskStatus.PENDING
    url: str
    method: str
    body: Optional[dict[str, Any]] = None
    idempotency_key: str
    retry_config: RetryConfig
    current_retry: int = 0
    next_retry_at: Optional[datetime] = None
    result: Optional[TaskResult] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None


class TaskResponse(BaseModel):
    id: UUID
    status: TaskStatus
    url: str
    method: str
    idempotency_key: str
    retry_config: RetryConfig
    current_retry: int = 0
    next_retry_at: Optional[datetime] = None
    result: Optional[TaskResult] = None
    created_at: datetime
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None

    class Config:
        from_attributes = True
