from pydantic import BaseModel, Field
from datetime import datetime, date
from typing import Optional, List
from enum import Enum

from .models import (
    LicenseStatus, PenaltyType, PenaltyStatus, RiverImportance
)


class LicenseBase(BaseModel):
    company_name: str
    company_identifier: str
    river_section: str
    mining_location: Optional[str] = None
    annual_quota: float
    start_date: date
    end_date: date


class LicenseCreate(LicenseBase):
    pass


class LicenseUpdate(BaseModel):
    company_name: Optional[str] = None
    mining_location: Optional[str] = None
    annual_quota: Optional[float] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None


class LicenseResponse(LicenseBase):
    id: int
    license_number: Optional[str] = None
    status: LicenseStatus
    submit_time: Optional[datetime] = None
    accept_time: Optional[datetime] = None
    inspect_time: Optional[datetime] = None
    publish_time: Optional[datetime] = None
    issue_time: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class LicenseListResponse(BaseModel):
    total: int
    items: List[LicenseResponse]


class InspectRequest(BaseModel):
    inspect_result: str
    passed: bool = True


class PublishRequest(BaseModel):
    publish_days: int = 7


class MiningReportBase(BaseModel):
    report_date: date
    start_time: datetime
    end_time: datetime
    mining_location: Optional[str] = None
    expected_volume: Optional[float] = None
    operator_name: Optional[str] = None


class MiningReportCreate(MiningReportBase):
    license_id: int


class MiningReportResponse(MiningReportBase):
    id: int
    license_id: int
    status: str
    created_at: datetime

    class Config:
        from_attributes = True


class WeighbridgeRecordBase(BaseModel):
    license_id: int
    record_time: datetime
    vehicle_plate: Optional[str] = None
    gross_weight: float
    tare_weight: float
    net_weight: float
    material_type: Optional[str] = None


class WeighbridgeRecordCreate(WeighbridgeRecordBase):
    pass


class WeighbridgeRecordResponse(WeighbridgeRecordBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class LawEnforcementBase(BaseModel):
    company_name: str
    violation_time: datetime
    violation_location: str
    violation_description: str
    penalty_type: PenaltyType
    penalty_amount: Optional[float] = None


class LawEnforcementCreate(LawEnforcementBase):
    license_id: Optional[int] = None


class LawEnforcementResponse(LawEnforcementBase):
    id: int
    license_id: Optional[int] = None
    penalty_status: PenaltyStatus
    penalty_time: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class RiverSectionBase(BaseModel):
    name: str
    importance: RiverImportance = RiverImportance.NORMAL
    description: Optional[str] = None


class RiverSectionCreate(RiverSectionBase):
    pass


class RiverSectionResponse(RiverSectionBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class PatrolTaskBase(BaseModel):
    river_section_id: int
    task_date: date
    assigned_to: Optional[str] = None


class PatrolTaskCreate(PatrolTaskBase):
    pass


class PatrolTaskUpdate(BaseModel):
    status: Optional[str] = None
    result: Optional[str] = None


class PatrolTaskResponse(PatrolTaskBase):
    id: int
    status: str
    result: Optional[str] = None
    completed_at: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class MonthlyReportResponse(BaseModel):
    id: int
    report_year: int
    report_month: int
    content: str
    status: str
    reviewer: Optional[str] = None
    review_time: Optional[datetime] = None
    review_comment: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class MiningStatisticsResponse(BaseModel):
    license_id: int
    year: int
    month: int
    total_volume: float
    warning_sent: bool


class WarningInfo(BaseModel):
    license_id: int
    license_number: Optional[str]
    company_name: str
    annual_quota: float
    current_volume: float
    percentage: float
    message: str


class AuditLogResponse(BaseModel):
    id: int
    action: str
    entity_type: str
    entity_id: int
    details: Optional[str] = None
    operator: str
    created_at: datetime

    class Config:
        from_attributes = True
