from datetime import date, datetime
from typing import List, Optional
from pydantic import BaseModel, Field


class QualificationBase(BaseModel):
    qualification_type: str
    aircraft_type: Optional[str] = None
    certificate_number: Optional[str] = None
    issue_date: date
    expiry_date: date
    is_mandatory: bool = False


class QualificationCreate(QualificationBase):
    pass


class QualificationUpdate(BaseModel):
    qualification_type: Optional[str] = None
    aircraft_type: Optional[str] = None
    certificate_number: Optional[str] = None
    issue_date: Optional[date] = None
    expiry_date: Optional[date] = None
    is_mandatory: Optional[bool] = None
    is_valid: Optional[bool] = None


class Qualification(QualificationBase):
    id: int
    pilot_id: int
    last_used_date: Optional[date] = None
    is_valid: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class MedicalBase(BaseModel):
    examination_date: date
    expiry_date: date
    result: str = "qualified"
    medical_type: str = "class1"
    notes: Optional[str] = None


class MedicalCreate(MedicalBase):
    pass


class MedicalUpdate(BaseModel):
    examination_date: Optional[date] = None
    expiry_date: Optional[date] = None
    result: Optional[str] = None
    medical_type: Optional[str] = None
    notes: Optional[str] = None


class Medical(MedicalBase):
    id: int
    pilot_id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class PilotBase(BaseModel):
    employee_id: str
    name: str
    gender: Optional[str] = None
    birth_date: date
    english_level: int = Field(default=1, ge=1, le=10)


class PilotCreate(PilotBase):
    pass


class PilotUpdate(BaseModel):
    name: Optional[str] = None
    gender: Optional[str] = None
    birth_date: Optional[date] = None
    english_level: Optional[int] = Field(default=None, ge=1, le=10)
    status: Optional[str] = None


class Pilot(PilotBase):
    id: int
    status: str
    monthly_hours: float
    yearly_hours: float
    priority_score: int
    created_at: datetime
    updated_at: datetime
    qualifications: List[Qualification] = []
    medicals: List[Medical] = []

    class Config:
        from_attributes = True


class PilotSummary(BaseModel):
    id: int
    employee_id: str
    name: str
    english_level: int
    status: str
    monthly_hours: float
    yearly_hours: float
    priority_score: int

    class Config:
        from_attributes = True


class FlightScheduleBase(BaseModel):
    flight_number: str
    departure_airport: str
    arrival_airport: str
    aircraft_type: str
    is_international: bool = False
    departure_time: datetime
    arrival_time: datetime


class FlightScheduleCreate(FlightScheduleBase):
    pilot_id: int


class FlightScheduleUpdate(BaseModel):
    flight_number: Optional[str] = None
    departure_airport: Optional[str] = None
    arrival_airport: Optional[str] = None
    aircraft_type: Optional[str] = None
    is_international: Optional[bool] = None
    departure_time: Optional[datetime] = None
    arrival_time: Optional[datetime] = None
    status: Optional[str] = None


class FlightSchedule(FlightScheduleBase):
    id: int
    pilot_id: int
    flight_duration: float
    status: str
    validation_result: Optional[str] = None
    validation_passed: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ValidationResult(BaseModel):
    valid: bool
    messages: List[str] = []
    warnings: List[str] = []


class PilotRecommendation(BaseModel):
    pilot: PilotSummary
    score: float
    warnings: List[str] = []
    reasons: List[str] = []
