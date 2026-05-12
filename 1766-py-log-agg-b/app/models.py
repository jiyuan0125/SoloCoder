import json
from datetime import datetime
from typing import Optional, Any

from sqlalchemy import (
    Column,
    Integer,
    String,
    Text,
    DateTime,
    Index,
    JSON,
    ForeignKey,
)
from sqlalchemy.orm import relationship

from app.database import Base


class LogEntry(Base):
    __tablename__ = "log_entries"

    id = Column(Integer, primary_key=True, autoincrement=True)
    service_name = Column(String(255), nullable=False, index=True)
    level = Column(String(20), nullable=False, index=True)
    message = Column(Text, nullable=False)
    timestamp = Column(DateTime, nullable=False, index=True)
    trace_id = Column(String(64), nullable=True, index=True)
    span_id = Column(String(64), nullable=True)
    source_host = Column(String(255), nullable=True)
    module = Column(String(255), nullable=True)
    function = Column(String(255), nullable=True)
    line_number = Column(Integer, nullable=True)
    extra = Column(JSON, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, index=True)

    __table_args__ = (
        Index(
            "ix_log_service_level_time",
            "service_name",
            "level",
            "timestamp",
        ),
        Index(
            "ix_log_trace_time",
            "trace_id",
            "timestamp",
        ),
    )

    def to_dict(self) -> dict:
        return {
            "id": self.id,
            "service_name": self.service_name,
            "level": self.level,
            "message": self.message,
            "timestamp": self.timestamp.isoformat() if self.timestamp else None,
            "trace_id": self.trace_id,
            "span_id": self.span_id,
            "source_host": self.source_host,
            "module": self.module,
            "function": self.function,
            "line_number": self.line_number,
            "extra": self.extra,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }


class TraceMetadata(Base):
    __tablename__ = "trace_metadata"

    id = Column(Integer, primary_key=True, autoincrement=True)
    trace_id = Column(String(64), unique=True, nullable=False, index=True)
    first_seen = Column(DateTime, nullable=False)
    last_seen = Column(DateTime, nullable=False, index=True)
    service_count = Column(Integer, default=0)
    entry_count = Column(Integer, default=0)
    root_service = Column(String(255), nullable=True)
    summary = Column(JSON, nullable=True)

    def to_dict(self) -> dict:
        return {
            "trace_id": self.trace_id,
            "first_seen": self.first_seen.isoformat() if self.first_seen else None,
            "last_seen": self.last_seen.isoformat() if self.last_seen else None,
            "service_count": self.service_count,
            "entry_count": self.entry_count,
            "root_service": self.root_service,
            "summary": self.summary,
        }
