from datetime import datetime
from typing import Optional
from pydantic import BaseModel, Field


class EducationEventBase(BaseModel):
    name: str = Field(..., max_length=200)
    description: Optional[str] = None
    venue_id: int
    start_time: datetime
    end_time: datetime
    min_age_months: int = Field(..., ge=0)
    max_age_months: int = Field(..., ge=0)
    max_participants: int = Field(..., ge=1)


class EducationEventCreate(EducationEventBase):
    pass


class EducationEventUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    venue_id: Optional[int] = None
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    min_age_months: Optional[int] = None
    max_age_months: Optional[int] = None
    max_participants: Optional[int] = None


class EducationEventOut(EducationEventBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class EducationParticipantBase(BaseModel):
    participant_name: str = Field(..., max_length=100)
    birth_date: datetime
    guardian_name: Optional[str] = Field(None, max_length=100)
    guardian_phone: Optional[str] = Field(None, max_length=20)


class EducationParticipantCreate(EducationParticipantBase):
    event_id: int


class EducationParticipantOut(EducationParticipantBase):
    id: int
    event_id: int
    registered_at: datetime

    class Config:
        from_attributes = True


class AgeCheckResponse(BaseModel):
    participant_id: int
    participant_name: str
    age_months: int
    is_eligible: bool
    min_age_months: int
    max_age_months: int
    message: str
