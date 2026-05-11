from datetime import datetime, date
from typing import Optional, List
from enum import Enum

from pydantic import BaseModel, Field

from server.models import (
    BeaconStatus, DredgingStatus, CheckResult,
    TodoPriority, TodoStatus, WarningType, WarningStatus
)


class BeaconCreate(BaseModel):
    identifier: str = Field(..., min_length=1)
    name: Optional[str] = None
    status: BeaconStatus = BeaconStatus.NORMAL
    notes: Optional[str] = None


class BeaconUpdate(BaseModel):
    identifier: Optional[str] = None
    name: Optional[str] = None
    status: Optional[BeaconStatus] = None
    notes: Optional[str] = None


class BeaconOut(BaseModel):
    id: int
    section_id: int
    identifier: str
    name: Optional[str] = None
    status: BeaconStatus
    notes: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class DepthRecordCreate(BaseModel):
    measured_depth: float = Field(..., gt=0)
    record_date: Optional[date] = None
    notes: Optional[str] = None


class DepthRecordOut(BaseModel):
    id: int
    section_id: int
    measured_depth: float
    record_date: date
    notes: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class DredgingPlanCreate(BaseModel):
    target_depth: float = Field(..., gt=0)
    planned_start_date: Optional[date] = None
    planned_end_date: Optional[date] = None
    notes: Optional[str] = None


class DredgingPlanUpdate(BaseModel):
    target_depth: Optional[float] = Field(None, gt=0)
    status: Optional[DredgingStatus] = None
    planned_start_date: Optional[date] = None
    planned_end_date: Optional[date] = None
    actual_start_date: Optional[date] = None
    actual_end_date: Optional[date] = None
    notes: Optional[str] = None


class DredgingPlanOut(BaseModel):
    id: int
    section_id: int
    target_depth: float
    status: DredgingStatus
    planned_start_date: Optional[date] = None
    planned_end_date: Optional[date] = None
    actual_start_date: Optional[date] = None
    actual_end_date: Optional[date] = None
    notes: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class DraftDeclarationCreate(BaseModel):
    ship_name: str = Field(..., min_length=1)
    imo_number: Optional[str] = None
    declared_draft: float = Field(..., gt=0)


class DraftDeclarationUpdate(BaseModel):
    is_completed: Optional[bool] = None


class DraftDeclarationOut(BaseModel):
    id: int
    section_id: int
    ship_name: str
    imo_number: Optional[str] = None
    declared_draft: float
    check_result: CheckResult
    check_message: Optional[str] = None
    checked_at: Optional[datetime] = None
    is_completed: bool
    completed_at: Optional[datetime] = None
    effective_depth_at_check: Optional[float] = None
    created_at: datetime

    class Config:
        from_attributes = True


class SectionCreate(BaseModel):
    name: str = Field(..., min_length=1)
    design_depth: float = Field(..., gt=0)
    current_depth: Optional[float] = None


class SectionUpdate(BaseModel):
    name: Optional[str] = None
    design_depth: Optional[float] = Field(None, gt=0)


class SectionOut(BaseModel):
    id: int
    name: str
    design_depth: float
    current_depth: Optional[float] = None
    is_restricted: bool
    beacon_anomaly: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class TodoCreate(BaseModel):
    title: str = Field(..., min_length=1)
    description: Optional[str] = None
    priority: TodoPriority = TodoPriority.MEDIUM
    deadline: Optional[datetime] = None
    section_id: Optional[int] = None


class TodoUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    priority: Optional[TodoPriority] = None
    status: Optional[TodoStatus] = None
    deadline: Optional[datetime] = None


class TodoOut(BaseModel):
    id: int
    title: str
    description: Optional[str] = None
    priority: TodoPriority
    status: TodoStatus
    deadline: Optional[datetime] = None
    section_id: Optional[int] = None
    related_warning_id: Optional[int] = None
    completed_at: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class WarningOut(BaseModel):
    id: int
    section_id: int
    warning_type: WarningType
    message: str
    status: WarningStatus
    acknowledged_at: Optional[datetime] = None
    created_at: datetime

    class Config:
        from_attributes = True


class MessageResponse(BaseModel):
    message: str


class DepthCheckResult(BaseModel):
    is_restricted: bool
    effective_depth: Optional[float]
    threshold: float
