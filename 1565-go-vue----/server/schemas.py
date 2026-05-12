from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional, List
from .models import CapacityStatus, FlightType


class WaypointBase(BaseModel):
    code: str = Field(..., max_length=20)
    name: Optional[str] = None
    latitude: float
    longitude: float


class WaypointCreate(WaypointBase):
    pass


class WaypointUpdate(BaseModel):
    code: Optional[str] = None
    name: Optional[str] = None
    latitude: Optional[float] = None
    longitude: Optional[float] = None


class WaypointResponse(WaypointBase):
    id: int
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class RouteWaypointBase(BaseModel):
    waypoint_id: int
    sequence: int
    estimated_time_minutes: int = 0


class RouteWaypointCreate(RouteWaypointBase):
    pass


class RouteWaypointResponse(RouteWaypointBase):
    id: int
    waypoint: Optional[WaypointResponse] = None

    class Config:
        from_attributes = True


class RouteBase(BaseModel):
    code: str
    name: Optional[str] = None
    capacity: int = 10
    busy_threshold: float = 0.8
    saturated_threshold: float = 1.0
    min_interval_minutes: int = 10
    average_speed: float = 800.0


class RouteCreate(RouteBase):
    waypoints: List[RouteWaypointCreate] = []


class RouteUpdate(BaseModel):
    code: Optional[str] = None
    name: Optional[str] = None
    capacity: Optional[int] = None
    busy_threshold: Optional[float] = None
    saturated_threshold: Optional[float] = None
    min_interval_minutes: Optional[int] = None
    average_speed: Optional[float] = None
    waypoints: Optional[List[RouteWaypointCreate]] = None


class RouteResponse(RouteBase):
    id: int
    status: CapacityStatus
    created_at: datetime
    updated_at: Optional[datetime] = None
    waypoints: List[RouteWaypointResponse] = []

    class Config:
        from_attributes = True


class FlightBase(BaseModel):
    flight_number: str
    route_id: int
    direction: str
    altitude: float
    flight_type: FlightType = FlightType.DOMESTIC
    estimated_entry_time: datetime


class FlightCreate(FlightBase):
    pass


class FlightUpdate(BaseModel):
    altitude: Optional[float] = None
    estimated_entry_time: Optional[datetime] = None
    actual_entry_time: Optional[datetime] = None
    exit_time: Optional[datetime] = None
    is_waiting: Optional[bool] = None
    is_completed: Optional[bool] = None


class FlightResponse(FlightBase):
    id: int
    is_waiting: bool = False
    is_completed: bool = False
    actual_entry_time: Optional[datetime] = None
    exit_time: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class ConflictResponse(BaseModel):
    id: int
    route_id: int
    flight1_id: int
    flight2_id: int
    conflict_type: str
    description: Optional[str] = None
    detected_at: datetime
    resolved: bool = False
    resolution_suggestion: Optional[str] = None

    class Config:
        from_attributes = True


class FlowRecordResponse(BaseModel):
    id: int
    route_id: int
    timestamp: datetime
    active_flights: int
    waiting_flights: int
    capacity: int
    utilization: float
    status: CapacityStatus

    class Config:
        from_attributes = True


class RouteStatus(BaseModel):
    route_id: int
    route_code: str
    current_status: CapacityStatus
    active_flights: int
    waiting_flights: int
    capacity: int
    utilization: float
    unresolved_conflicts: int = 0


class SimulationInput(BaseModel):
    flights: List[FlightCreate]


class SimulationReport(BaseModel):
    total_flights: int
    conflicts: List[ConflictResponse]
    suggestions: List[str]
