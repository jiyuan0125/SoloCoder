from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from .models import (
    TrainStatus,
    SignalStatus,
    PowerStatus,
    LogLevel,
)


class LineBase(BaseModel):
    name: str
    description: Optional[str] = None


class LineCreate(LineBase):
    pass


class LineUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None


class LineResponse(LineBase):
    id: int
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class StationBase(BaseModel):
    line_id: int
    name: str
    sequence: int
    is_terminal: bool = False


class StationCreate(StationBase):
    pass


class StationUpdate(BaseModel):
    name: Optional[str] = None
    sequence: Optional[int] = None
    is_terminal: Optional[bool] = None


class StationResponse(StationBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class TrainBase(BaseModel):
    train_number: str
    capacity: int = 1000


class TrainCreate(TrainBase):
    pass


class TrainUpdate(BaseModel):
    status: Optional[TrainStatus] = None
    current_station_id: Optional[int] = None
    current_section_id: Optional[int] = None
    speed_limit: Optional[float] = None
    capacity: Optional[int] = None


class TrainResponse(TrainBase):
    id: int
    status: TrainStatus
    current_station_id: Optional[int] = None
    current_section_id: Optional[int] = None
    speed_limit: float
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class SignalBase(BaseModel):
    signal_code: str
    station_id: int
    section_id: Optional[int] = None


class SignalCreate(SignalBase):
    pass


class SignalUpdate(BaseModel):
    status: Optional[SignalStatus] = None
    fault_description: Optional[str] = None


class SignalResponse(SignalBase):
    id: int
    status: SignalStatus
    fault_time: Optional[datetime] = None
    recovery_time: Optional[datetime] = None
    fault_description: Optional[str] = None
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class PowerSectionBase(BaseModel):
    line_id: int
    section_code: str
    name: str
    start_station_id: Optional[int] = None
    end_station_id: Optional[int] = None


class PowerSectionCreate(PowerSectionBase):
    pass


class PowerSectionUpdate(BaseModel):
    name: Optional[str] = None
    status: Optional[PowerStatus] = None


class PowerSectionResponse(PowerSectionBase):
    id: int
    status: PowerStatus
    outage_time: Optional[datetime] = None
    recovery_time: Optional[datetime] = None
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class FaultReportBase(BaseModel):
    signal_id: int
    fault_description: Optional[str] = None


class FaultReportCreate(FaultReportBase):
    pass


class FaultReportResponse(FaultReportBase):
    id: int
    report_time: datetime
    resolved_by: Optional[str] = None
    resolved_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class ScheduleBase(BaseModel):
    line_id: int
    name: str
    first_departure: datetime
    last_departure: datetime
    interval_seconds: int = Field(gt=0, default=300)


class ScheduleCreate(ScheduleBase):
    auto_generate_trains: bool = True


class ScheduleUpdate(BaseModel):
    name: Optional[str] = None
    first_departure: Optional[datetime] = None
    last_departure: Optional[datetime] = None
    interval_seconds: Optional[int] = Field(default=None, gt=0)
    is_active: Optional[bool] = None


class ScheduleResponse(ScheduleBase):
    id: int
    is_active: bool
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class ScheduleTrainBase(BaseModel):
    schedule_id: int
    train_id: Optional[int] = None
    sequence: int
    scheduled_departure: datetime
    scheduled_arrival: Optional[datetime] = None


class ScheduleTrainCreate(ScheduleTrainBase):
    pass


class ScheduleTrainUpdate(BaseModel):
    train_id: Optional[int] = None
    scheduled_departure: Optional[datetime] = None
    actual_departure: Optional[datetime] = None
    scheduled_arrival: Optional[datetime] = None
    actual_arrival: Optional[datetime] = None
    turnaround_time: Optional[int] = None
    is_completed: Optional[bool] = None


class ScheduleTrainResponse(ScheduleTrainBase):
    id: int
    actual_departure: Optional[datetime] = None
    actual_arrival: Optional[datetime] = None
    turnaround_time: Optional[int] = None
    is_completed: bool
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class PassengerDataBase(BaseModel):
    line_id: int
    station_id: Optional[int] = None
    timestamp: datetime
    passenger_count: int = 0


class PassengerDataCreate(PassengerDataBase):
    pass


class PassengerDataResponse(PassengerDataBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class DispatchLogBase(BaseModel):
    level: LogLevel = LogLevel.INFO
    message: str
    entity_type: Optional[str] = None
    entity_id: Optional[int] = None
    operator: Optional[str] = None


class DispatchLogCreate(DispatchLogBase):
    pass


class DispatchLogResponse(DispatchLogBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class ConflictWarning(BaseModel):
    type: str
    message: str
    schedule_train_id: int
    schedule_train_sequence: int
    previous_departure: datetime
    current_departure: datetime
    interval_seconds: int
    min_safe_interval: int


class SignalFaultReport(BaseModel):
    signal_id: int
    signal_code: str
    fault_time: datetime
    fault_description: Optional[str] = None
    affected_section_id: Optional[int] = None
    affected_section_name: Optional[str] = None


class PowerOutageReport(BaseModel):
    section_id: int
    section_code: str
    section_name: str
    outage_time: datetime
    needs_confirmation: bool = True
