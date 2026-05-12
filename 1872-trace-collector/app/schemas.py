from typing import Optional, Dict, List, Union, Any
from datetime import datetime
from pydantic import BaseModel, Field
from enum import Enum


class RedactionType(str, Enum):
    mask = "mask"
    replace = "replace"


class SpanCreate(BaseModel):
    trace_id: str
    span_id: str
    operation: str
    start_time: datetime
    end_time: datetime
    tags: Dict[str, Any] = Field(default_factory=dict)
    parent_span_id: Optional[str] = None

    class Config:
        from_attributes = True


class Span(BaseModel):
    trace_id: str
    span_id: str
    operation: str
    start_time: datetime
    end_time: datetime
    duration_ms: float
    tags: Dict[str, Any] = Field(default_factory=dict)
    parent_span_id: Optional[str] = None

    class Config:
        from_attributes = True


class SpanTreeNode(BaseModel):
    span: Span
    children: List["SpanTreeNode"] = Field(default_factory=list)
    is_orphan: bool = False


SpanTreeNode.model_rebuild()


class TraceTree(BaseModel):
    trace_id: str
    root_spans: List[SpanTreeNode]
    orphan_spans: List[Span] = Field(default_factory=list)


class SlowSpan(BaseModel):
    trace_id: str
    span_id: str
    operation: str
    service: Optional[str]
    start_time: datetime
    end_time: datetime
    duration_ms: float


class RedactionRule(BaseModel):
    tag_key: str
    redaction_type: RedactionType
    replacement: Optional[str] = None
