from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field, EmailStr
from app.models.training import TrainingStatus, RegistrationStatus


class TrainingBase(BaseModel):
    name: str = Field(..., max_length=200)
    description: Optional[str] = None
    instructor: Optional[str] = Field(None, max_length=100)
    max_participants: int = Field(..., ge=1)
    venue_id: int
    start_date: datetime
    end_date: datetime
    schedule: Optional[str] = Field(None, max_length=200)
    status: TrainingStatus = TrainingStatus.REGISTRATION


class TrainingCreate(TrainingBase):
    pass


class TrainingUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    instructor: Optional[str] = None
    max_participants: Optional[int] = None
    venue_id: Optional[int] = None
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    schedule: Optional[str] = None
    status: Optional[TrainingStatus] = None


class TrainingOut(TrainingBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RegistrationBase(BaseModel):
    student_name: str = Field(..., max_length=100)
    student_phone: Optional[str] = Field(None, max_length=20)
    student_email: Optional[EmailStr] = None


class RegistrationCreate(RegistrationBase):
    training_id: int


class RegistrationOut(RegistrationBase):
    id: int
    training_id: int
    status: RegistrationStatus
    waitlist_order: Optional[int]
    consecutive_absences: int
    registered_at: datetime
    confirmed_at: Optional[datetime]
    cancelled_at: Optional[datetime]

    class Config:
        from_attributes = True


class AttendanceBase(BaseModel):
    registration_id: int
    session_date: datetime
    is_present: int = Field(..., ge=0, le=1)


class AttendanceCreate(AttendanceBase):
    pass


class AttendanceOut(AttendanceBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class NotificationResponse(BaseModel):
    notified: List[int]
    skipped: List[int]
    log_entry: Optional[str]
