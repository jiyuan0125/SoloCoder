from datetime import date, datetime
from typing import List, Optional
from pydantic import BaseModel, Field


class CrewBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    identity_number: str = Field(..., min_length=1, max_length=50)
    phone: Optional[str] = Field(None, max_length=20)
    email: Optional[str] = Field(None, max_length=100)
    position: Optional[str] = Field(None, max_length=50)
    is_intern: bool = False
    daily_rate: float = Field(..., gt=0)


class CrewCreate(CrewBase):
    pass


class CrewUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=100)
    phone: Optional[str] = Field(None, max_length=20)
    email: Optional[str] = Field(None, max_length=100)
    position: Optional[str] = Field(None, max_length=50)
    is_intern: Optional[bool] = None
    daily_rate: Optional[float] = Field(None, gt=0)
    status: Optional[str] = Field(None, max_length=20)


class CrewResponse(CrewBase):
    id: int
    status: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CertificateBase(BaseModel):
    certificate_type: str = Field(..., min_length=1, max_length=100)
    certificate_number: Optional[str] = Field(None, max_length=100)
    issue_date: date
    expiry_date: date


class CertificateCreate(CertificateBase):
    crew_id: int


class CertificateUpdate(BaseModel):
    certificate_type: Optional[str] = Field(None, min_length=1, max_length=100)
    certificate_number: Optional[str] = Field(None, max_length=100)
    issue_date: Optional[date] = None
    expiry_date: Optional[date] = None


class CertificateResponse(CertificateBase):
    id: int
    crew_id: int
    validity_days: Optional[int]
    status: str
    warning_level: Optional[str] = None
    days_to_expiry: Optional[int] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class TrainingRecordBase(BaseModel):
    training_type: str = Field(..., min_length=1, max_length=100)
    training_name: str = Field(..., min_length=1, max_length=200)
    training_date: date
    next_renewal_date: Optional[date] = None
    certificate_id: Optional[int] = None


class TrainingRecordCreate(TrainingRecordBase):
    crew_id: int


class TrainingRecordUpdate(BaseModel):
    training_type: Optional[str] = Field(None, min_length=1, max_length=100)
    training_name: Optional[str] = Field(None, min_length=1, max_length=200)
    training_date: Optional[date] = None
    next_renewal_date: Optional[date] = None
    certificate_id: Optional[int] = None


class TrainingRecordResponse(TrainingRecordBase):
    id: int
    crew_id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ShipBase(BaseModel):
    ship_name: str = Field(..., min_length=1, max_length=100)
    imo_number: Optional[str] = Field(None, max_length=50)
    vessel_type: Optional[str] = Field(None, max_length=50)
    total_crew_quota: int = Field(..., gt=0)


class ShipCreate(ShipBase):
    pass


class ShipUpdate(BaseModel):
    ship_name: Optional[str] = Field(None, min_length=1, max_length=100)
    imo_number: Optional[str] = Field(None, max_length=50)
    vessel_type: Optional[str] = Field(None, max_length=50)
    total_crew_quota: Optional[int] = Field(None, gt=0)
    status: Optional[str] = Field(None, max_length=20)


class ShipResponse(ShipBase):
    id: int
    current_crew_count: int
    available_slots: Optional[int] = None
    status: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AssignmentBase(BaseModel):
    ship_id: int
    position: str = Field(..., min_length=1, max_length=50)
    is_watch_keeper: bool = False
    embark_date: date


class AssignmentCreate(AssignmentBase):
    crew_id: int


class AssignmentUpdate(BaseModel):
    position: Optional[str] = Field(None, min_length=1, max_length=50)
    is_watch_keeper: Optional[bool] = None
    disembark_date: Optional[date] = None
    status: Optional[str] = Field(None, max_length=20)


class AssignmentResponse(AssignmentBase):
    id: int
    crew_id: int
    disembark_date: Optional[date] = None
    status: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AttendanceBase(BaseModel):
    attendance_date: date
    status: str = Field(..., min_length=1, max_length=20)
    is_half_day: bool = False
    remarks: Optional[str] = None


class AttendanceCreate(AttendanceBase):
    crew_id: int
    assignment_id: Optional[int] = None
    ship_id: Optional[int] = None


class AttendanceUpdate(BaseModel):
    status: Optional[str] = Field(None, min_length=1, max_length=20)
    is_half_day: Optional[bool] = None
    remarks: Optional[str] = None


class AttendanceResponse(AttendanceBase):
    id: int
    crew_id: int
    assignment_id: Optional[int]
    ship_id: Optional[int]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AttendanceSummary(BaseModel):
    month: str
    normal_days: float = 0.0
    overtime_days: float = 0.0
    sick_leave_days: float = 0.0
    personal_leave_days: float = 0.0
    total_days: float = 0.0


class SalaryBase(BaseModel):
    salary_month: str
    base_amount: float = 0.0


class SalaryCreate(SalaryBase):
    crew_id: int
    assignment_id: Optional[int] = None
    ship_id: Optional[int] = None


class SalaryResponse(BaseModel):
    id: int
    crew_id: int
    assignment_id: Optional[int]
    ship_id: Optional[int]
    salary_month: str
    base_amount: float
    normal_days: float
    overtime_days: float
    sick_leave_days: float
    personal_leave_days: float
    normal_amount: float
    overtime_amount: float
    sick_leave_amount: float
    total_amount: float
    status: str
    settlement_date: Optional[date]
    remarks: Optional[str]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CrewDetailResponse(CrewResponse):
    certificates: List[CertificateResponse] = []
    trainings: List[TrainingRecordResponse] = []
    assignments: List[AssignmentResponse] = []
    salaries: List[SalaryResponse] = []

    class Config:
        from_attributes = True


class ShipDetailResponse(ShipResponse):
    assignments: List[AssignmentResponse] = []
    crew_list: List[CrewResponse] = []
    certificate_summary: Optional[dict] = None
    attendance_summary: Optional[AttendanceSummary] = None

    class Config:
        from_attributes = True


class AssignmentValidation(BaseModel):
    valid: bool
    message: str
    issues: List[str] = []
