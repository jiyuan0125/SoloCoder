from pydantic import BaseModel
from datetime import datetime
from typing import Optional, List
from .models import CarType, CargoType, OrderStatus


class OrderCreate(BaseModel):
    cargo_type: CargoType
    cargo_name: str
    weight_tons: float
    origin_station: str
    dest_station: str


class AssignRequest(BaseModel):
    wagon_id: int


class LoadRequest(BaseModel):
    train_id: int


class TrainCreate(BaseModel):
    train_no: str
    origin_station: str
    dest_station: str
    line_traction_tons: float
    max_wagons: int = 50


class TrainAddWagon(BaseModel):
    wagon_id: int
    order_id: Optional[int] = None


class StationCreate(BaseModel):
    code: str
    name: str
    line_traction_tons: float


class WagonCreate(BaseModel):
    wagon_number: str
    car_type: CarType
    capacity_tons: float
    current_station: str


class OrderResponse(BaseModel):
    id: int
    order_no: str
    cargo_type: CargoType
    cargo_name: str
    weight_tons: float
    origin_station: str
    dest_station: str
    status: OrderStatus
    assigned_wagon_id: Optional[int]
    train_id: Optional[int]
    created_at: datetime
    accepted_at: Optional[datetime]
    assigned_at: Optional[datetime]
    loaded_at: Optional[datetime]
    departed_at: Optional[datetime]
    arrived_at: Optional[datetime]
    delivered_at: Optional[datetime]
    cancelled_at: Optional[datetime]

    class Config:
        from_attributes = True


class WagonResponse(BaseModel):
    id: int
    wagon_number: str
    car_type: CarType
    capacity_tons: float
    current_station: str
    status: str
    current_train_id: Optional[int]

    class Config:
        from_attributes = True


class StationSummary(BaseModel):
    station_code: str
    empty_wagons: int
    loaded_wagons: int
    shipped_tons: float
    avg_turnaround_hours: float
    slow_turnaround_count: int


class TrainResponse(BaseModel):
    id: int
    train_no: str
    origin_station: str
    dest_station: str
    line_traction_tons: float
    max_wagons: int
    departure_time: Optional[datetime]
    arrival_time: Optional[datetime]
    wagon_count: int
    total_weight: float

    class Config:
        from_attributes = True
