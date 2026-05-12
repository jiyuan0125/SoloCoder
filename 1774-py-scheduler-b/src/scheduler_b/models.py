from __future__ import annotations
from datetime import datetime
from typing import Any, Optional
from enum import Enum

from pydantic import BaseModel, Field


class TaskStatus(str, Enum):
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"
    PAUSED = "paused"
    CANCELLED = "cancelled"


class CallbackType(str, Enum):
    HTTP = "http"
    WEBHOOK = "webhook"


class TaskCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    cron_expression: str = Field(..., description="Cron 表达式，如 '0 * * * *'")
    callback_url: str = Field(..., description="任务执行完成后的回调 URL")
    callback_type: CallbackType = CallbackType.HTTP
    callback_headers: Optional[dict[str, str]] = None
    callback_method: str = "POST"
    max_retries: int = Field(default=3, ge=0, le=10)
    retry_delay_seconds: int = Field(default=60, ge=1)
    timeout_seconds: int = Field(default=300, ge=1)
    description: Optional[str] = Field(default=None, max_length=500)
    tags: Optional[list[str]] = None
    enabled: bool = Field(default=True)


class TaskUpdate(BaseModel):
    name: Optional[str] = Field(default=None, min_length=1, max_length=100)
    cron_expression: Optional[str] = None
    callback_url: Optional[str] = None
    callback_type: Optional[CallbackType] = None
    callback_headers: Optional[dict[str, str]] = None
    callback_method: Optional[str] = None
    max_retries: Optional[int] = Field(default=None, ge=0, le=10)
    retry_delay_seconds: Optional[int] = Field(default=None, ge=1)
    timeout_seconds: Optional[int] = Field(default=None, ge=1)
    description: Optional[str] = Field(default=None, max_length=500)
    tags: Optional[list[str]] = None
    enabled: Optional[bool] = None


class Task(BaseModel):
    id: str
    name: str
    cron_expression: str
    callback_url: str
    callback_type: CallbackType
    callback_headers: Optional[dict[str, str]]
    callback_method: str
    max_retries: int
    retry_delay_seconds: int
    timeout_seconds: int
    description: Optional[str]
    tags: list[str]
    enabled: bool
    status: TaskStatus
    created_at: datetime
    updated_at: datetime
    last_run_at: Optional[datetime] = None
    next_run_at: Optional[datetime] = None
    last_success_at: Optional[datetime] = None
    last_failure_at: Optional[datetime] = None
    total_runs: int = 0
    total_successes: int = 0
    total_failures: int = 0
    current_retry_count: int = 0


class ExecutionLog(BaseModel):
    id: str
    task_id: str
    started_at: datetime
    finished_at: Optional[datetime] = None
    status: TaskStatus
    result: Optional[Any] = None
    error: Optional[str] = None
    duration_seconds: Optional[float] = None
    try_count: int = 0


class TaskWithLogs(Task):
    recent_logs: list[ExecutionLog] = []


class CronValidationResponse(BaseModel):
    valid: bool
    message: str
    next_runs: Optional[list[datetime]] = None
