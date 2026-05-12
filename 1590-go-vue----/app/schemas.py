from datetime import datetime, date, time
from typing import Optional, List
from pydantic import BaseModel, Field
from app.models import CabinStatus, MaintenanceType, MaintenanceStatus


class QueueRealtimeResponse(BaseModel):
    total_queue: int
    upward_passengers: int
    effective_queue: int
    queue_groups: int
    average_wait_seconds: float
    first_join_time: Optional[datetime] = None


class QueueEstimateResponse(BaseModel):
    estimated_wait_minutes: float
    queue_size: int
    current_interval_minutes: float
    next_dispatch_minutes: float


class CabinStatusResponse(BaseModel):
    cabin_number: str
    status: CabinStatus
    current_weight: float
    max_weight: float
    passenger_count: int
    last_dispatch_time: Optional[datetime] = None
    sensor_status_ok: bool


class DispatchRequest(BaseModel):
    is_first_trip: bool = False
    has_passengers: bool = True


class DispatchResponse(BaseModel):
    success: bool
    message: str
    cabin_number: str
    dispatch_time: datetime
    weight: float
    weight_is_estimated: bool
    passenger_count: Optional[int] = None
    weight_status: str
    direction: str


class ArriveResponse(BaseModel):
    success: bool
    message: str
    cabin_number: str
    arrive_time: datetime


class SensorWeightRequest(BaseModel):
    weight: float


class SensorWeightResponse(BaseModel):
    cabin_number: str
    current_weight: float
    max_weight: float
    weight_ratio: float
    status: str
    can_dispatch: bool


class SensorStatusRequest(BaseModel):
    status_ok: bool


class SensorStatusResponse(BaseModel):
    cabin_number: str
    sensor_status_ok: bool
    message: str


class MaintenanceStartResponse(BaseModel):
    task_id: int
    status: MaintenanceStatus
    start_time: datetime
    message: str


class MaintenanceCompleteRequest(BaseModel):
    notes: Optional[str] = None


class MaintenanceCompleteResponse(BaseModel):
    task_id: int
    status: MaintenanceStatus
    complete_time: datetime
    message: str


class DailyStatsResponse(BaseModel):
    date: date
    total_trips: int
    total_passengers: Optional[int]
    total_weight: Optional[float]
    average_weight_per_trip: Optional[float]
    peak_hour_passengers: Optional[int]
    maintenance_completed: int
    maintenance_overdue: int


class ShiftStatsItem(BaseModel):
    shift_number: int
    start_time: datetime
    end_time: Optional[datetime] = None
    total_passengers: Optional[int]
    total_weight: Optional[float]
    total_trips: Optional[int]
    total_revenue: Optional[float]
    average_wait_time: Optional[float]


class ShiftStatsResponse(BaseModel):
    date: date
    shifts: List[ShiftStatsItem]
