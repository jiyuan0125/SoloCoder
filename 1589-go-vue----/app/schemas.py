from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field
from app.models import RopewayStatus, InspectionType, InspectionStatus


class WeatherDataBase(BaseModel):
    wind_speed: float = Field(..., ge=0, description="风速，单位：米/秒")
    has_lightning: bool = False
    temperature: Optional[float] = None
    humidity: Optional[float] = None


class WeatherDataCreate(WeatherDataBase):
    pass


class WeatherDataResponse(WeatherDataBase):
    id: int
    timestamp: datetime

    class Config:
        from_attributes = True


class StatusRecordBase(BaseModel):
    to_status: RopewayStatus
    reason: str


class StatusRecordResponse(StatusRecordBase):
    id: int
    timestamp: datetime
    from_status: Optional[RopewayStatus]
    operator_id: Optional[int]

    class Config:
        from_attributes = True


class OperatorBase(BaseModel):
    username: str
    full_name: str


class OperatorCreate(OperatorBase):
    pass


class OperatorResponse(OperatorBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class InspectionBase(BaseModel):
    inspection_type: InspectionType
    scheduled_date: datetime
    notes: Optional[str] = None


class InspectionCreate(InspectionBase):
    pass


class InspectionUpdate(BaseModel):
    completed_date: Optional[datetime] = None
    status: Optional[InspectionStatus] = None
    is_urgent: Optional[bool] = None
    resolved: Optional[bool] = None
    issues_found: Optional[str] = None
    notes: Optional[str] = None
    operator_id: Optional[int] = None


class InspectionResponse(InspectionBase):
    id: int
    completed_date: Optional[datetime]
    status: InspectionStatus
    is_urgent: bool
    resolved: bool
    issues_found: Optional[str]
    operator_id: Optional[int]

    class Config:
        from_attributes = True


class GondolaCapacityBase(BaseModel):
    total_capacity: int = Field(..., gt=0)


class GondolaCapacityCreate(GondolaCapacityBase):
    pass


class GondolaCapacityResponse(GondolaCapacityBase):
    id: int
    effective_date: datetime
    is_active: bool

    class Config:
        from_attributes = True


class PassengerRecordBase(BaseModel):
    gondola_id: int
    adult_count: int = Field(default=0, ge=0)
    child_count: int = Field(default=0, ge=0)


class PassengerRecordCreate(PassengerRecordBase):
    pass


class PassengerRecordResponse(PassengerRecordBase):
    id: int
    timestamp: datetime
    total_counted: int

    class Config:
        from_attributes = True


class QueueDataBase(BaseModel):
    queue_count: int = Field(..., ge=0)


class QueueDataCreate(QueueDataBase):
    pass


class QueueDataResponse(QueueDataBase):
    id: int
    timestamp: datetime
    gondola_capacity: int
    is_limited: bool

    class Config:
        from_attributes = True


class DailyReportBase(BaseModel):
    report_date: datetime


class DailyReportResponse(DailyReportBase):
    id: int
    total_operating_hours: float
    effective_operating_hours: float
    total_passengers: int
    total_passengers_counted: int
    below_min_hours: bool
    weather_events: Optional[str]
    urgent_inspections: Optional[str]
    created_at: datetime

    class Config:
        from_attributes = True


class CurrentStatusResponse(BaseModel):
    current_status: RopewayStatus
    last_update: Optional[datetime]
    current_wind_speed: Optional[float]
    has_lightning: bool
    pending_confirmation: bool
    lag_counter: int
    lag_threshold: int


class RecoveryConfirm(BaseModel):
    operator_id: int
    notes: Optional[str] = None


class StatusTransitionResponse(BaseModel):
    success: bool
    new_status: RopewayStatus
    reason: str
