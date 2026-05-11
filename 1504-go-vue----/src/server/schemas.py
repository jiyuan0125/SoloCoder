from datetime import datetime, date
from typing import Optional
from pydantic import BaseModel, Field

class BatchCreate(BaseModel):
    breed: str
    initial_count: int = Field(gt=0)
    gestation_period_days: int = Field(default=114, gt=0)

class BatchResponse(BaseModel):
    id: int
    breed: str
    initial_count: int
    current_count: int
    status: str
    created_at: datetime
    gestation_period_days: int

    class Config:
        from_attributes = True

class BreedingRecordCreate(BaseModel):
    batch_id: int
    female_id: str
    breeding_date: date

class BreedingRecordResponse(BaseModel):
    id: int
    batch_id: int
    female_id: str
    breeding_date: date
    status: str
    expected_birth_date: Optional[date]
    failure_reason: Optional[str]
    created_at: datetime

    class Config:
        from_attributes = True

class BreedingFailureRequest(BaseModel):
    failure_reason: str

class VaccinationRecordCreate(BaseModel):
    batch_id: int
    vaccine_name: str
    vaccination_date: date
    interval_days: Optional[int] = Field(default=None, ge=0)

class VaccinationRecordResponse(BaseModel):
    id: int
    batch_id: int
    vaccine_name: str
    vaccination_date: date
    next_due_date: Optional[date]
    created_at: datetime

    class Config:
        from_attributes = True

class SlaughterRecordCreate(BaseModel):
    batch_id: int
    count: int = Field(gt=0)
    avg_weight: float = Field(gt=0)
    unit_price: float = Field(ge=0)
    slaughter_date: date

class SlaughterRecordResponse(BaseModel):
    id: int
    batch_id: int
    count: int
    avg_weight: float
    unit_price: float
    slaughter_date: date
    created_at: datetime

    class Config:
        from_attributes = True

class ReminderResponse(BaseModel):
    id: int
    type: str
    message: str
    due_date: date
    is_read: bool
    created_at: datetime

    class Config:
        from_attributes = True

class DashboardResponse(BaseModel):
    total_stock: int
    monthly_input_count: int
    monthly_slaughter_count: int
    breed_distribution: dict
    total_output_value: float

    class Config:
        from_attributes = True
