from pydantic import BaseModel, Field
from datetime import date, datetime
from typing import Optional, List
from enum import Enum


class StylistStatus(str, Enum):
    AVAILABLE = "available"
    BUSY = "busy"
    OFF_DUTY = "off_duty"


class AppointmentStatus(str, Enum):
    PENDING = "pending"
    CONFIRMED = "confirmed"
    COMPLETED = "completed"
    CANCELLED = "cancelled"
    WAITLIST = "waitlist"


class ProjectCreate(BaseModel):
    name: str
    duration_minutes: int
    price: float


class ProjectResponse(BaseModel):
    id: int
    name: str
    duration_minutes: int
    price: float

    class Config:
        from_attributes = True


class StylistCreate(BaseModel):
    name: str
    status: StylistStatus = StylistStatus.AVAILABLE
    skilled_projects: List[int] = []


class StylistResponse(BaseModel):
    id: int
    name: str
    status: StylistStatus
    skilled_projects: List[int]

    class Config:
        from_attributes = True


class ClientCreate(BaseModel):
    name: str
    phone: str
    pets: List[str] = []


class ClientResponse(BaseModel):
    id: int
    name: str
    phone: str
    pets: List[str]

    class Config:
        from_attributes = True


class AppointmentCreate(BaseModel):
    client_id: int
    pet_name: str
    project_id: int
    stylist_id: int
    appointment_date: date
    start_time: str


class AppointmentResponse(BaseModel):
    id: int
    client_id: int
    pet_name: str
    project_id: int
    stylist_id: int
    appointment_date: date
    start_time: str
    end_time: str
    status: AppointmentStatus
    actual_duration_minutes: Optional[int] = None
    is_overtime: bool = False
    created_at: datetime

    class Config:
        from_attributes = True


class AppointmentComplete(BaseModel):
    actual_duration_minutes: int


class ReviewCreate(BaseModel):
    appointment_id: int
    client_id: int
    rating: int
    comment: Optional[str] = None


class ReviewResponse(BaseModel):
    id: int
    appointment_id: int
    client_id: int
    rating: int
    comment: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class SupplyCreate(BaseModel):
    name: str
    current_stock: int
    min_stock: int
    unit: str


class SupplyResponse(BaseModel):
    id: int
    name: str
    current_stock: int
    min_stock: int
    unit: str

    class Config:
        from_attributes = True


class PurchaseTodoResponse(BaseModel):
    id: int
    supply_id: int
    quantity_needed: int
    created_at: datetime
    is_completed: bool

    class Config:
        from_attributes = True
