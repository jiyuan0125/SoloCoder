from sqlalchemy import Column, Integer, String, DateTime, Text, ForeignKey, Float, Enum
from sqlalchemy.orm import relationship
from sqlalchemy.sql import func
from .database import Base
import enum


class AlertStatus(str, enum.Enum):
    PENDING = "pending"
    VERIFIED = "verified"
    DISMISSED = "dismissed"
    MERGED = "merged"


class AlertUrgency(str, enum.Enum):
    NORMAL = "normal"
    EMERGENCY = "emergency"


class EventStatus(str, enum.Enum):
    RECEIVED = "received"
    VERIFIED = "verified"
    DISPATCHED = "dispatched"
    RESCUING = "rescuing"
    COMPLETED = "completed"


class ResponseLevel(str, enum.Enum):
    NORMAL = "normal"
    YELLOW = "yellow"
    RED = "red"


class TimelineEventType(str, enum.Enum):
    ALERT_RECEIVED = "alert_received"
    ALERT_VERIFIED = "alert_verified"
    EVENT_CREATED = "event_created"
    EVENT_VERIFIED = "event_verified"
    FORCE_DISPATCHED = "force_dispatched"
    FORCE_ARRIVED = "force_arrived"
    RESCUE_STARTED = "rescue_started"
    RESCUE_ENDED = "rescue_ended"
    EVENT_COMPLETED = "event_completed"


class ForceStatus(str, enum.Enum):
    AVAILABLE = "available"
    DISPATCHED = "dispatched"
    WORKING = "working"


class Alert(Base):
    __tablename__ = "alerts"

    id = Column(Integer, primary_key=True, index=True)
    location = Column(String, nullable=False)
    latitude = Column(Float)
    longitude = Column(Float)
    description = Column(Text)
    source = Column(String, nullable=False)
    status = Column(Enum(AlertStatus), default=AlertStatus.PENDING)
    urgency = Column(Enum(AlertUrgency), default=AlertUrgency.NORMAL)
    created_at = Column(DateTime, server_default=func.now())
    verified_at = Column(DateTime)
    verified_by = Column(String)
    verification_notes = Column(Text)
    event_id = Column(Integer, ForeignKey("events.id"), nullable=True)

    event = relationship("Event", back_populates="alerts")
    timeline_events = relationship("TimelineEvent", back_populates="alert")


class Event(Base):
    __tablename__ = "events"

    id = Column(Integer, primary_key=True, index=True)
    title = Column(String, nullable=False)
    location = Column(String, nullable=False)
    latitude = Column(Float)
    longitude = Column(Float)
    description = Column(Text)
    status = Column(Enum(EventStatus), default=EventStatus.RECEIVED)
    urgency = Column(Enum(AlertUrgency), default=AlertUrgency.NORMAL)
    created_at = Column(DateTime, server_default=func.now())
    verified_at = Column(DateTime)
    response_level = Column(Enum(ResponseLevel), default=ResponseLevel.NORMAL)

    alerts = relationship("Alert", back_populates="event")
    force_assignments = relationship("ForceAssignment", back_populates="event")
    timeline_events = relationship("TimelineEvent", back_populates="event")
    result = relationship("EventResult", uselist=False, back_populates="event")


class SearchRescueForce(Base):
    __tablename__ = "rescue_forces"

    id = Column(Integer, primary_key=True, index=True)
    name = Column(String, nullable=False)
    type = Column(String)
    capacity = Column(Integer)
    current_location = Column(String)
    latitude = Column(Float)
    longitude = Column(Float)
    estimated_arrival_minutes = Column(Integer, default=0)
    status = Column(Enum(ForceStatus), default=ForceStatus.AVAILABLE)

    assignments = relationship("ForceAssignment", back_populates="force")


class ForceAssignment(Base):
    __tablename__ = "force_assignments"

    id = Column(Integer, primary_key=True, index=True)
    event_id = Column(Integer, ForeignKey("events.id"), nullable=False)
    force_id = Column(Integer, ForeignKey("rescue_forces.id"), nullable=False)
    assigned_at = Column(DateTime, server_default=func.now())
    arrived_at = Column(DateTime)
    completed_at = Column(DateTime)
    assignment_notes = Column(Text)

    event = relationship("Event", back_populates="force_assignments")
    force = relationship("SearchRescueForce", back_populates="assignments")


class TimelineEvent(Base):
    __tablename__ = "timeline_events"

    id = Column(Integer, primary_key=True, index=True)
    event_id = Column(Integer, ForeignKey("events.id"), nullable=True)
    alert_id = Column(Integer, ForeignKey("alerts.id"), nullable=True)
    event_type = Column(Enum(TimelineEventType), nullable=False)
    description = Column(Text)
    timestamp = Column(DateTime, server_default=func.now())
    actor = Column(String)

    event = relationship("Event", back_populates="timeline_events")
    alert = relationship("Alert", back_populates="timeline_events")


class EventResult(Base):
    __tablename__ = "event_results"

    id = Column(Integer, primary_key=True, index=True)
    event_id = Column(Integer, ForeignKey("events.id"), nullable=False, unique=True)
    successful = Column(Integer, default=0)
    injured = Column(Integer, default=0)
    fatalities = Column(Integer, default=0)
    total_people = Column(Integer, default=0)
    notes = Column(Text)
    response_time_minutes = Column(Integer)
    completed_at = Column(DateTime, server_default=func.now())

    event = relationship("Event", back_populates="result")
