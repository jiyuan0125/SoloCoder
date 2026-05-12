from pydantic import BaseModel, Field
from typing import Optional
from datetime import datetime


class FlightCreate(BaseModel):
    flight_number: str
    route: str
    scheduled_departure: datetime
    scheduled_arrival: datetime
    unit_price: float
    min_load_rate: Optional[float] = 0.60


class FlightResponse(BaseModel):
    id: int
    flight_number: str
    route: str
    scheduled_departure: datetime
    scheduled_arrival: datetime
    unit_price: float
    min_load_rate: float

    class Config:
        from_attributes = True


class CompartmentCreate(BaseModel):
    compartment_number: str
    flight_number: str
    position: str
    total_capacity_weight: float
    is_temperature_controlled: Optional[bool] = False


class CompartmentResponse(BaseModel):
    id: int
    compartment_number: str
    flight_id: int
    position: str
    total_capacity_weight: float
    remaining_capacity_weight: float
    is_temperature_controlled: bool

    class Config:
        from_attributes = True


class CargoAccept(BaseModel):
    cargo_number: str
    shipper: str
    consignee: str
    cargo_type: str
    description: Optional[str] = None
    declaration_number: Optional[str] = None
    quarantine_certificate_number: Optional[str] = None


class CargoWeigh(BaseModel):
    actual_weight: float
    volume: Optional[float] = None
    volume_weight: Optional[float] = None


class CargoResponse(BaseModel):
    id: int
    cargo_number: str
    shipper: str
    consignee: str
    cargo_type: str
    description: Optional[str] = None
    actual_weight: Optional[float] = None
    volume: Optional[float] = None
    volume_weight: Optional[float] = None
    chargeable_weight: Optional[float] = None
    price_per_kg: Optional[float] = None
    freight_charge: Optional[float] = None
    storage_charge: float
    arrival_time: Optional[datetime] = None
    status: str
    declaration_number: Optional[str] = None
    quarantine_certificate_number: Optional[str] = None

    class Config:
        from_attributes = True


class LoadAssignmentResponse(BaseModel):
    id: int
    cargo_id: int
    cargo_number: str
    compartment_id: int
    compartment_number: str
    assigned_weight: float
    is_confirmed: bool
    confirmed_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class DispatcherCreate(BaseModel):
    name: str
    contact: Optional[str] = None


class DispatcherResponse(BaseModel):
    id: int
    name: str
    contact: Optional[str] = None
    is_active: bool

    class Config:
        from_attributes = True
