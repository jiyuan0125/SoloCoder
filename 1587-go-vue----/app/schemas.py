from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime
from app.models import (
    StationStatus, ChargerStatus, VehicleStatus, RouteStatus,
    AssignmentStatus, AlertType, AlertStatus, NotificationType
)

class StationBase(BaseModel):
    name: str
    location: Optional[str] = None

class StationCreate(StationBase):
    pass

class StationUpdate(BaseModel):
    name: Optional[str] = None
    status: Optional[StationStatus] = None
    location: Optional[str] = None

class StationResponse(StationBase):
    id: int
    status: StationStatus
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class ChargerBase(BaseModel):
    name: str
    station_id: Optional[int] = None
    power_kw: float = 50.0

class ChargerCreate(ChargerBase):
    pass

class ChargerUpdate(BaseModel):
    name: Optional[str] = None
    station_id: Optional[int] = None
    status: Optional[ChargerStatus] = None
    power_kw: Optional[float] = None

class ChargerResponse(ChargerBase):
    id: int
    status: ChargerStatus
    total_charging_count: int
    locked_by_vehicle_id: Optional[int] = None
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class VehicleBase(BaseModel):
    plate_number: str
    max_battery_capacity_kwh: float = 80.0

class VehicleCreate(VehicleBase):
    pass

class VehicleUpdate(BaseModel):
    status: Optional[VehicleStatus] = None
    current_battery: Optional[float] = None
    current_route_id: Optional[int] = None
    current_station_id: Optional[int] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None

class VehicleStatusReport(BaseModel):
    latitude: float
    longitude: float
    current_battery: float = Field(..., ge=0, le=100)

class VehicleResponse(VehicleBase):
    id: int
    status: VehicleStatus
    current_battery: float
    current_route_id: Optional[int] = None
    current_station_id: Optional[int] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    last_report_time: Optional[datetime] = None
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class RouteBase(BaseModel):
    name: str
    total_distance_km: float = 0.0
    estimated_duration_min: int = 0
    start_station_id: Optional[int] = None
    end_station_id: Optional[int] = None

class RouteCreate(RouteBase):
    pass

class RouteUpdate(BaseModel):
    name: Optional[str] = None
    status: Optional[RouteStatus] = None
    total_distance_km: Optional[float] = None
    estimated_duration_min: Optional[int] = None
    start_station_id: Optional[int] = None
    end_station_id: Optional[int] = None

class RouteResponse(RouteBase):
    id: int
    status: RouteStatus
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class ScheduleBase(BaseModel):
    route_id: int
    vehicle_id: Optional[int] = None
    departure_time: datetime
    return_time: datetime
    is_last_run: bool = False

class ScheduleCreate(ScheduleBase):
    pass

class ScheduleResponse(ScheduleBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True

class ChargingSessionResponse(BaseModel):
    id: int
    vehicle_id: int
    charger_id: int
    start_time: datetime
    end_time: Optional[datetime] = None
    start_battery: float
    target_battery: float
    current_battery: Optional[float] = None
    estimated_end_time: Optional[datetime] = None
    status: AssignmentStatus
    created_at: datetime

    class Config:
        from_attributes = True

class AssignmentResponse(BaseModel):
    id: int
    vehicle_id: int
    route_id: Optional[int] = None
    schedule_id: Optional[int] = None
    status: AssignmentStatus
    reason: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True

class AlertResponse(BaseModel):
    id: int
    type: AlertType
    vehicle_id: Optional[int] = None
    charger_id: Optional[int] = None
    message: str
    status: AlertStatus
    notification_type: NotificationType
    created_at: datetime
    resolved_at: Optional[datetime] = None

    class Config:
        from_attributes = True

class DispatchLogResponse(BaseModel):
    id: int
    level: str
    message: str
    related_vehicle_id: Optional[int] = None
    related_charger_id: Optional[int] = None
    created_at: datetime

    class Config:
        from_attributes = True

class ChargingProgress(BaseModel):
    session_id: int
    vehicle_id: int
    charger_id: int
    current_battery: float
    target_battery: float
    progress_percent: float
    estimated_end_time: Optional[datetime] = None
