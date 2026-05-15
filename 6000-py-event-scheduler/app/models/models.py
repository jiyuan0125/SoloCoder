from sqlalchemy import Column, Integer, String, DateTime, Boolean, Text, Enum as SQLEnum
from sqlalchemy.sql import func
from app.core.database import Base
import enum


class TaskStatus(str, enum.Enum):
    PENDING = "pending"
    RUNNING = "running"
    SUCCESS = "success"
    FAILED = "failed"
    TIMEOUT = "timeout"


class AuditAction(str, enum.Enum):
    CREATE = "create"
    UPDATE = "update"
    DELETE = "delete"
    EXECUTE = "execute"


class TaskConfig(Base):
    __tablename__ = "task_configs"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(255), unique=True, index=True, nullable=False)
    cron_expression = Column(String(100), nullable=False)
    callback_url = Column(String(500), nullable=False)
    timeout = Column(Integer, default=30)
    max_retry = Column(Integer, default=0)
    created_by = Column(String(100), nullable=False)
    operation_source = Column(String(100), nullable=False)
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime(timezone=True), server_default=func.now())
    updated_at = Column(DateTime(timezone=True), onupdate=func.now())


class ExecutionAudit(Base):
    __tablename__ = "execution_audits"

    id = Column(Integer, primary_key=True, index=True)
    task_id = Column(Integer, index=True)
    task_name = Column(String(255), index=True)
    trigger_time = Column(DateTime(timezone=True), nullable=False)
    start_time = Column(DateTime(timezone=True))
    end_time = Column(DateTime(timezone=True))
    status = Column(SQLEnum(TaskStatus), default=TaskStatus.PENDING)
    callback_url = Column(String(500))
    callback_result = Column(Text)
    executor = Column(String(100), default="system")
    retry_count = Column(Integer, default=0)
    error_message = Column(Text)


class OperationAudit(Base):
    __tablename__ = "operation_audits"

    id = Column(Integer, primary_key=True, index=True)
    task_id = Column(Integer, index=True)
    task_name = Column(String(255), index=True)
    action = Column(SQLEnum(AuditAction), nullable=False)
    operator = Column(String(100), nullable=False)
    operation_source = Column(String(100), nullable=False)
    operation_time = Column(DateTime(timezone=True), server_default=func.now())
    details = Column(Text)
