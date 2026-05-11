from datetime import date, datetime
from typing import List, Optional
from pydantic import BaseModel, Field
from .models import (
    FishermanStatus, LicenseStatus, LicenseType, ViolationType,
    Severity, PenaltyType, CaseStatus, TodoType, TodoStatus
)


class FishermanBase(BaseModel):
    name: str
    id_card: str
    phone: Optional[str] = None
    address: Optional[str] = None
    vessel_name: Optional[str] = None
    vessel_registration: Optional[str] = None


class FishermanCreate(FishermanBase):
    pass


class FishermanUpdate(BaseModel):
    name: Optional[str] = None
    phone: Optional[str] = None
    address: Optional[str] = None
    vessel_name: Optional[str] = None
    vessel_registration: Optional[str] = None
    status: Optional[FishermanStatus] = None


class FishermanResponse(FishermanBase):
    id: int
    status: FishermanStatus
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class FishingLicenseBase(BaseModel):
    fisherman_id: int
    license_type: LicenseType
    valid_from: date
    valid_to: date
    fishing_area: Optional[str] = None
    allowed_gear: Optional[str] = None


class FishingLicenseCreate(FishingLicenseBase):
    pass


class FishingLicenseUpdate(BaseModel):
    valid_from: Optional[date] = None
    valid_to: Optional[date] = None
    fishing_area: Optional[str] = None
    allowed_gear: Optional[str] = None
    status: Optional[LicenseStatus] = None


class FishingLicenseResponse(FishingLicenseBase):
    id: int
    license_number: Optional[str] = None
    status: LicenseStatus
    application_date: date
    approval_date: Optional[date] = None
    rejection_reason: Optional[str] = None
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class EvidenceBase(BaseModel):
    evidence_type: Optional[str] = None
    description: Optional[str] = None
    file_path: Optional[str] = None


class EvidenceCreate(EvidenceBase):
    case_id: int


class EvidenceResponse(EvidenceBase):
    id: int
    submitted_at: datetime

    class Config:
        from_attributes = True


class AppealBase(BaseModel):
    case_id: int
    appeal_reason: str


class AppealCreate(AppealBase):
    pass


class AppealResponse(AppealBase):
    id: int
    appeal_date: date
    decision: Optional[str] = None
    decision_date: Optional[date] = None
    decision_reason: Optional[str] = None

    class Config:
        from_attributes = True


class ViolationCaseBase(BaseModel):
    fisherman_id: int
    license_id: Optional[int] = None
    violation_type: ViolationType
    severity: Severity
    description: Optional[str] = None
    location: Optional[str] = None
    violation_date: date


class ViolationCaseCreate(ViolationCaseBase):
    case_number: str


class ViolationCaseDecision(BaseModel):
    penalty_type: PenaltyType
    fine_amount: Optional[float] = None
    suspension_days: Optional[int] = None
    decision_reason: Optional[str] = None


class ViolationCaseResponse(ViolationCaseBase):
    id: int
    case_number: str
    status: CaseStatus
    penalty_type: Optional[PenaltyType] = None
    fine_amount: Optional[float] = None
    suspension_days: Optional[int] = None
    decision_reason: Optional[str] = None
    decision_date: Optional[date] = None
    appeal_deadline: Optional[date] = None
    is_closed: bool
    closed_date: Optional[date] = None
    evidences: List[EvidenceResponse] = []
    appeals: List[AppealResponse] = []
    created_at: datetime
    updated_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class TodoBase(BaseModel):
    title: str
    description: Optional[str] = None
    todo_type: TodoType
    related_fisherman_id: Optional[int] = None
    related_case_id: Optional[int] = None
    due_date: Optional[date] = None
    location: Optional[str] = None


class TodoCreate(TodoBase):
    pass


class TodoUpdate(BaseModel):
    status: Optional[TodoStatus] = None
    description: Optional[str] = None


class TodoResponse(TodoBase):
    id: int
    status: TodoStatus
    created_at: datetime
    completed_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class SeedFarmBase(BaseModel):
    name: str
    address: Optional[str] = None
    contact_person: Optional[str] = None
    phone: Optional[str] = None
    license_number: Optional[str] = None


class SeedFarmCreate(SeedFarmBase):
    pass


class SeedFarmUpdate(BaseModel):
    name: Optional[str] = None
    address: Optional[str] = None
    contact_person: Optional[str] = None
    phone: Optional[str] = None
    license_number: Optional[str] = None


class SeedFarmResponse(SeedFarmBase):
    id: int

    class Config:
        from_attributes = True


class ReleaseActivityBase(BaseModel):
    activity_code: str
    activity_name: str
    seed_farm_id: Optional[int] = None
    species: str
    quantity: int
    unit: str = "tail"
    release_date: date
    release_location: str
    water_area: Optional[str] = None
    coordinator: Optional[str] = None
    remarks: Optional[str] = None


class ReleaseActivityCreate(ReleaseActivityBase):
    pass


class ReleaseActivityResponse(ReleaseActivityBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class PortBase(BaseModel):
    name: str
    location: Optional[str] = None
    berth_count: int
    max_tonnage: float
    manager: Optional[str] = None
    phone: Optional[str] = None


class PortCreate(PortBase):
    pass


class PortUpdate(BaseModel):
    name: Optional[str] = None
    location: Optional[str] = None
    berth_count: Optional[int] = None
    max_tonnage: Optional[float] = None
    manager: Optional[str] = None
    phone: Optional[str] = None


class PortResponse(PortBase):
    id: int

    class Config:
        from_attributes = True


class PortRecordBase(BaseModel):
    port_id: int
    fisherman_id: int
    record_type: str
    vessel_name: Optional[str] = None
    vessel_registration: Optional[str] = None
    destination: Optional[str] = None
    cargo: Optional[str] = None
    remarks: Optional[str] = None


class PortRecordCreate(PortRecordBase):
    pass


class PortRecordResponse(PortRecordBase):
    id: int
    record_time: datetime

    class Config:
        from_attributes = True


class MessageResponse(BaseModel):
    message: str
