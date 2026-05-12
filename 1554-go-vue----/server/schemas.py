from datetime import datetime, date
from typing import Optional, List
from pydantic import BaseModel, Field, field_validator
from server.models import (
    LicenseType, LicenseStatus, InspectionResult,
    RectificationStatus, PenaltyStatus, ActionType
)


class ShipBase(BaseModel):
    imo_number: Optional[str] = Field(None, max_length=20)
    name: Optional[str] = Field(None, max_length=100)
    registration_port: Optional[str] = Field(None, max_length=100)
    gross_tonnage: Optional[float] = None
    length: Optional[float] = None
    width: Optional[float] = None
    owner_name: Optional[str] = Field(None, max_length=100)


class ShipCreate(ShipBase):
    pass


class ShipUpdate(ShipBase):
    pass


class ShipResponse(ShipBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class LicenseBase(BaseModel):
    license_type: LicenseType
    application_date: date
    expected_effective_date: date
    applicant: str = Field(..., max_length=100)


class LicenseCreate(LicenseBase):
    ship_id: Optional[int] = None

    @field_validator('expected_effective_date')
    def check_effective_date(cls, v, info):
        from server.services import add_working_days
        application_date = info.data.get('application_date')
        if application_date and v < add_working_days(application_date, 3):
            raise ValueError('期望生效日期不能早于申请日期加3个工作日')
        return v


class LicenseUpdate(BaseModel):
    status: Optional[LicenseStatus] = None
    rejection_reason: Optional[str] = None
    effective_date: Optional[date] = None
    expiration_date: Optional[date] = None


class LicenseResponse(LicenseBase):
    id: int
    ship_id: Optional[int]
    status: LicenseStatus
    rejection_reason: Optional[str]
    effective_date: Optional[date]
    expiration_date: Optional[date]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class InspectionBase(BaseModel):
    inspection_date: date
    inspector: str = Field(..., max_length=100)
    location: Optional[str] = Field(None, max_length=200)
    result: InspectionResult
    findings: Optional[str] = None
    remarks: Optional[str] = None


class InspectionCreate(InspectionBase):
    ship_id: Optional[int] = None
    ship_temp_identifier: Optional[str] = Field(None, max_length=100)


class InspectionUpdate(BaseModel):
    result: Optional[InspectionResult] = None
    findings: Optional[str] = None
    remarks: Optional[str] = None


class InspectionResponse(InspectionBase):
    id: int
    ship_id: Optional[int]
    ship_temp_identifier: Optional[str]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RectificationBase(BaseModel):
    requirement: str
    deadline: date


class RectificationCreate(RectificationBase):
    ship_id: Optional[int] = None
    ship_temp_identifier: Optional[str] = Field(None, max_length=100)
    inspection_id: Optional[int] = None


class RectificationUpdate(BaseModel):
    status: Optional[RectificationStatus] = None
    rectification_details: Optional[str] = None
    rectification_date: Optional[date] = None
    recheck_result: Optional[str] = None
    recheck_date: Optional[date] = None


class RectificationResponse(RectificationBase):
    id: int
    ship_id: Optional[int]
    ship_temp_identifier: Optional[str]
    inspection_id: Optional[int]
    status: RectificationStatus
    rectification_details: Optional[str]
    rectification_date: Optional[date]
    recheck_result: Optional[str]
    recheck_date: Optional[date]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class PenaltyBase(BaseModel):
    violation: str
    fine_amount: float
    issue_date: date
    due_date: date
    remarks: Optional[str] = None

    @field_validator('fine_amount')
    def check_fine_amount(cls, v):
        if v < 1000 or v > 50000:
            raise ValueError('罚款金额必须在1000到50000元之间')
        if v % 100 != 0:
            raise ValueError('罚款金额必须是100的整数倍')
        return v


class PenaltyCreate(PenaltyBase):
    ship_id: Optional[int] = None
    ship_temp_identifier: Optional[str] = Field(None, max_length=100)
    inspection_id: Optional[int] = None


class PenaltyUpdate(BaseModel):
    payment_date: Optional[date] = None
    status: Optional[PenaltyStatus] = None
    remarks: Optional[str] = None


class PenaltyResponse(PenaltyBase):
    id: int
    ship_id: Optional[int]
    ship_temp_identifier: Optional[str]
    inspection_id: Optional[int]
    payment_date: Optional[date]
    status: PenaltyStatus
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AuditLogResponse(BaseModel):
    id: int
    entity_type: str
    entity_id: int
    action: ActionType
    details: Optional[str]
    created_at: datetime

    class Config:
        from_attributes = True


class ShipDetailResponse(ShipResponse):
    licenses: List[LicenseResponse] = []
    inspections: List[InspectionResponse] = []
    penalties: List[PenaltyResponse] = []
    rectifications: List[RectificationResponse] = []
    is_high_priority: bool = False
