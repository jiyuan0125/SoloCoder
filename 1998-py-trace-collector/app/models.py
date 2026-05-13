from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class Span(BaseModel):
    trace_id: str
    span_id: str
    parent_span_id: Optional[str] = None
    service_name: str
    operation: str
    start_time: float
    end_time: float
    status: str

    @property
    def duration(self) -> float:
        return self.end_time - self.start_time

    def dict(self, **kwargs):
        data = super().model_dump(**kwargs)
        data["duration"] = self.duration
        return data


class SpanTreeNode(BaseModel):
    trace_id: str
    span_id: str
    parent_span_id: Optional[str] = None
    service_name: str
    operation: str
    start_time: float
    end_time: float
    status: str
    duration: float
    duration_ms: float
    total_duration: float
    percentage: float
    children: list["SpanTreeNode"] = []


class ServiceStats(BaseModel):
    service_name: str
    average_duration_ms: float
    error_rate: float
    p99_duration_ms: float
    total_spans: int
    error_spans: int
