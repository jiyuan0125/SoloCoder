from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional, List
from app.models import Direction, ScheduleStatus


class StationBase(BaseModel):
    name: str
    code: str
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    is_transfer: bool = False


class StationCreate(StationBase):
    pass


class StationUpdate(BaseModel):
    name: Optional[str] = None
    code: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None
    is_transfer: Optional[bool] = None


class StationResponse(StationBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RouteBase(BaseModel):
    name: str
    code: str
    start_station: str
    end_station: str


class RouteCreate(RouteBase):
    pass


class RouteUpdate(BaseModel):
    name: Optional[str] = None
    code: Optional[str] = None
    start_station: Optional[str] = None
    end_station: Optional[str] = None


class RouteResponse(RouteBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RouteStationBase(BaseModel):
    route_id: int
    station_id: int
    direction: Direction
    sequence: int
    travel_time_from_prev: int = 2
    stop_time: int = 1


class RouteStationCreate(RouteStationBase):
    pass


class RouteStationUpdate(BaseModel):
    sequence: Optional[int] = None
    travel_time_from_prev: Optional[int] = None
    stop_time: Optional[int] = None


class RouteStationResponse(RouteStationBase):
    id: int
    station_name: Optional[str] = None
    station_code: Optional[str] = None

    class Config:
        from_attributes = True


class VehicleBase(BaseModel):
    plate_number: str
    vehicle_code: str
    capacity: int = 80


class VehicleCreate(VehicleBase):
    pass


class VehicleUpdate(BaseModel):
    plate_number: Optional[str] = None
    vehicle_code: Optional[str] = None
    capacity: Optional[int] = None
    current_route_id: Optional[int] = None
    current_direction: Optional[Direction] = None
    status: Optional[str] = None
    current_load: Optional[int] = None


class VehicleResponse(VehicleBase):
    id: int
    current_route_id: Optional[int]
    current_direction: Optional[Direction]
    status: str
    current_load: int
    is_crowded: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class SignalPriorityBase(BaseModel):
    station_id: int
    intersection_name: str
    priority_interval: int
    is_active: bool = True


class SignalPriorityCreate(SignalPriorityBase):
    pass


class SignalPriorityUpdate(BaseModel):
    intersection_name: Optional[str] = None
    priority_interval: Optional[int] = None
    is_active: Optional[bool] = None


class SignalPriorityResponse(SignalPriorityBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class OperationParameterBase(BaseModel):
    route_id: int
    turnaround_time: int = 10
    default_interval: int = 10
    start_time: str = "06:00"
    end_time: str = "22:00"
    max_signal_priority_daily: int = 20


class OperationParameterCreate(OperationParameterBase):
    pass


class OperationParameterUpdate(BaseModel):
    turnaround_time: Optional[int] = None
    default_interval: Optional[int] = None
    start_time: Optional[str] = None
    end_time: Optional[str] = None
    max_signal_priority_daily: Optional[int] = None


class OperationParameterResponse(OperationParameterBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ScheduleBase(BaseModel):
    route_id: int
    vehicle_id: int
    direction: Direction
    start_time: datetime
    end_time: datetime


class ScheduleResponse(ScheduleBase):
    id: int
    status: ScheduleStatus
    schedule_data: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ScheduleStation(BaseModel):
    station_id: int
    station_name: str
    sequence: int
    planned_arrival_time: datetime


class ScheduleDetailResponse(BaseModel):
    schedule: ScheduleResponse
    stations: List[ScheduleStation]


class ArrivalRecordBase(BaseModel):
    schedule_id: int
    station_id: int
    planned_arrival_time: datetime
    actual_arrival_time: Optional[datetime] = None


class ArrivalRecordCreate(BaseModel):
    schedule_id: int
    station_id: int
    actual_arrival_time: datetime


class ArrivalRecordResponse(ArrivalRecordBase):
    id: int
    is_on_time: Optional[bool]
    created_at: datetime

    class Config:
        from_attributes = True


class SignalPriorityRequestResponse(BaseModel):
    id: int
    vehicle_id: int
    intersection_name: str
    request_time: datetime
    granted: bool

    class Config:
        from_attributes = True


class PunctualityReport(BaseModel):
    route_id: int
    route_name: str
    total_schedules: int
    on_time_count: int
    punctuality_rate: float


class SignalPriorityRequest(BaseModel):
    vehicle_id: int
    intersection_name: str
    distance: float


class ScheduleGenerationRequest(BaseModel):
    route_id: int
    date: str
