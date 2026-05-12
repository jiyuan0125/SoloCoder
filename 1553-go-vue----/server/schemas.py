from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional, List
from .models import (
    NavigationStatus, BeaconStatus, DredgingStatus,
    DraftDeclarationStatus, AlertSeverity, TodoPriority
)


class SectionBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    design_depth: float = Field(..., gt=0)


class SectionCreate(SectionBase):
    pass


class SectionUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=100)
    design_depth: Optional[float] = Field(None, gt=0)


class Section(SectionBase):
    id: int
    current_measured_depth: Optional[float]
    navigation_status: NavigationStatus
    beacon_status_normal: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class BeaconBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    code: str = Field(..., min_length=1, max_length=50)
    status: BeaconStatus = BeaconStatus.NORMAL
    description: Optional[str] = None


class BeaconCreate(BeaconBase):
    section_id: int


class BeaconUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=100)
    code: Optional[str] = Field(None, min_length=1, max_length=50)
    status: Optional[BeaconStatus] = None
    description: Optional[str] = None


class Beacon(BeaconBase):
    id: int
    section_id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class DepthRecordBase(BaseModel):
    measured_depth: float = Field(..., gt=0)
    recorded_at: Optional[datetime] = None


class DepthRecordCreate(DepthRecordBase):
    section_id: int


class DepthRecord(DepthRecordBase):
    id: int
    section_id: int
    recorded_at: datetime
    created_at: datetime

    class Config:
        from_attributes = True


class DredgingPlanBase(BaseModel):
    target_depth: float = Field(..., gt=0)
    status: DredgingStatus = DredgingStatus.PLANNED
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    description: Optional[str] = None


class DredgingPlanCreate(DredgingPlanBase):
    section_id: int


class DredgingPlanUpdate(BaseModel):
    target_depth: Optional[float] = Field(None, gt=0)
    status: Optional[DredgingStatus] = None
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    description: Optional[str] = None


class DredgingPlan(DredgingPlanBase):
    id: int
    section_id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class DraftDeclarationBase(BaseModel):
    vessel_name: str = Field(..., min_length=1, max_length=100)
    declared_draft: float = Field(..., gt=0)
    safety_margin: float = Field(0.3, ge=0)


class DraftDeclarationCreate(DraftDeclarationBase):
    section_id: int
    dredging_plan_id: Optional[int] = None


class DraftDeclarationUpdate(BaseModel):
    status: Optional[DraftDeclarationStatus] = None
    vessel_name: Optional[str] = Field(None, min_length=1, max_length=100)
    declared_draft: Optional[float] = Field(None, gt=0)
    safety_margin: Optional[float] = Field(None, ge=0)


class DraftDeclaration(DraftDeclarationBase):
    id: int
    section_id: int
    dredging_plan_id: Optional[int]
    status: DraftDeclarationStatus
    validation_result: Optional[str]
    approved_at: Optional[datetime]
    passed_at: Optional[datetime]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class AlertBase(BaseModel):
    type: str = Field(..., min_length=1, max_length=50)
    severity: AlertSeverity = AlertSeverity.MEDIUM
    message: str = Field(..., min_length=1)
    is_active: bool = True


class AlertCreate(AlertBase):
    section_id: int


class AlertUpdate(BaseModel):
    is_active: Optional[bool] = None
    severity: Optional[AlertSeverity] = None
    message: Optional[str] = None


class Alert(AlertBase):
    id: int
    section_id: int
    created_at: datetime
    resolved_at: Optional[datetime]

    class Config:
        from_attributes = True


class TodoBase(BaseModel):
    title: str = Field(..., min_length=1, max_length=200)
    description: Optional[str] = None
    priority: TodoPriority = TodoPriority.MEDIUM
    deadline: Optional[datetime] = None
    is_completed: bool = False


class TodoCreate(TodoBase):
    section_id: Optional[int] = None


class TodoUpdate(BaseModel):
    title: Optional[str] = Field(None, min_length=1, max_length=200)
    description: Optional[str] = None
    priority: Optional[TodoPriority] = None
    deadline: Optional[datetime] = None
    is_completed: Optional[bool] = None


class Todo(TodoBase):
    id: int
    section_id: Optional[int]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ValidationResult(BaseModel):
    valid: bool
    message: str
    max_allowed_draft: float
