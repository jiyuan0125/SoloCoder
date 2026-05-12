from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional, List, Literal
from enum import Enum


class LogLevel(str, Enum):
    DEBUG = "DEBUG"
    INFO = "INFO"
    WARN = "WARN"
    ERROR = "ERROR"


class LogCreate(BaseModel):
    service_name: str = Field(..., max_length=100)
    level: LogLevel
    message: str
    trace_id: Optional[str] = Field(None, max_length=64)


class LogResponse(BaseModel):
    id: int
    service_name: str
    level: str
    message: str
    trace_id: Optional[str]
    is_exceptional: bool
    created_at: datetime

    class Config:
        from_attributes = True


class TraceAggregate(BaseModel):
    trace_id: str
    logs: List[LogResponse]
    is_exceptional: bool


class LogQueryParams(BaseModel):
    service_name: Optional[str] = None
    level: Optional[LogLevel] = None
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    keyword: Optional[str] = None
    trace_id: Optional[str] = None
    is_exceptional: Optional[bool] = None
    limit: int = Field(100, ge=1, le=1000)
    offset: int = Field(0, ge=0)


class TimeWindow(str, Enum):
    ONE_HOUR = "1h"
    TWENTY_FOUR_HOURS = "24h"
    SEVEN_DAYS = "7d"


class ServiceStats(BaseModel):
    service_name: str
    debug_count: int
    info_count: int
    warn_count: int
    error_count: int
    total_count: int


class StatsResponse(BaseModel):
    time_window: str
    stats: List[ServiceStats]
