from datetime import datetime
from typing import Optional
from pydantic import BaseModel
from .models import (
    MaintenanceGrade, FacilityType, FacilityStatus, 
    ActivityStatus, TaskStatus, Season
)


class ParkBase(BaseModel):
    park_code: str
    name: str
    location: Optional[str] = None
    area: Optional[float] = None


class ParkCreate(ParkBase):
    pass


class Park(ParkBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class GreenZoneBase(BaseModel):
    zone_code: str
    name: str
    area: Optional[float] = None
    plant_types: Optional[str] = None
    location_description: Optional[str] = None


class GreenZoneCreate(GreenZoneBase):
    park_code: str


class GreenZone(GreenZoneBase):
    id: int
    park_id: int
    current_grade: MaintenanceGrade
    consecutive_missed_tasks: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class GradeAdjustmentCreate(BaseModel):
    new_grade: MaintenanceGrade
    reason: Optional[str] = None
    season: Optional[Season] = None
    adjusted_by: Optional[str] = None


class GradeAdjustment(BaseModel):
    id: int
    green_zone_id: int
    old_grade: MaintenanceGrade
    new_grade: MaintenanceGrade
    reason: Optional[str]
    season: Optional[Season]
    adjusted_by: Optional[str]
    adjusted_at: datetime

    class Config:
        from_attributes = True


class MaintenanceTaskBase(BaseModel):
    scheduled_date: datetime


class MaintenanceTaskCreate(MaintenanceTaskBase):
    pass


class MaintenanceTaskRecord(BaseModel):
    status: TaskStatus = TaskStatus.COMPLETED
    actual_completion_date: Optional[datetime] = None
    executed_by: Optional[str] = None
    notes: Optional[str] = None


class MaintenanceTask(BaseModel):
    id: int
    task_code: str
    green_zone_id: int
    maintenance_grade: MaintenanceGrade
    scheduled_date: datetime
    status: TaskStatus
    actual_completion_date: Optional[datetime]
    executed_by: Optional[str]
    notes: Optional[str]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class FacilityBase(BaseModel):
    facility_code: str
    name: str
    facility_type: FacilityType
    is_safety_related: bool = False
    location_description: Optional[str] = None


class FacilityCreate(FacilityBase):
    park_code: str


class Facility(FacilityBase):
    id: int
    park_id: int
    status: FacilityStatus
    installation_date: Optional[datetime]
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class FacilityReportCreate(BaseModel):
    reporter_name: Optional[str] = None
    reporter_contact: Optional[str] = None
    damage_description: str


class FacilityReport(BaseModel):
    id: int
    report_code: str
    facility_id: int
    reporter_name: Optional[str]
    reporter_contact: Optional[str]
    damage_description: str
    reported_at: datetime
    deadline: Optional[datetime]
    handled_by: Optional[str]
    handled_at: Optional[datetime]
    handling_notes: Optional[str]
    is_handled: bool

    class Config:
        from_attributes = True


class FacilityStatusUpdate(BaseModel):
    status: FacilityStatus
    handled_by: Optional[str] = None
    handling_notes: Optional[str] = None


class ActivityBase(BaseModel):
    activity_code: str
    name: str
    organizer: Optional[str] = None
    organizer_contact: Optional[str] = None
    area: float
    start_time: datetime
    end_time: datetime
    expected_participants: Optional[int] = None


class ActivityCreate(ActivityBase):
    park_code: str


class Activity(ActivityBase):
    id: int
    park_id: int
    requires_security: bool
    status: ActivityStatus
    approved_by: Optional[str]
    approved_at: Optional[datetime]
    rejection_reason: Optional[str]
    completion_notes: Optional[str]
    created_at: datetime

    class Config:
        from_attributes = True


class ActivityApprove(BaseModel):
    approved: bool
    approved_by: Optional[str] = None
    rejection_reason: Optional[str] = None


class ActivityRecord(BaseModel):
    completion_notes: Optional[str] = None


class TaskScheduleItem(BaseModel):
    task_id: int
    task_code: str
    green_zone_id: int
    green_zone_name: str
    scheduled_date: datetime
    maintenance_grade: MaintenanceGrade
    status: TaskStatus
