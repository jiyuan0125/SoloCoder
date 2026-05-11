from enum import Enum
from datetime import date, datetime
from typing import Optional, List, Dict
from pydantic import BaseModel, Field


class BatchStatus(str, Enum):
    ACTIVE = "active"
    EMPTY = "empty"


class BreedingStatus(str, Enum):
    PENDING = "pending"
    SUCCESS = "success"
    FAILED = "failed"
    PENDING_REBREED = "pending_rebreed"


class VaccinationType(str, Enum):
    ROUTINE = "routine"
    EMERGENCY = "emergency"


class ReminderType(str, Enum):
    BIRTH_IMMINENT = "birth_imminent"
    VACCINATION_DUE = "vaccination_due"


class Breed(BaseModel):
    id: str
    name: str
    pregnancy_days: int
    description: Optional[str] = None


class Batch(BaseModel):
    id: str
    breed_id: str
    entry_date: date
    entry_quantity: int
    current_stock: int
    status: BatchStatus = BatchStatus.ACTIVE
    notes: Optional[str] = None


class Breeding(BaseModel):
    id: str
    batch_id: str
    female_id: str
    breeding_date: date
    sire_id: Optional[str] = None
    expected_birth_date: Optional[date] = None
    status: BreedingStatus = BreedingStatus.PENDING
    failure_reason: Optional[str] = None
    actual_birth_date: Optional[date] = None
    offspring_count: Optional[int] = None
    notes: Optional[str] = None


class Vaccination(BaseModel):
    id: str
    batch_id: str
    vaccine_name: str
    vaccination_date: date
    next_vaccination_date: Optional[date] = None
    interval_days: Optional[int] = None
    type: VaccinationType = VaccinationType.ROUTINE
    administered_by: Optional[str] = None
    notes: Optional[str] = None


class Slaughter(BaseModel):
    id: str
    batch_id: str
    slaughter_date: date
    quantity: int
    average_weight: float
    unit_price: float
    total_value: Optional[float] = None
    notes: Optional[str] = None


class Reminder(BaseModel):
    id: str
    type: ReminderType
    related_id: str
    related_type: str
    message: str
    due_date: date
    created_at: datetime
    is_read: bool = False


class DashboardMetrics(BaseModel):
    total_stock: int
    monthly_entry: int
    monthly_slaughter: int
    breed_distribution: Dict[str, int]
    active_batches: int
    pending_reminders: int


class BatchCreate(BaseModel):
    breed_id: str
    entry_date: date
    entry_quantity: int
    notes: Optional[str] = None


class BatchUpdate(BaseModel):
    notes: Optional[str] = None


class BreedingCreate(BaseModel):
    batch_id: str
    female_id: str
    breeding_date: date
    sire_id: Optional[str] = None
    notes: Optional[str] = None


class BreedingUpdate(BaseModel):
    status: Optional[BreedingStatus] = None
    failure_reason: Optional[str] = None
    actual_birth_date: Optional[date] = None
    offspring_count: Optional[int] = None
    notes: Optional[str] = None


class VaccinationCreate(BaseModel):
    batch_id: str
    vaccine_name: str
    vaccination_date: date
    interval_days: Optional[int] = None
    type: VaccinationType = VaccinationType.ROUTINE
    administered_by: Optional[str] = None
    notes: Optional[str] = None


class SlaughterCreate(BaseModel):
    batch_id: str
    slaughter_date: date
    quantity: int
    average_weight: float
    unit_price: float
    notes: Optional[str] = None


class ReminderUpdate(BaseModel):
    is_read: bool
