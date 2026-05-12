from datetime import datetime
from typing import Optional, List, Any, Dict
from enum import Enum

from pydantic import BaseModel, Field, field_validator


class LogLevel(str, Enum):
    DEBUG = "DEBUG"
    INFO = "INFO"
    WARNING = "WARNING"
    ERROR = "ERROR"
    CRITICAL = "CRITICAL"


class LogEntryCreate(BaseModel):
    service_name: str = Field(..., min_length=1, max_length=255)
    level: LogLevel
    message: str = Field(..., min_length=1)
    timestamp: Optional[datetime] = None
    trace_id: Optional[str] = Field(None, max_length=64)
    span_id: Optional[str] = Field(None, max_length=64)
    source_host: Optional[str] = Field(None, max_length=255)
    module: Optional[str] = Field(None, max_length=255)
    function: Optional[str] = Field(None, max_length=255)
    line_number: Optional[int] = None
    extra: Optional[Dict[str, Any]] = None

    @field_validator("level", mode="before")
    @classmethod
    def uppercase_level(cls, v):
        if isinstance(v, str):
            return v.upper()
        return v


class LogEntryBatch(BaseModel):
    entries: List[LogEntryCreate]

    @field_validator("entries")
    @classmethod
    def check_batch_size(cls, v):
        if len(v) > 1000:
            raise ValueError("Batch size cannot exceed 1000")
        return v


class LogEntryResponse(BaseModel):
    id: int
    service_name: str
    level: str
    message: str
    timestamp: Optional[str]
    trace_id: Optional[str]
    span_id: Optional[str]
    source_host: Optional[str]
    module: Optional[str]
    function: Optional[str]
    line_number: Optional[int]
    extra: Optional[Dict[str, Any]]
    created_at: Optional[str]


class PaginatedResponse(BaseModel):
    items: List[LogEntryResponse]
    total: int
    page: int
    page_size: int
    has_more: bool


class TraceSearchRequest(BaseModel):
    service_name: Optional[str] = None
    levels: Optional[List[LogLevel]] = None
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    trace_id: Optional[str] = None
    message_contains: Optional[str] = None
    source_host: Optional[str] = None
    module: Optional[str] = None
    page: int = Field(1, ge=1)
    page_size: int = Field(100, ge=1, le=1000)
    sort_by: str = Field("timestamp", pattern="^(timestamp|created_at|id)$")
    sort_order: str = Field("desc", pattern="^(asc|desc)$")


class ServiceStats(BaseModel):
    service_name: str
    total_logs: int
    by_level: Dict[str, int]
    last_log_time: Optional[str]


class LevelStats(BaseModel):
    level: str
    count: int
    percentage: float


class HourlyStats(BaseModel):
    hour: str
    total: int
    by_level: Dict[str, int]


class DashboardStats(BaseModel):
    total_logs: int
    unique_services: int
    by_level: List[LevelStats]
    top_services: List[ServiceStats]
    hourly_trend: List[HourlyStats]
    oldest_log: Optional[str]
    newest_log: Optional[str]


class TraceInfo(BaseModel):
    trace_id: str
    first_seen: Optional[str]
    last_seen: Optional[str]
    service_count: int
    entry_count: int
    root_service: Optional[str]
    summary: Optional[Dict[str, Any]]


class TraceDetail(BaseModel):
    trace: TraceInfo
    logs: List[LogEntryResponse]
