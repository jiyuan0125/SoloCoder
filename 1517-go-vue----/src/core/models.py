from datetime import datetime, date
from enum import Enum
from typing import Optional
from pydantic import BaseModel, Field, field_validator


class AnimalSpecies(str, Enum):
    DOG = "dog"
    CAT = "cat"
    BIRD = "bird"
    RABBIT = "rabbit"
    OTHER = "other"


class HealthStatus(str, Enum):
    HEALTHY = "healthy"
    ISOLATION = "isolation"
    TREATMENT = "treatment"
    RECOVERING = "recovering"


class AdoptionStatus(str, Enum):
    PENDING_REVIEW = "pending_review"
    APPROVED = "approved"
    REJECTED = "rejected"
    SCHEDULED_INTERVIEW = "scheduled_interview"
    INTERVIEW_PASSED = "interview_passed"
    INTERVIEW_FAILED = "interview_failed"
    ADOPTED = "adopted"
    CANCELLED = "cancelled"


class FollowUpStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"
    OVERDUE = "overdue"


class DonationType(str, Enum):
    MONEY = "money"
    FOOD = "food"
    SUPPLIES = "supplies"
    OTHER = "other"


class AppointmentStatus(str, Enum):
    SCHEDULED = "scheduled"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class Animal(BaseModel):
    id: str
    name: str
    species: AnimalSpecies
    breed: str
    health_status: HealthStatus
    cage: str
    arrival_date: date
    description: Optional[str] = None
    is_adoptable: bool = True
    created_at: datetime = Field(default_factory=datetime.now)


class AnimalCreate(BaseModel):
    name: str
    species: AnimalSpecies
    breed: str
    health_status: HealthStatus
    cage: str
    arrival_date: date
    description: Optional[str] = None
    is_adoptable: bool = True


class AnimalUpdate(BaseModel):
    name: Optional[str] = None
    species: Optional[AnimalSpecies] = None
    breed: Optional[str] = None
    health_status: Optional[HealthStatus] = None
    cage: Optional[str] = None
    arrival_date: Optional[date] = None
    description: Optional[str] = None
    is_adoptable: Optional[bool] = None


class Adopter(BaseModel):
    id: str
    name: str
    phone: str
    address: str
    has_bad_record: bool = False
    created_at: datetime = Field(default_factory=datetime.now)


class AdopterCreate(BaseModel):
    name: str
    phone: str
    address: str
    has_bad_record: bool = False


class AdopterUpdate(BaseModel):
    name: Optional[str] = None
    phone: Optional[str] = None
    address: Optional[str] = None
    has_bad_record: Optional[bool] = None


class AdoptionApplication(BaseModel):
    id: str
    animal_id: str
    adopter_id: str
    status: AdoptionStatus
    queue_position: int = 0
    needs_extra_review: bool = False
    application_date: datetime = Field(default_factory=datetime.now)
    interview_date: Optional[datetime] = None
    adoption_date: Optional[date] = None
    notes: Optional[str] = None


class AdoptionApplicationCreate(BaseModel):
    animal_id: str
    adopter_id: str
    notes: Optional[str] = None


class AdoptionApplicationUpdate(BaseModel):
    status: Optional[AdoptionStatus] = None
    interview_date: Optional[datetime] = None
    adoption_date: Optional[date] = None
    notes: Optional[str] = None


class FollowUp(BaseModel):
    id: str
    adoption_id: str
    animal_id: str
    adopter_id: str
    scheduled_date: date
    days_after_adoption: int
    status: FollowUpStatus
    completed_date: Optional[date] = None
    notes: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class FollowUpUpdate(BaseModel):
    status: Optional[FollowUpStatus] = None
    completed_date: Optional[date] = None
    notes: Optional[str] = None


class Donation(BaseModel):
    id: str
    donor_name: str
    donor_phone: Optional[str] = None
    donation_type: DonationType
    amount: float = 0.0
    description: Optional[str] = None
    donation_date: datetime = Field(default_factory=datetime.now)
    created_at: datetime = Field(default_factory=datetime.now)

    @field_validator('amount')
    @classmethod
    def check_amount(cls, v: float) -> float:
        if v < 0:
            raise ValueError('捐赠金额不能为负数')
        return v


class DonationCreate(BaseModel):
    donor_name: str
    donor_phone: Optional[str] = None
    donation_type: DonationType
    amount: float = 0.0
    description: Optional[str] = None
    donation_date: Optional[datetime] = None

    @field_validator('amount')
    @classmethod
    def check_amount(cls, v: float) -> float:
        if v <= 0:
            raise ValueError('捐赠金额必须大于0')
        return v


class Appointment(BaseModel):
    id: str
    adoption_id: str
    animal_id: str
    adopter_id: str
    scheduled_time: datetime
    location: str
    status: AppointmentStatus
    notes: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class AppointmentCreate(BaseModel):
    adoption_id: str
    scheduled_time: datetime
    location: str
    notes: Optional[str] = None


class AppointmentUpdate(BaseModel):
    scheduled_time: Optional[datetime] = None
    location: Optional[str] = None
    status: Optional[AppointmentStatus] = None
    notes: Optional[str] = None


class DonationStats(BaseModel):
    total_amount: float = 0.0
    donation_count: int = 0
    last_donation_date: Optional[datetime] = None
