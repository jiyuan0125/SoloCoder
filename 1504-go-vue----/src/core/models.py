from datetime import datetime, date
from typing import Optional, List
from enum import Enum
from pydantic import BaseModel, Field, validator

class BatchStatus(str, Enum):
    ACTIVE = "active"
    EMPTY = "empty"

class BreedingStatus(str, Enum):
    PENDING = "pending"
    SUCCESS = "success"
    FAILED = "failed"

class ReminderType(str, Enum):
    BIRTH = "birth"
    VACCINATION = "vaccination"

class Batch(BaseModel):
    id: int
    breed: str
    initial_count: int = Field(gt=0)
    current_count: int = Field(ge=0)
    status: BatchStatus = BatchStatus.ACTIVE
    created_at: datetime
    gestation_period_days: int = Field(default=114, gt=0)

    @validator('current_count')
    def current_count_not_exceed_initial(cls, v, values):
        if 'initial_count' in values and v > values['initial_count']:
            raise ValueError('当前存栏数不能超过入栏数量')
        return v

class BreedingRecord(BaseModel):
    id: int
    batch_id: int
    female_id: str
    breeding_date: date
    status: BreedingStatus = BreedingStatus.PENDING
    expected_birth_date: Optional[date] = None
    failure_reason: Optional[str] = None
    created_at: datetime

class VaccinationRecord(BaseModel):
    id: int
    batch_id: int
    vaccine_name: str
    vaccination_date: date
    next_due_date: Optional[date] = None
    created_at: datetime

class SlaughterRecord(BaseModel):
    id: int
    batch_id: int
    count: int = Field(gt=0)
    avg_weight: float = Field(gt=0)
    unit_price: float = Field(ge=0)
    slaughter_date: date
    created_at: datetime

class Reminder(BaseModel):
    id: int
    type: ReminderType
    related_record_id: int
    related_record_type: str
    message: str
    due_date: date
    is_read: bool = False
    created_at: datetime
