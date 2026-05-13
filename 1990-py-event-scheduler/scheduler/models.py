from datetime import datetime
from enum import Enum
from typing import Any, Dict, Optional

from pydantic import BaseModel, Field
from sqlalchemy import (
    JSON,
    Column,
    DateTime,
    Enum as SAEnum,
    Integer,
    String,
    Text,
    create_engine,
)
from sqlalchemy.orm import DeclarativeBase, sessionmaker

DATABASE_URL = "sqlite:///./scheduler.db"
engine = create_engine(DATABASE_URL, connect_args={"check_same_thread": False})
SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)


class Base(DeclarativeBase):
    pass


class TaskType(str, Enum):
    CRON = "cron"
    DELAYED = "delayed"


class TaskStatus(str, Enum):
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"


class Task(Base):
    __tablename__ = "tasks"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), nullable=False)
    task_type = Column(SAEnum(TaskType), nullable=False)
    cron_expression = Column(String(100), nullable=True)
    delay_seconds = Column(Integer, nullable=True)
    callback_url = Column(String(500), nullable=False)
    callback_payload = Column(JSON, nullable=False, default=dict)
    timeout_seconds = Column(Integer, default=60)
    max_retries = Column(Integer, default=3)
    retry_count = Column(Integer, default=0)
    retry_backoff = Column(JSON, default=lambda: [10, 30, 60])
    status = Column(SAEnum(TaskStatus), default=TaskStatus.PENDING)
    next_run_at = Column(DateTime, nullable=True)
    last_run_at = Column(DateTime, nullable=True)
    queue_size = Column(Integer, default=0)
    max_queue_size = Column(Integer, default=5)
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)


class ExecutionHistory(Base):
    __tablename__ = "execution_history"

    id = Column(Integer, primary_key=True, index=True)
    task_id = Column(Integer, index=True)
    status = Column(SAEnum(TaskStatus), nullable=False)
    start_time = Column(DateTime, default=datetime.utcnow)
    end_time = Column(DateTime, nullable=True)
    duration_seconds = Column(Integer, nullable=True)
    response_status = Column(Integer, nullable=True)
    error_message = Column(Text, nullable=True)
    retry_attempt = Column(Integer, default=0)


class TaskCreate(BaseModel):
    name: str = Field(..., description="任务名称")
    task_type: TaskType = Field(..., description="任务类型：cron 或 delayed")
    cron_expression: Optional[str] = Field(None, description="Cron 表达式（cron 类型必填）")
    delay_seconds: Optional[int] = Field(None, description="延迟秒数（delayed 类型必填）")
    callback_url: str = Field(..., description="回调 URL")
    callback_payload: Dict[str, Any] = Field(default_factory=dict, description="回调 JSON 数据")
    timeout_seconds: int = Field(60, ge=1, description="超时时间（秒），默认 60")
    max_retries: int = Field(3, ge=0, description="最大重试次数，默认 3")
    max_queue_size: int = Field(5, ge=1, description="最大排队数，默认 5")


class TaskUpdate(BaseModel):
    name: Optional[str] = None
    cron_expression: Optional[str] = None
    timeout_seconds: Optional[int] = Field(None, ge=1)
    callback_url: Optional[str] = None
    callback_payload: Optional[Dict[str, Any]] = None
    max_retries: Optional[int] = Field(None, ge=0)
    max_queue_size: Optional[int] = Field(None, ge=1)


class TaskResponse(BaseModel):
    id: int
    name: str
    task_type: TaskType
    cron_expression: Optional[str]
    delay_seconds: Optional[int]
    callback_url: str
    callback_payload: Dict[str, Any]
    timeout_seconds: int
    max_retries: int
    retry_count: int
    status: TaskStatus
    next_run_at: Optional[datetime]
    last_run_at: Optional[datetime]
    queue_size: int
    max_queue_size: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class HistoryResponse(BaseModel):
    id: int
    task_id: int
    status: TaskStatus
    start_time: datetime
    end_time: Optional[datetime]
    duration_seconds: Optional[int]
    response_status: Optional[int]
    error_message: Optional[str]
    retry_attempt: int

    class Config:
        from_attributes = True


def init_db():
    Base.metadata.create_all(bind=engine)


def get_db():
    db = SessionLocal()
    try:
        yield db
    finally:
        db.close()
