from typing import List, Optional
from datetime import datetime
from pydantic import BaseModel, Field, field_validator
from enum import Enum


class LogLevel(str, Enum):
    DEBUG = "DEBUG"
    INFO = "INFO"
    WARN = "WARN"
    ERROR = "ERROR"


class LogEntry(BaseModel):
    service: str
    level: LogLevel
    message: str
    timestamp: datetime

    class Config:
        json_encoders = {
            datetime: lambda v: v.isoformat()
        }


class BatchLogRequest(BaseModel):
    logs: List[dict]

    @field_validator('logs')
    @classmethod
    def check_max_logs(cls, v):
        if len(v) > 1000:
            raise ValueError(f"单次批量上报最多 1000 条，当前 {len(v)} 条")
        return v


class BatchLogResponse(BaseModel):
    received: int
    rejected: int


class ServiceStats(BaseModel):
    service: str
    count: int
    earliest: Optional[datetime]
    latest: Optional[datetime]
    estimated_memory_bytes: int


class StorageStats(BaseModel):
    total_logs: int
    total_estimated_memory_bytes: int
    services: List[ServiceStats]


class LogQueryResponse(BaseModel):
    logs: List[LogEntry]
    total: int
    page: int
    page_size: int
    total_pages: int
