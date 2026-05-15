from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional
from app.models.models import TaskStatus, AuditAction


class TaskConfigBase(BaseModel):
    name: str = Field(..., max_length=255)
    cron_expression: str = Field(..., max_length=100)
    callback_url: str = Field(..., max_length=500)
    timeout: Optional[int] = Field(default=30, ge=1)
    max_retry: Optional[int] = Field(default=0, ge=0)
    is_active: Optional[bool] = True


class TaskConfigCreate(TaskConfigBase):
    pass


class TaskConfigUpdate(BaseModel):
    name: Optional[str] = Field(None, max_length=255)
    cron_expression: Optional[str] = Field(None, max_length=100)
    callback_url: Optional[str] = Field(None, max_length=500)
    timeout: Optional[int] = Field(None, ge=1)
    max_retry: Optional[int] = Field(None, ge=0)
    is_active: Optional[bool] = None


class TaskConfigResponse(TaskConfigBase):
    id: int
    created_by: str
    operation_source: str
    created_at: datetime
    updated_at: Optional[datetime]

    class Config:
        from_attributes = True


class ExecutionAuditResponse(BaseModel):
    id: int
    task_id: int
    task_name: str
    trigger_time: datetime
    start_time: Optional[datetime]
    end_time: Optional[datetime]
    status: TaskStatus
    callback_url: Optional[str]
    callback_result: Optional[str]
    executor: str
    retry_count: int
    error_message: Optional[str]

    class Config:
        from_attributes = True


class OperationAuditResponse(BaseModel):
    id: int
    task_id: Optional[int]
    task_name: Optional[str]
    action: AuditAction
    operator: str
    operation_source: str
    operation_time: datetime
    details: Optional[str]

    class Config:
        from_attributes = True


class AuditQueryParams(BaseModel):
    task_name: Optional[str] = None
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    status: Optional[TaskStatus] = None
    created_by: Optional[str] = None
    skip: int = 0
    limit: int = 100
