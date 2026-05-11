from datetime import datetime, date
from typing import Optional, List
from pydantic import BaseModel, Field
from .config import settings


class VesselBase(BaseModel):
    vessel_name: str = Field(..., min_length=1, max_length=100)
    vessel_number: Optional[str] = Field(None, max_length=50)
    vessel_type: Optional[str] = Field(None, max_length=50)
    gross_tonnage: Optional[float] = None
    length: Optional[float] = None
    owner_name: Optional[str] = Field(None, max_length=100)
    owner_id_card: Optional[str] = Field(None, max_length=50)
    registry_port: Optional[str] = Field(None, max_length=100)
    operating_company: Optional[str] = Field(None, max_length=100)


class VesselCreate(VesselBase):
    pass


class VesselUpdate(BaseModel):
    vessel_name: Optional[str] = Field(None, min_length=1, max_length=100)
    vessel_number: Optional[str] = Field(None, max_length=50)
    vessel_type: Optional[str] = Field(None, max_length=50)
    gross_tonnage: Optional[float] = None
    length: Optional[float] = None
    owner_name: Optional[str] = Field(None, max_length=100)
    owner_id_card: Optional[str] = Field(None, max_length=50)
    registry_port: Optional[str] = Field(None, max_length=100)
    operating_company: Optional[str] = Field(None, max_length=100)


class Vessel(VesselBase):
    id: int
    is_focus_attention: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class VesselDetail(Vessel):
    permits: List["Permit"] = []
    inspections: List["Inspection"] = []
    penalties: List["Penalty"] = []
    rectifications: List["Rectification"] = []


class PermitBase(BaseModel):
    vessel_id: Optional[int] = None
    permit_type: str
    applicant: str
    application_date: date
    expected_effective_date: date
    effective_date: Optional[date] = None
    expiry_date: Optional[date] = None
    remarks: Optional[str] = None


class PermitCreate(PermitBase):
    pass


class PermitUpdate(BaseModel):
    effective_date: Optional[date] = None
    expiry_date: Optional[date] = None
    status: Optional[str] = None
    reject_reason: Optional[str] = None
    remarks: Optional[str] = None


class Permit(PermitBase):
    id: int
    application_number: str
    status: str
    reject_reason: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class InspectionBase(BaseModel):
    vessel_id: Optional[int] = None
    inspector: str
    inspection_date: date
    inspection_location: Optional[str] = None
    inspection_items: Optional[str] = None
    result: str
    unqualified_items: Optional[str] = None
    remarks: Optional[str] = None


class InspectionCreate(InspectionBase):
    pass


class InspectionUpdate(BaseModel):
    inspector: Optional[str] = None
    inspection_date: Optional[date] = None
    inspection_location: Optional[str] = None
    inspection_items: Optional[str] = None
    result: Optional[str] = None
    unqualified_items: Optional[str] = None
    remarks: Optional[str] = None


class Inspection(InspectionBase):
    id: int
    inspection_number: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class PenaltyBase(BaseModel):
    vessel_id: Optional[int] = None
    inspection_id: Optional[int] = None
    violation_description: str
    fine_amount: float = Field(..., ge=settings.min_fine, le=settings.max_fine)
    issue_date: date
    deadline_date: date
    remarks: Optional[str] = None


class PenaltyCreate(PenaltyBase):
    pass


class PenaltyUpdate(BaseModel):
    payment_date: Optional[date] = None
    status: Optional[str] = None
    remarks: Optional[str] = None


class Penalty(PenaltyBase):
    id: int
    penalty_number: str
    payment_date: Optional[date] = None
    status: str
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class RectificationBase(BaseModel):
    vessel_id: Optional[int] = None
    inspection_id: Optional[int] = None
    rectification_content: str
    rectification_deadline: Optional[date] = None
    remarks: Optional[str] = None


class RectificationCreate(RectificationBase):
    pass


class RectificationSubmit(BaseModel):
    rectification_measure: str
    rectification_date: Optional[date] = None


class RectificationRecheck(BaseModel):
    recheck_date: date
    recheck_result: str


class RectificationUpdate(BaseModel):
    rectification_deadline: Optional[date] = None
    remarks: Optional[str] = None


class Rectification(RectificationBase):
    id: int
    rectification_number: str
    created_date: date
    rectification_measure: Optional[str] = None
    rectification_date: Optional[date] = None
    recheck_date: Optional[date] = None
    recheck_result: Optional[str] = None
    status: str
    recheck_count: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AuditLogBase(BaseModel):
    action: str
    model_name: str
    record_id: Optional[int] = None
    detail: Optional[str] = None


class AuditLog(AuditLogBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


VesselDetail.model_rebuild()
