from datetime import datetime
from enum import Enum
from typing import Optional, Dict, Any, List
from uuid import uuid4
from pydantic import BaseModel, Field


class TaskType(str, Enum):
    CRON = "cron"
    DELAYED = "delayed"


class TaskStatus(str, Enum):
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"
    PAUSED = "paused"
    DEAD_LETTER = "dead_letter"


class ExecutionStatus(str, Enum):
    PENDING = "pending"
    SUCCESS = "success"
    FAILED = "failed"
    TIMEOUT = "timeout"


class TaskBase(BaseModel):
    name: str = Field(..., description="任务名称，唯一")
    task_type: TaskType = Field(..., description="任务类型：cron 或 delayed")
    callback_url: Optional[str] = Field(None, description="回调 URL，为空则标记成功但不执行")
    payload: Optional[Dict[str, Any]] = Field(None, description="回调时传递的 JSON 数据")
    cron_expression: Optional[str] = Field(None, description="Cron 表达式，任务类型为 cron 时必填")
    execute_at: Optional[datetime] = Field(None, description="延迟任务的执行时间，任务类型为 delayed 时必填")


class TaskCreate(TaskBase):
    pass


class Task(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    name: str
    task_type: TaskType
    callback_url: Optional[str] = None
    payload: Optional[Dict[str, Any]] = None
    cron_expression: Optional[str] = None
    execute_at: Optional[datetime] = None
    status: TaskStatus = TaskStatus.PENDING
    is_paused: bool = False
    next_run_at: Optional[datetime] = None
    last_run_at: Optional[datetime] = None
    retry_count: int = 0
    max_retries: int = 3
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)


class TaskStatusResponse(BaseModel):
    id: str
    name: str
    task_type: TaskType
    status: TaskStatus
    is_paused: bool
    next_run_at: Optional[datetime]
    last_run_at: Optional[datetime]
    retry_count: int
    max_retries: int
    callback_url: Optional[str]
    cron_expression: Optional[str]
    execute_at: Optional[datetime]
    created_at: datetime
    updated_at: datetime


class ExecutionRecord(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    task_id: str
    task_name: str
    status: ExecutionStatus
    started_at: datetime = Field(default_factory=datetime.utcnow)
    finished_at: Optional[datetime] = None
    error_message: Optional[str] = None
    retry_attempt: int = 0
    response_status_code: Optional[int] = None


class ExecutionHistoryResponse(BaseModel):
    task_name: str
    total_count: int
    records: List[ExecutionRecord]
