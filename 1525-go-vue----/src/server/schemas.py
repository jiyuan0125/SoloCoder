from datetime import date
from typing import Optional, List
from pydantic import BaseModel, Field, field_validator

from src.core.models import (
    CertificateType,
    CertificateStatus,
    TodoType,
    TodoStatus,
)


class CertificateCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=200)
    type: CertificateType
    number: str = Field(..., min_length=1, max_length=100)
    issuing_date: date
    expiry_date: date
    remarks: Optional[str] = None

    @field_validator("expiry_date")
    @classmethod
    def check_expiry_after_issuing(cls, v: date, info) -> date:
        issuing_date = info.data.get("issuing_date")
        if issuing_date and v < issuing_date:
            raise ValueError("有效期不能早于发证日期")
        return v


class CertificateUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=200)
    number: Optional[str] = Field(None, min_length=1, max_length=100)
    issuing_date: Optional[date] = None
    expiry_date: Optional[date] = None
    remarks: Optional[str] = None


class CertificateResponse(BaseModel):
    id: int
    name: str
    type: CertificateType
    number: str
    issuing_date: date
    expiry_date: date
    is_cancelled: bool
    remarks: Optional[str]
    status: CertificateStatus

    class Config:
        from_attributes = True


class AnnualInspectionCreate(BaseModel):
    certificate_id: int
    year: int = Field(..., ge=2000, le=2100)
    inspection_date: date
    result: bool
    remarks: Optional[str] = None


class AnnualInspectionResponse(BaseModel):
    id: int
    certificate_id: int
    year: int
    inspection_date: date
    result: bool
    remarks: Optional[str]

    class Config:
        from_attributes = True


class ComplianceCheckCreate(BaseModel):
    certificate_id: int
    check_date: date
    check_items: str
    is_compliant: bool
    has_safety_issues: bool = False
    remarks: Optional[str] = None


class ComplianceCheckResponse(BaseModel):
    id: int
    certificate_id: int
    check_date: date
    check_items: str
    is_compliant: bool
    has_safety_issues: bool
    remarks: Optional[str]

    class Config:
        from_attributes = True


class TodoUpdate(BaseModel):
    status: TodoStatus


class TodoResponse(BaseModel):
    id: int
    type: TodoType
    certificate_id: Optional[int]
    compliance_check_id: Optional[int]
    due_date: date
    status: TodoStatus
    description: str
    created_at: date

    class Config:
        from_attributes = True