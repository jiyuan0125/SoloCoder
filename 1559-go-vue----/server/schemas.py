from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field
from .models import (
    AlertStatus, AlertPriority, EventStatus, EventLevel, ForceStatus
)


class AlertCreate(BaseModel):
    location: str
    location_lat: Optional[float] = None
    location_lon: Optional[float] = None
    description: Optional[str] = None
    reporter_name: Optional[str] = None
    reporter_contact: Optional[str] = None
    alert_time: Optional[datetime] = None


class AlertVerify(BaseModel):
    verified_by: str
    verification_notes: Optional[str] = None
    accept: bool = True


class AlertResponse(BaseModel):
    id: int
    alert_time: datetime
    location: str
    location_lat: Optional[float] = None
    location_lon: Optional[float] = None
    description: Optional[str] = None
    reporter_name: Optional[str] = None
    reporter_contact: Optional[str] = None
    status: AlertStatus
    priority: AlertPriority
    verified_at: Optional[datetime] = None
    verified_by: Optional[str] = None
    verification_notes: Optional[str] = None
    event_id: Optional[int] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ForceCreate(BaseModel):
    name: str
    type: Optional[str] = None
    location: str
    location_lat: Optional[float] = None
    location_lon: Optional[float] = None
    estimated_arrival_minutes: int = 30
    capacity: Optional[int] = None
    contact_person: Optional[str] = None
    contact_phone: Optional[str] = None


class ForceResponse(BaseModel):
    id: int
    name: str
    type: Optional[str] = None
    location: str
    location_lat: Optional[float] = None
    location_lon: Optional[float] = None
    estimated_arrival_minutes: int
    capacity: Optional[int] = None
    contact_person: Optional[str] = None
    contact_phone: Optional[str] = None
    status: ForceStatus
    current_event_id: Optional[int] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class EventForceAssignmentResponse(BaseModel):
    id: int
    event_id: int
    force_id: int
    assigned_at: datetime
    arrived_at: Optional[datetime] = None
    departed_at: Optional[datetime] = None
    force: Optional[ForceResponse] = None

    class Config:
        from_attributes = True


class TimelineEntryResponse(BaseModel):
    id: int
    event_id: int
    entry_type: str
    entry_time: datetime
    description: str
    operator: Optional[str] = None

    class Config:
        from_attributes = True


class RescueResultCreate(BaseModel):
    people_rescued: int = 0
    people_injured: int = 0
    people_deceased: int = 0
    is_successful: bool = True
    summary: Optional[str] = None
    notes: Optional[str] = None


class RescueResultResponse(BaseModel):
    id: int
    event_id: int
    people_rescued: int
    people_injured: int
    people_deceased: int
    is_successful: bool
    summary: Optional[str] = None
    notes: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class EventResponse(BaseModel):
    id: int
    event_code: str
    status: EventStatus
    level: EventLevel
    location: str
    location_lat: Optional[float] = None
    location_lon: Optional[float] = None
    description: Optional[str] = None
    people_involved: int
    received_at: datetime
    verified_at: Optional[datetime] = None
    dispatched_at: Optional[datetime] = None
    rescuing_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    response_time_minutes: Optional[int] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class EventDetailResponse(EventResponse):
    alerts: List[AlertResponse] = []
    force_assignments: List[EventForceAssignmentResponse] = []
    timeline: List[TimelineEntryResponse] = []
    result: Optional[RescueResultResponse] = None


class DispatchForceRequest(BaseModel):
    force_id: int
    operator: Optional[str] = None


class ForceArrivalRequest(BaseModel):
    operator: Optional[str] = None


class CompleteRescueRequest(BaseModel):
    result: RescueResultCreate
    operator: Optional[str] = None


class MonthlyStatsResponse(BaseModel):
    year: int
    month: int
    total_alerts: int
    total_events: int
    total_people_rescued: int


class EventLevelUpdate(BaseModel):
    level: EventLevel
