from __future__ import annotations

from datetime import date, datetime
from typing import List, Optional
from uuid import UUID

from pydantic import BaseModel, Field

from .models import (
    AcceptanceStatus,
    InspectionStatus,
    MaterialStatus,
    SafetyLevel,
)


class TailingPondBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    capacity: float = Field(..., gt=0)
    dam_height: float = Field(..., gt=0)
    safety_level: SafetyLevel


class TailingPondCreate(TailingPondBase):
    pass


class TailingPongUpdate(BaseModel):
    name: Optional[str] = None
    capacity: Optional[float] = Field(None, gt=0)
    dam_height: Optional[float] = Field(None, gt=0)
    safety_level: Optional[SafetyLevel] = None


class TailingPondResponse(TailingPondBase):
    id: UUID
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class MonitoringSectionBase(BaseModel):
    pond_id: UUID
    name: str = Field(..., min_length=1, max_length=100)


class MonitoringSectionCreate(MonitoringSectionBase):
    pass


class MonitoringSectionResponse(MonitoringSectionBase):
    id: UUID
    created_at: datetime

    class Config:
        from_attributes = True


class MonitoringDataBase(BaseModel):
    section_id: UUID
    timestamp: datetime
    dry_beach_length: Optional[float] = None
    phreatic_line: Optional[float] = None
    dam_displacement: Optional[float] = None
    water_level: Optional[float] = None


class MonitoringDataCreate(MonitoringDataBase):
    pass


class MonitoringDataResponse(MonitoringDataBase):
    id: UUID

    class Config:
        from_attributes = True


class InspectionTaskBase(BaseModel):
    pond_id: UUID
    planned_date: date


class InspectionTaskCreate(InspectionTaskBase):
    pass


class InspectionTaskUpdate(BaseModel):
    actual_date: Optional[date] = None
    inspector: Optional[str] = None
    remarks: Optional[str] = None


class InspectionTaskResponse(InspectionTaskBase):
    id: UUID
    actual_date: Optional[date] = None
    inspector: Optional[str] = None
    remarks: Optional[str] = None
    status: InspectionStatus
    created_at: datetime

    class Config:
        from_attributes = True


class ReviewRecordResponse(BaseModel):
    id: UUID
    reviewer: str
    opinion: str
    is_approved: bool
    review_date: datetime

    class Config:
        from_attributes = True


class ReviewCreate(BaseModel):
    reviewer: str = Field(..., min_length=1, max_length=100)
    opinion: str = Field(..., min_length=1)
    is_approved: bool


class AcceptanceMaterialBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=200)


class AcceptanceMaterialCreate(AcceptanceMaterialBase):
    pass


class AcceptanceMaterialSubmit(BaseModel):
    content: str = Field(..., min_length=1)


class AcceptanceMaterialResponse(AcceptanceMaterialBase):
    id: UUID
    status: MaterialStatus
    content: Optional[str] = None
    submission_date: Optional[datetime] = None
    review_records: List[ReviewRecordResponse] = []

    class Config:
        from_attributes = True


class ClosureAcceptanceBase(BaseModel):
    pond_id: UUID


class ClosureAcceptanceCreate(ClosureAcceptanceBase):
    pass


class ClosureAcceptanceResponse(ClosureAcceptanceBase):
    id: UUID
    status: AcceptanceStatus
    materials: List[AcceptanceMaterialResponse] = []
    created_at: datetime
    approval_date: Optional[datetime] = None

    class Config:
        from_attributes = True


class WarningResponse(BaseModel):
    id: UUID
    pond_id: UUID
    section_id: UUID
    warning_type: str
    message: str
    value: float
    threshold: float
    timestamp: datetime
    is_acknowledged: bool

    class Config:
        from_attributes = True


class AggregatedDataResponse(BaseModel):
    id: UUID
    section_id: UUID
    data_type: str
    period_start: datetime
    period_end: datetime
    value: float
    created_at: datetime

    class Config:
        from_attributes = True
