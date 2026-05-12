from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime, date
from app.models import (
    DeviceType, DeviceStatus, BridgeGrade, 
    InspectionType, InspectionStatus, ProblemLevel
)

class DeviceBase(BaseModel):
    device_code: str
    device_type: DeviceType
    name: str
    mileage: str
    bridge_length: Optional[float] = None
    bridge_grade: Optional[BridgeGrade] = None

class DeviceCreate(DeviceBase):
    pass

class DeviceUpdate(BaseModel):
    name: Optional[str] = None
    mileage: Optional[str] = None
    bridge_length: Optional[float] = None
    bridge_grade: Optional[BridgeGrade] = None

class DeviceStatusUpdate(BaseModel):
    status: DeviceStatus
    remark: Optional[str] = None

class DeviceResponse(DeviceBase):
    id: int
    status: DeviceStatus
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True

class DeviceStatusHistoryResponse(BaseModel):
    id: int
    old_status: DeviceStatus
    new_status: DeviceStatus
    change_time: datetime
    remark: Optional[str] = None

    class Config:
        from_attributes = True

class InspectorBase(BaseModel):
    name: str
    contact: Optional[str] = None

class InspectorCreate(InspectorBase):
    pass

class InspectorResponse(InspectorBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True

class InspectionPlanBase(BaseModel):
    device_id: int
    inspection_type: InspectionType
    frequency_per_month: int
    effective_month: date
    is_active: int = 1

class InspectionPlanCreate(InspectionPlanBase):
    pass

class InspectionPlanResponse(InspectionPlanBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True

class InspectionBase(BaseModel):
    device_id: int
    inspector_id: int
    inspection_type: InspectionType
    planned_start_time: datetime
    planned_end_time: datetime

class InspectionCreate(InspectionBase):
    plan_id: Optional[int] = None

class InspectionComplete(BaseModel):
    problem_level: ProblemLevel
    handling_suggestion: Optional[str] = None

class InspectionResponse(InspectionBase):
    id: int
    inspection_code: str
    plan_id: Optional[int] = None
    actual_start_time: Optional[datetime] = None
    actual_end_time: Optional[datetime] = None
    status: InspectionStatus
    problem_level: Optional[ProblemLevel] = None
    handling_suggestion: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True

class StatisticsResponse(BaseModel):
    completion_rate: float
    problems_by_level: dict
    device_status_distribution: dict
