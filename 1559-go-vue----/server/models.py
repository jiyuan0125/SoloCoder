from datetime import datetime
from sqlalchemy import (
    Column, Integer, String, DateTime, Text, Float, Boolean, ForeignKey, Enum
)
from sqlalchemy.orm import relationship
from sqlalchemy.ext.declarative import declarative_base
import enum

Base = declarative_base()


class AlertStatus(enum.Enum):
    RECEIVED = "received"
    VERIFIED = "verified"
    REJECTED = "rejected"


class AlertPriority(enum.Enum):
    NORMAL = "normal"
    URGENT = "urgent"


class EventStatus(enum.Enum):
    RECEIVED = "received"
    VERIFIED = "verified"
    DISPATCHED = "dispatched"
    RESCUING = "rescuing"
    COMPLETED = "completed"


class EventLevel(enum.Enum):
    NORMAL = "normal"
    YELLOW = "yellow"
    RED = "red"


class ForceStatus(enum.Enum):
    AVAILABLE = "available"
    DISPATCHED = "dispatched"
    RESCUING = "rescuing"


class Alert(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, index=True)
    alert_time = Column(DateTime, default=datetime.utcnow, nullable=False)
    location = Column(String(500), nullable=False, index=True)
    location_lat = Column(Float, nullable=True)
    location_lon = Column(Float, nullable=True)
    description = Column(Text, nullable=True)
    reporter_name = Column(String(200), nullable=True)
    reporter_contact = Column(String(100), nullable=True)
    status = Column(Enum(AlertStatus), default=AlertStatus.RECEIVED, nullable=False)
    priority = Column(Enum(AlertPriority), default=AlertPriority.NORMAL, nullable=False)
    verified_at = Column(DateTime, nullable=True)
    verified_by = Column(String(200), nullable=True)
    verification_notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

    event_id = Column(Integer, ForeignKey("events.id"), nullable=True)
    event = relationship("Event", back_populates="alerts")


class RescueForce(Base):
    __tablename__ = "rescue_forces"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String(200), nullable=False)
    type = Column(String(100), nullable=True)
    location = Column(String(500), nullable=False)
    location_lat = Column(Float, nullable=True)
    location_lon = Column(Float, nullable=True)
    estimated_arrival_minutes = Column(Integer, nullable=False, default=30)
    capacity = Column(Integer, nullable=True)
    contact_person = Column(String(200), nullable=True)
    contact_phone = Column(String(100), nullable=True)
    status = Column(Enum(ForceStatus), default=ForceStatus.AVAILABLE, nullable=False)
    current_event_id = Column(Integer, ForeignKey("events.id"), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

    event_assignments = relationship("EventForceAssignment", back_populates="force")


class Event(Base):
    __tablename__ = "events"

    id = Column(Integer, primary_key=True, index=True)
    event_code = Column(String(50), unique=True, index=True, nullable=False)
    status = Column(Enum(EventStatus), default=EventStatus.RECEIVED, nullable=False)
    level = Column(Enum(EventLevel), default=EventLevel.NORMAL, nullable=False)
    location = Column(String(500), nullable=False)
    location_lat = Column(Float, nullable=True)
    location_lon = Column(Float, nullable=True)
    description = Column(Text, nullable=True)
    people_involved = Column(Integer, default=0)
    received_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    verified_at = Column(DateTime, nullable=True)
    dispatched_at = Column(DateTime, nullable=True)
    rescuing_at = Column(DateTime, nullable=True)
    completed_at = Column(DateTime, nullable=True)
    response_time_minutes = Column(Integer, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

    alerts = relationship("Alert", back_populates="event")
    force_assignments = relationship("EventForceAssignment", back_populates="event")
    timeline = relationship("TimelineEntry", back_populates="event")
    result = relationship("RescueResult", back_populates="event", uselist=False)


class EventForceAssignment(Base):
    __tablename__ = "event_force_assignments"

    id = Column(Integer, primary_key=True, index=True)
    event_id = Column(Integer, ForeignKey("events.id"), nullable=False)
    force_id = Column(Integer, ForeignKey("rescue_forces.id"), nullable=False)
    assigned_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    arrived_at = Column(DateTime, nullable=True)
    departed_at = Column(DateTime, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)

    event = relationship("Event", back_populates="force_assignments")
    force = relationship("RescueForce", back_populates="event_assignments")


class TimelineEntry(Base):
    __tablename__ = "timeline_entries"

    id = Column(Integer, primary_key=True, index=True)
    event_id = Column(Integer, ForeignKey("events.id"), nullable=False)
    entry_type = Column(String(100), nullable=False)
    entry_time = Column(DateTime, default=datetime.utcnow, nullable=False)
    description = Column(Text, nullable=False)
    operator = Column(String(200), nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)

    event = relationship("Event", back_populates="timeline")


class RescueResult(Base):
    __tablename__ = "rescue_results"

    id = Column(Integer, primary_key=True, index=True)
    event_id = Column(Integer, ForeignKey("events.id"), nullable=False, unique=True)
    people_rescued = Column(Integer, default=0)
    people_injured = Column(Integer, default=0)
    people_deceased = Column(Integer, default=0)
    is_successful = Column(Boolean, default=True)
    summary = Column(Text, nullable=True)
    notes = Column(Text, nullable=True)
    created_at = Column(DateTime, default=datetime.utcnow, nullable=False)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow, nullable=False)

    event = relationship("Event", back_populates="result")
