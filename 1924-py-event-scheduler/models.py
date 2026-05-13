from enum import Enum
from typing import Optional, List, Dict, Any
from datetime import datetime
from pydantic import BaseModel, Field


class TaskStatus(str, Enum):
    PENDING = "pending"
    WAITING_DEPENDS = "waiting_depends"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"
    TIMEOUT = "timeout"


class ScheduleType(str, Enum):
    CRON = "cron"
    DELAY = "delay"


class ExecutorType(str, Enum):
    HTTP = "http"
    COMMAND = "command"


class Schedule(BaseModel):
    type: ScheduleType
    cron_expression: Optional[str] = None
    delay_seconds: Optional[int] = None


class TaskCreate(BaseModel):
    name: str
    schedule: Schedule
    executor_type: ExecutorType
    executor_config: str
    depends_on: List[str] = Field(default_factory=list)
    timeout_seconds: int = 300


class Task(BaseModel):
    id: str
    name: str
    schedule: Schedule
    executor_type: ExecutorType
    executor_config: str
    depends_on: List[str]
    timeout_seconds: int
    created_at: datetime
    canceled: bool = False


class ExecutionInstance(BaseModel):
    id: str
    task_id: str
    status: TaskStatus
    started_at: Optional[datetime] = None
    finished_at: Optional[datetime] = None
    error_message: Optional[str] = None
    depends_on_status: Dict[str, TaskStatus] = Field(default_factory=dict)
    exit_code: Optional[int] = None
    stdout: Optional[str] = None
    stderr: Optional[str] = None
    http_status_code: Optional[int] = None
