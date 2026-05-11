from datetime import datetime
from enum import Enum
from typing import List, Optional
from pydantic import BaseModel, Field


class Location(BaseModel):
    latitude: float
    longitude: float


class Origin(BaseModel):
    id: str
    name: str
    location: Location
    address: str


class VehicleStatus(str, Enum):
    IDLE = "idle"
    IN_TRANSIT = "in_transit"
    MAINTENANCE = "maintenance"


class Vehicle(BaseModel):
    id: str
    plate_number: str
    current_temperature: float
    max_load: float
    current_load: float = 0.0
    status: VehicleStatus = VehicleStatus.IDLE
    location: Optional[Location] = None
    last_report_time: Optional[datetime] = None
    current_task_ids: List[str] = Field(default_factory=list)


class Order(BaseModel):
    id: str
    origin_id: str
    destination: Location
    destination_address: str
    weight: float
    min_temperature: float
    max_temperature: float
    created_at: datetime


class DeliveryTaskStatus(str, Enum):
    PENDING = "pending"
    ASSIGNED = "assigned"
    IN_TRANSIT = "in_transit"
    COMPLETED = "completed"
    TEMPERATURE_ABNORMAL = "temperature_abnormal"
    COMMUNICATION_LOST = "communication_lost"


class TemperatureReport(BaseModel):
    timestamp: datetime
    temperature: float
    location: Location


class DeliveryTask(BaseModel):
    id: str
    order_id: str
    vehicle_id: Optional[str] = None
    origin_id: str
    destination: Location
    destination_address: str
    weight: float
    min_temperature: float
    max_temperature: float
    status: DeliveryTaskStatus = DeliveryTaskStatus.PENDING
    temperature_reports: List[TemperatureReport] = Field(default_factory=list)
    created_at: datetime
    assigned_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
