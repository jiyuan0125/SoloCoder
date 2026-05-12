from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from .models import (
    AlertStatus, AlertUrgency, EventStatus, ResponseLevel,
    TimelineEventType, ForceStatus
)


class AlertBase(BaseModel):
    location: str
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    description: Optional[str] = None
    source: str


class AlertCreate(AlertBase):
    pass


class AlertVerify(BaseModel):
    verified_by: str
    verification_notes: Optional[str] = None
    verify_as_valid: bool = True


class AlertUpdate(BaseModel):
    location: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    description: Optional[str] = None


class Alert(BaseModel):
    id: int
    location: str
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    description: Optional[str] = None
    source: str
    status: AlertStatus
    urgency: AlertUrgency
    created_at: datetime
    verified_at: Optional[datetime] = None
    verified_by: Optional[str] = None
    verification_notes: Optional[str] = None
    event_id: Optional[int] = None

    class Config:
        from_attributes = True


class EventBase(BaseModel):
    title: str
    location: str
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    description: Optional[str] = None


class EventCreate(EventBase):
    alert_ids: Optional[List[int]] = None


class Event(BaseModel):
    id: int
    title: str
    location: str
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    description: Optional[str] = None
    status: EventStatus
    urgency: AlertUrgency
    created_at: datetime
    verified_at: Optional[datetime] = None
    response_level: ResponseLevel

    class Config:
        from_attributes = True


class EventDetail(Event):
    alerts: List[Alert] = []


class EventVerify(BaseModel):
    verified_by: str
    notes: Optional[str] = None


class RescueForceBase(BaseModel):
    name: str
    type: Optional[str] = None
    capacity: Optional[int] = None
    current_location: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    estimated_arrival_minutes: int = 0


class RescueForceCreate(RescueForceBase):
    pass


class RescueForce(RescueForceBase):
    id: int
    status: ForceStatus

    class Config:
        from_attributes = True


class ForceAssignmentBase(BaseModel):
    force_id: int
    assignment_notes: Optional[str] = None


class ForceAssignmentCreate(ForceAssignmentBase):
    pass


class ForceAssignment(BaseModel):
    id: int
    event_id: int
    force_id: int
    assigned_at: datetime
    arrived_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    assignment_notes: Optional[str] = None
    force: RescueForce

    class Config:
        from_attributes = True


class TimelineEventBase(BaseModel):
    event_type: TimelineEventType
    description: Optional[str] = None
    actor: Optional[str] = None


class TimelineEvent(BaseModel):
    id: int
    event_id: Optional[int] = None
    alert_id: Optional[int] = None
    event_type: TimelineEventType
    description: Optional[str] = None
    timestamp: datetime
    actor: Optional[str] = None

    class Config:
        from_attributes = True


class EventResultBase(BaseModel):
    successful: int = 0
    injured: int = 0
    fatalities: int = 0
    total_people: int = 0
    notes: Optional[str] = None


class EventResultCreate(EventResultBase):
    pass


class EventResult(BaseModel):
    id: int
    event_id: int
    successful: int
    injured: int
    fatalities: int
    total_people: int
    notes: Optional[str] = None
    response_time_minutes: Optional[int] = None
    completed_at: datetime

    class Config:
        from_attributes = True


class MonthlyStats(BaseModel):
    year: int
    month: int
    alert_count: int
    event_count: int
    successful_rescues: int


class AlertRecommendation(BaseModel):
    id: int
    name: str
    type: Optional[str] = None
    capacity: Optional[int] = None
    current_location: Optional[str] = None
    estimated_arrival_minutes: int
    status: ForceStatus

    class Config:
        from_attributes = True
