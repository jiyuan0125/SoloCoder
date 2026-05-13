from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import Optional, Union
from uuid import uuid4

from pydantic import BaseModel, Field


class TaskStatus(str, Enum):
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"
    CANCELLED = "cancelled"


class ScheduleType(str, Enum):
    CRON = "cron"
    DELAY = "delay"


class Schedule(BaseModel):
    type: ScheduleType
    cron: Optional[str] = None
    delay_seconds: Optional[int] = None


class TaskCreate(BaseModel):
    name: str
    schedule: Union[dict, Schedule]
    callback_url: str

    def get_schedule(self) -> Schedule:
        if isinstance(self.schedule, dict):
            return Schedule(**self.schedule)
        return self.schedule


class ExecutionHistory(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    executed_at: datetime
    status: TaskStatus
    duration_ms: int
    error: Optional[str] = None


class Task(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    name: str
    schedule: Schedule
    callback_url: str
    status: TaskStatus = TaskStatus.PENDING
    created_at: datetime = Field(default_factory=datetime.now)
    last_run_at: Optional[datetime] = None
    execution_history: list[ExecutionHistory] = Field(default_factory=list)

    def can_retry(self) -> bool:
        return self.status in (TaskStatus.FAILED,)

    def can_cancel(self) -> bool:
        return self.status not in (TaskStatus.CANCELLED, TaskStatus.SUCCESS)

    def is_cron(self) -> bool:
        return self.schedule.type == ScheduleType.CRON

    def is_delay(self) -> bool:
        return self.schedule.type == ScheduleType.DELAY
