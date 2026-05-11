from datetime import datetime, date
from typing import Optional, List
from enum import Enum
from pydantic import BaseModel, Field


class Project(BaseModel):
    id: Optional[int] = None
    name: str
    duration_minutes: int
    price: float


class StylistStatus(str, Enum):
    AVAILABLE = "available"
    BUSY = "busy"
    OFF_DUTY = "off_duty"


class Stylist(BaseModel):
    id: Optional[int] = None
    name: str
    status: StylistStatus = StylistStatus.AVAILABLE
    skilled_projects: List[int] = []


class Client(BaseModel):
    id: Optional[int] = None
    name: str
    phone: str
    pets: List[str] = []


class AppointmentStatus(str, Enum):
    PENDING = "pending"
    CONFIRMED = "confirmed"
    COMPLETED = "completed"
    CANCELLED = "cancelled"
    WAITLIST = "waitlist"


class Appointment(BaseModel):
    id: Optional[int] = None
    client_id: int
    pet_name: str
    project_id: int
    stylist_id: int
    appointment_date: date
    start_time: str
    end_time: str
    status: AppointmentStatus = AppointmentStatus.CONFIRMED
    actual_duration_minutes: Optional[int] = None
    is_overtime: bool = False
    created_at: datetime = Field(default_factory=datetime.now)


class Review(BaseModel):
    id: Optional[int] = None
    appointment_id: int
    client_id: int
    rating: int
    comment: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class Supply(BaseModel):
    id: Optional[int] = None
    name: str
    current_stock: int
    min_stock: int
    unit: str


class SupplyUsage(BaseModel):
    id: Optional[int] = None
    supply_id: int
    appointment_id: int
    quantity: int
    used_at: datetime = Field(default_factory=datetime.now)


class PurchaseTodo(BaseModel):
    id: Optional[int] = None
    supply_id: int
    quantity_needed: int
    created_at: datetime = Field(default_factory=datetime.now)
    is_completed: bool = False
