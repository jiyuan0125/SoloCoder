from sqlalchemy import Column, Integer, String, Text, DateTime, Boolean, Index
from sqlalchemy.sql import func
from app.database import Base


class LogEntry(Base):
    __tablename__ = "logs"

    id = Column(Integer, primary_key=True, autoincrement=True)
    service_name = Column(String(100), nullable=False, index=True)
    level = Column(String(20), nullable=False, index=True)
    message = Column(Text, nullable=False)
    trace_id = Column(String(64), nullable=True, index=True)
    is_exceptional = Column(Boolean, default=False, index=True)
    created_at = Column(DateTime, server_default=func.now(), index=True)

    __table_args__ = (
        Index("idx_service_level", "service_name", "level"),
        Index("idx_trace_created", "trace_id", "created_at"),
    )
