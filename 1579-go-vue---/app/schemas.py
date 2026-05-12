from pydantic import BaseModel, Field
from datetime import datetime
from typing import List, Optional
from app.models import (
    TransportMode, ContainerSize, ContainerStatus, 
    SegmentStatus, VehicleStatus, OrderStatus
)


class StationBase(BaseModel):
    code: str
    name: str
    city: str
    country: str


class StationCreate(StationBase):
    pass


class StationResponse(StationBase):
    id: int

    class Config:
        from_attributes = True


class RouteSegmentBase(BaseModel):
    sequence: int
    mode: TransportMode
    origin_station_id: int
    destination_station_id: int
    estimated_hours: float = 0
    is_transit_point: int = 0


class RouteBase(BaseModel):
    code: str
    name: str
    origin_station_id: int
    destination_station_id: int


class RouteCreate(RouteBase):
    segments: List[RouteSegmentBase]


class RouteSegmentResponse(RouteSegmentBase):
    id: int
    origin_station: Optional[StationResponse] = None
    destination_station: Optional[StationResponse] = None

    class Config:
        from_attributes = True


class RouteResponse(RouteBase):
    id: int
    segments: List[RouteSegmentResponse] = []

    class Config:
        from_attributes = True


class ContainerBase(BaseModel):
    container_number: str
    size: ContainerSize


class ContainerCreate(ContainerBase):
    current_station_id: Optional[int] = None


class ContainerResponse(ContainerBase):
    id: int
    status: ContainerStatus
    current_station_id: Optional[int] = None
    current_vehicle_id: Optional[int] = None

    class Config:
        from_attributes = True


class VehicleBase(BaseModel):
    vehicle_id: str
    name: str
    mode: TransportMode
    max_container_size: ContainerSize


class VehicleCreate(VehicleBase):
    current_station_id: Optional[int] = None


class VehicleResponse(VehicleBase):
    id: int
    status: VehicleStatus
    current_station_id: Optional[int] = None

    class Config:
        from_attributes = True


class OrderContainerCreate(BaseModel):
    container_number: str


class OrderCreate(BaseModel):
    origin_station_code: str
    destination_station_code: str
    deadline: datetime
    containers: List[str]


class SegmentResponse(BaseModel):
    id: int
    segment_number: str
    sequence: int
    mode: TransportMode
    status: SegmentStatus
    is_transit_point: int
    origin_station: Optional[StationResponse] = None
    destination_station: Optional[StationResponse] = None
    planned_departure: Optional[datetime] = None
    planned_arrival: Optional[datetime] = None
    actual_departure: Optional[datetime] = None
    actual_arrival: Optional[datetime] = None
    assigned_vehicle: Optional[VehicleResponse] = None

    class Config:
        from_attributes = True


class OrderResponse(BaseModel):
    id: int
    order_number: str
    status: OrderStatus
    origin_station: Optional[StationResponse] = None
    destination_station: Optional[StationResponse] = None
    deadline: datetime
    route: Optional[RouteResponse] = None
    segments: List[SegmentResponse] = []
    containers: List[ContainerResponse] = []

    class Config:
        from_attributes = True


class ContainerStatusResponse(BaseModel):
    container_number: str
    status: ContainerStatus
    current_station: Optional[StationResponse] = None
    current_vehicle: Optional[VehicleResponse] = None
    last_update: datetime


class OperationResponse(BaseModel):
    success: bool
    message: str
    timestamp: datetime = Field(default_factory=datetime.utcnow)


class TransportPlanResponse(BaseModel):
    order_number: str
    status: OrderStatus
    deadline: datetime
    segments: List[dict]
    containers: List[dict]


class TrajectoryResponse(BaseModel):
    container_number: str
    events: List[dict]
