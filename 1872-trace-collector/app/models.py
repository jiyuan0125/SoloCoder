from sqlalchemy import Column, String, DateTime, Text, Float, Integer, Index
from sqlalchemy.dialects.sqlite import JSON
from .database import Base


class SpanModel(Base):
    __tablename__ = "spans"

    id = Column(Integer, primary_key=True, autoincrement=True)
    trace_id = Column(String(64), nullable=False, index=True)
    span_id = Column(String(64), nullable=False, index=True)
    operation = Column(String(255), nullable=False)
    start_time = Column(DateTime, nullable=False, index=True)
    end_time = Column(DateTime, nullable=False)
    duration_ms = Column(Float, nullable=False, index=True)
    tags = Column(JSON, nullable=False, default=dict)
    parent_span_id = Column(String(64), nullable=True, index=True)
    created_at = Column(DateTime, nullable=False, index=True)

    __table_args__ = (
        Index("ix_spans_trace_span", "trace_id", "span_id", unique=True),
    )


class RedactionRuleModel(Base):
    __tablename__ = "redaction_rules"

    id = Column(Integer, primary_key=True, autoincrement=True)
    tag_key = Column(String(255), nullable=False, unique=True, index=True)
    redaction_type = Column(String(20), nullable=False)
    replacement = Column(Text, nullable=True)
