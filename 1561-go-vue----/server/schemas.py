from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, ConfigDict
from .models import FlightStatus, AircraftType, VehicleStatus, VehicleType


class StandBase(BaseModel):
    code: str
    is_bridge: bool = False
    supports_wide_body: bool = False


class StandCreate(StandBase):
    pass


class StandUpdate(BaseModel):
    code: Optional[str] = None
    is_bridge: Optional[bool] = None
    supports_wide_body: Optional[bool] = None


class StandResponse(StandBase):
    id: int
    is_occupied: bool
    model_config = ConfigDict(from_attributes=True)


class GateBase(BaseModel):
    code: str


class GateCreate(GateBase):
    pass


class GateAdjacencyCreate(BaseModel):
    gate_id_1: int
    gate_id_2: int


class GateResponse(GateBase):
    id: int
    is_available: bool
    model_config = ConfigDict(from_attributes=True)


class FlightBase(BaseModel):
    flight_number: str
    aircraft_type: AircraftType
    scheduled_departure: datetime
    scheduled_arrival: Optional[datetime] = None
    check_in_start: Optional[datetime] = None
    check_in_end: Optional[datetime] = None
    boarding_start: Optional[datetime] = None
    boarding_end: Optional[datetime] = None


class FlightCreate(FlightBase):
    pass


class FlightUpdate(BaseModel):
    flight_number: Optional[str] = None
    aircraft_type: Optional[AircraftType] = None
    scheduled_departure: Optional[datetime] = None
    scheduled_arrival: Optional[datetime] = None
    check_in_start: Optional[datetime] = None
    check_in_end: Optional[datetime] = None
    boarding_start: Optional[datetime] = None
    boarding_end: Optional[datetime] = None
    actual_departure: Optional[datetime] = None
    actual_arrival: Optional[datetime] = None
    status: Optional[FlightStatus] = None


class FlightResponse(FlightBase):
    id: int
    status: FlightStatus
    actual_departure: Optional[datetime] = None
    actual_arrival: Optional[datetime] = None
    stand_id: Optional[int] = None
    gate_id: Optional[int] = None
    model_config = ConfigDict(from_attributes=True)


class StandAssignRequest(BaseModel):
    flight_id: int


class GateAssignRequest(BaseModel):
    flight_id: int


class VehicleBase(BaseModel):
    vehicle_code: str
    vehicle_type: VehicleType


class VehicleCreate(VehicleBase):
    pass


class VehicleUpdate(BaseModel):
    vehicle_code: Optional[str] = None
    vehicle_type: Optional[VehicleType] = None
    status: Optional[VehicleStatus] = None


class VehicleResponse(VehicleBase):
    id: int
    status: VehicleStatus
    model_config = ConfigDict(from_attributes=True)


class VehicleDispatchBase(BaseModel):
    flight_id: int
    priority: int = 0


class VehicleDispatchCreate(VehicleDispatchBase):
    vehicle_type: VehicleType


class VehicleDispatchResponse(BaseModel):
    id: int
    vehicle_id: Optional[int] = None
    flight_id: int
    priority: int
    is_waiting: bool
    assigned_at: datetime
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    model_config = ConfigDict(from_attributes=True)


class VehicleMaintenanceBase(BaseModel):
    description: str
    scheduled_at: Optional[datetime] = None


class VehicleMaintenanceCreate(VehicleMaintenanceBase):
    pass


class VehicleMaintenanceUpdate(BaseModel):
    description: Optional[str] = None
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    is_completed: Optional[bool] = None


class VehicleMaintenanceResponse(BaseModel):
    id: int
    vehicle_id: int
    description: str
    scheduled_at: datetime
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    is_completed: bool
    model_config = ConfigDict(from_attributes=True)
