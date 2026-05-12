from datetime import datetime, date
from typing import Optional, List
from pydantic import BaseModel, Field
from .models import (
    ShiftType, RoleType, ContrabandType,
    ChannelType, ChannelStatus, DisposalType
)


class OfficerBase(BaseModel):
    name: str
    badge_number: str
    role: RoleType


class OfficerCreate(OfficerBase):
    pass


class OfficerResponse(OfficerBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class ChannelBase(BaseModel):
    name: str
    channel_type: ChannelType = ChannelType.STANDARD
    capacity_per_hour: int = 120


class ChannelCreate(ChannelBase):
    pass


class ChannelStatusUpdate(BaseModel):
    status: ChannelStatus


class ChannelResponse(ChannelBase):
    id: int
    status: ChannelStatus
    created_at: datetime

    class Config:
        from_attributes = True


class ChannelDetail(ChannelResponse):
    schedules_href: str
    contrabands_href: str
    traffic_href: str


class ScheduleBase(BaseModel):
    officer_id: int
    channel_id: int
    schedule_date: date
    shift: ShiftType


class ScheduleCreate(ScheduleBase):
    pass


class ScheduleResponse(ScheduleBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class CheckInBase(BaseModel):
    officer_id: int


class CheckInCreate(CheckInBase):
    check_in_time: Optional[datetime] = None


class CheckInResponse(BaseModel):
    id: int
    officer_id: int
    schedule_id: Optional[int]
    check_in_time: datetime
    is_temporary: bool
    created_at: datetime

    class Config:
        from_attributes = True


class ContrabandBase(BaseModel):
    item_type: ContrabandType
    description: Optional[str] = None
    disposal_type: DisposalType
    police_badge: Optional[str] = None
    passenger_name: Optional[str] = None


class ContrabandCreate(ContrabandBase):
    channel_id: int
    recorded_at: Optional[datetime] = None


class ContrabandResponse(ContrabandBase):
    id: int
    channel_id: int
    recorded_at: datetime

    class Config:
        from_attributes = True


class TrafficLogBase(BaseModel):
    log_date: date
    hour: int = Field(ge=0, le=23)
    passenger_count: int = Field(ge=0)
    queue_length: int = Field(ge=0)


class TrafficLogCreate(TrafficLogBase):
    channel_id: int


class TrafficLogResponse(TrafficLogBase):
    id: int
    channel_id: int
    created_at: datetime

    class Config:
        from_attributes = True


class TrafficSummary(BaseModel):
    channel_id: int
    total_passengers: int
    max_queue_length: int
    avg_queue_length: float
    date_range: str


class ChannelStatusInfo(BaseModel):
    channel_id: int
    channel_name: str
    status: ChannelStatus
    scanner_present: bool
    handcheck_count: int
    verifier_present: bool
    is_valid: bool
    missing_roles: List[str]


class ChannelCapacityAlert(BaseModel):
    channel_id: int
    channel_name: str
    current_queue: int
    capacity_per_hour: int
    max_allowed: int
    should_add_channel: bool
    reason: str
