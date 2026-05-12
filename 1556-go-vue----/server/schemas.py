from datetime import date, datetime
from typing import Optional, List
from pydantic import BaseModel, Field

from .models import CrewStatus, AttendanceType, CertificateStatus


class CrewBase(BaseModel):
    name: str
    id_number: str
    phone: Optional[str] = None
    email: Optional[str] = None
    is_intern: bool = False


class CrewCreate(CrewBase):
    pass


class CrewUpdate(BaseModel):
    name: Optional[str] = None
    phone: Optional[str] = None
    email: Optional[str] = None
    is_intern: Optional[bool] = None
    status: Optional[str] = None


class CrewResponse(CrewBase):
    id: int
    status: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CrewDetailResponse(CrewResponse):
    certificates: List["CertificateResponse"] = []
    training_records: List["TrainingRecordResponse"] = []
    assignments: List["AssignmentResponse"] = []
    attendance_summary: Optional["MonthlyAttendanceSummary"] = None


class CertificateBase(BaseModel):
    certificate_type: str
    certificate_number: str
    issue_date: date
    expiry_date: date


class CertificateCreate(CertificateBase):
    crew_id: int


class CertificateUpdate(BaseModel):
    certificate_type: Optional[str] = None
    issue_date: Optional[date] = None
    expiry_date: Optional[date] = None


class CertificateResponse(CertificateBase):
    id: int
    crew_id: int
    status: str
    days_until_expiry: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CertificateWithStatusResponse(CertificateResponse):
    status: str
    warning_level: Optional[str] = None


class TrainingRecordBase(BaseModel):
    training_type: str
    training_date: date
    description: Optional[str] = None


class TrainingRecordCreate(TrainingRecordBase):
    crew_id: int


class TrainingRecordResponse(TrainingRecordBase):
    id: int
    crew_id: int
    created_at: datetime

    class Config:
        from_attributes = True


class ShipBase(BaseModel):
    name: str
    imo_number: Optional[str] = None
    capacity: int
    description: Optional[str] = None


class ShipCreate(ShipBase):
    pass


class ShipUpdate(BaseModel):
    name: Optional[str] = None
    imo_number: Optional[str] = None
    capacity: Optional[int] = None
    description: Optional[str] = None


class ShipResponse(ShipBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AssignmentBase(BaseModel):
    position: str
    is_watch_keeper: bool = False
    start_date: date
    end_date: Optional[date] = None


class AssignmentCreate(AssignmentBase):
    crew_id: int
    ship_id: int


class AssignmentUpdate(BaseModel):
    position: Optional[str] = None
    is_watch_keeper: Optional[bool] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None
    is_active: Optional[bool] = None


class AssignmentResponse(AssignmentBase):
    id: int
    crew_id: int
    ship_id: int
    is_active: bool
    created_at: datetime
    updated_at: datetime
    ship: Optional[ShipResponse] = None
    crew: Optional[CrewResponse] = None

    class Config:
        from_attributes = True


class AttendanceRecordBase(BaseModel):
    attendance_date: date
    attendance_type: str
    is_half_day: bool = False


class AttendanceRecordCreate(AttendanceRecordBase):
    crew_id: int
    assignment_id: Optional[int] = None
    ship_id: Optional[int] = None


class AttendanceRecordResponse(AttendanceRecordBase):
    id: int
    crew_id: int
    assignment_id: Optional[int] = None
    ship_id: Optional[int] = None
    created_at: datetime

    class Config:
        from_attributes = True


class MonthlyAttendanceSummary(BaseModel):
    year: int
    month: int
    normal_days: float = 0.0
    overtime_days: float = 0.0
    sick_leave_days: float = 0.0
    personal_leave_days: float = 0.0
    total_days: float = 0.0


class SalaryRecordBase(BaseModel):
    year: int
    month: int
    base_salary: float = 0.0


class SalaryRecordCreate(SalaryRecordBase):
    crew_id: int
    ship_id: Optional[int] = None
    assignment_id: Optional[int] = None


class SalaryRecordResponse(SalaryRecordBase):
    id: int
    crew_id: int
    ship_id: Optional[int] = None
    assignment_id: Optional[int] = None
    normal_days: float = 0.0
    overtime_days: float = 0.0
    sick_leave_days: float = 0.0
    personal_leave_days: float = 0.0
    total_amount: float = 0.0
    is_settled: bool = False
    settlement_date: Optional[date] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class CrewCertificateStatus(BaseModel):
    crew_id: int
    crew_name: str
    has_expired_certificates: bool
    certificate_summary: List[dict] = []


class ShipDetailResponse(ShipResponse):
    crew_list: List[AssignmentResponse] = []
    attendance_summary: List[MonthlyAttendanceSummary] = []
    certificate_statuses: List[CrewCertificateStatus] = []


class CrewAssignmentHistory(BaseModel):
    assignment: AssignmentResponse
    attendance_records: List[AttendanceRecordResponse] = []
    salary_records: List[SalaryRecordResponse] = []


class SalarySettlementRequest(BaseModel):
    crew_id: int
    end_date: date


CrewDetailResponse.model_rebuild()
