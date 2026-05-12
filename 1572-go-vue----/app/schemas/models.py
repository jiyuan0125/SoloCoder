from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional, List
from app.models import TrainStatus, RouteStatus, CrewType, AlertType, AlertStatus


class TrainBase(BaseModel):
    train_number: str
    train_type: str
    seat_capacity: int
    crew_quota: int


class TrainCreate(TrainBase):
    pass


class TrainUpdate(BaseModel):
    train_type: Optional[str] = None
    seat_capacity: Optional[int] = None
    crew_quota: Optional[int] = None
    status: Optional[TrainStatus] = None


class TrainResponse(TrainBase):
    id: int
    status: TrainStatus
    current_route_id: Optional[int] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RouteBase(BaseModel):
    route_code: str
    train_id: int
    departure_station: str
    arrival_station: str
    scheduled_departure: datetime
    scheduled_arrival: datetime


class RouteCreate(RouteBase):
    crew_group_id: Optional[int] = None


class RouteUpdate(BaseModel):
    actual_departure: Optional[datetime] = None
    actual_arrival: Optional[datetime] = None
    status: Optional[RouteStatus] = None
    delay_minutes: Optional[int] = None
    crew_group_id: Optional[int] = None


class RouteResponse(RouteBase):
    id: int
    actual_departure: Optional[datetime] = None
    actual_arrival: Optional[datetime] = None
    status: RouteStatus
    delay_minutes: int
    crew_group_id: Optional[int] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CrewMemberBase(BaseModel):
    employee_id: str
    name: str
    crew_type: CrewType
    phone: Optional[str] = None


class CrewMemberCreate(CrewMemberBase):
    pass


class CrewMemberUpdate(BaseModel):
    name: Optional[str] = None
    crew_type: Optional[CrewType] = None
    phone: Optional[str] = None


class CrewMemberResponse(CrewMemberBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CrewGroupBase(BaseModel):
    group_code: str
    name: str


class CrewGroupCreate(CrewGroupBase):
    member_ids: List[int] = []
    lead_member_id: Optional[int] = None


class CrewGroupUpdate(BaseModel):
    name: Optional[str] = None


class CrewGroupResponse(CrewGroupBase):
    id: int
    created_at: datetime
    updated_at: datetime
    members: List[CrewMemberResponse] = []

    class Config:
        from_attributes = True


class DutyRecordBase(BaseModel):
    crew_member_id: int
    route_id: int
    start_time: datetime


class DutyRecordCreate(DutyRecordBase):
    pass


class DutyRecordUpdate(BaseModel):
    end_time: Optional[datetime] = None
    work_duration_minutes: Optional[int] = None
    rest_period_minutes: Optional[int] = None
    is_completed: Optional[bool] = None


class DutyRecordResponse(DutyRecordBase):
    id: int
    end_time: Optional[datetime] = None
    work_duration_minutes: int
    rest_period_minutes: int
    is_completed: bool
    created_at: datetime

    class Config:
        from_attributes = True


class MonthlyWorkHoursResponse(BaseModel):
    id: int
    crew_member_id: int
    year: int
    month: int
    total_minutes: int
    max_limit_minutes: int
    is_over_limit: bool
    last_recalculated: datetime

    class Config:
        from_attributes = True


class DelayAdjustmentResponse(BaseModel):
    id: int
    route_id: int
    delay_minutes: int
    adjustment_type: str
    adjustment_plan: str
    notified_to: str
    created_at: datetime

    class Config:
        from_attributes = True


class RouteOptimizationResponse(BaseModel):
    id: int
    train_id: int
    optimization_type: str
    description: str
    suggestion: str
    avg_utilization_hours: Optional[float] = None
    avg_turnaround_minutes: Optional[float] = None
    is_implemented: bool
    created_at: datetime

    class Config:
        from_attributes = True


class AlertResponse(BaseModel):
    id: int
    alert_type: AlertType
    title: str
    message: str
    target_role: str
    status: AlertStatus
    related_crew_member_id: Optional[int] = None
    related_route_id: Optional[int] = None
    created_at: datetime
    acknowledged_at: Optional[datetime] = None

    class Config:
        from_attributes = True
