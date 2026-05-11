from __future__ import annotations

from datetime import datetime
from enum import Enum
from typing import List, Optional

from pydantic import BaseModel, Field


class VehicleStatus(str, Enum):
    IDLE = "idle"
    IN_TRANSIT = "in_transit"
    MAINTENANCE = "maintenance"


class TaskStatus(str, Enum):
    PENDING = "pending"
    ASSIGNED = "assigned"
    IN_TRANSIT = "in_transit"
    COMPLETED = "completed"
    TEMPERATURE_ABNORMAL = "temperature_abnormal"
    COMMUNICATION_LOST = "communication_lost"


class Location(BaseModel):
    id: str
    name: str
    latitude: float
    longitude: float
    is_origin: bool = True


class Vehicle(BaseModel):
    id: str
    vehicle_number: str
    current_temperature: float
    max_load: float
    current_load: float = 0.0
    status: VehicleStatus = VehicleStatus.IDLE
    current_location_id: Optional[str] = None
    assigned_tasks: List[str] = Field(default_factory=list)
    last_report_time: Optional[datetime] = None


class TemperatureConflict(Exception):
    pass


class DeliveryTask(BaseModel):
    id: str
    order_id: str
    origin_id: str
    destination_id: str
    weight: float
    required_min_temp: float
    required_max_temp: float
    status: TaskStatus = TaskStatus.PENDING
    assigned_vehicle_id: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.utcnow)
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    temperature_reports: List[TemperatureReport] = Field(default_factory=list)
    location_reports: List[LocationReport] = Field(default_factory=list)


class TemperatureReport(BaseModel):
    task_id: str
    vehicle_id: str
    temperature: float
    reported_at: datetime = Field(default_factory=datetime.utcnow)
    is_normal: bool = True


class LocationReport(BaseModel):
    task_id: str
    vehicle_id: str
    latitude: float
    longitude: float
    reported_at: datetime = Field(default_factory=datetime.utcnow)


class DeliveryRecord(BaseModel):
    task_id: str
    order_id: str
    vehicle_id: str
    vehicle_number: str
    origin_name: str
    destination_name: str
    weight: float
    required_min_temp: float
    required_max_temp: float
    created_at: datetime
    started_at: Optional[datetime]
    completed_at: Optional[datetime]
    final_status: TaskStatus
