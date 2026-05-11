from datetime import date, datetime
from typing import List, Optional

from pydantic import BaseModel, Field

from src.core.models import (
    AcceptanceResult,
    AcceptanceStatus,
    MilestoneStatus,
    ProjectStatus,
    RemediationStatus
)


class Message(BaseModel):
    message: str


class ProjectCreate(BaseModel):
    name: str
    area: float = Field(gt=0)
    remediation_type: str
    planned_start_date: Optional[date] = None
    planned_end_date: Optional[date] = None
    description: str = ''


class ProjectUpdate(BaseModel):
    name: Optional[str] = None
    area: Optional[float] = Field(None, gt=0)
    remediation_type: Optional[str] = None
    planned_start_date: Optional[date] = None
    planned_end_date: Optional[date] = None
    actual_start_date: Optional[date] = None
    actual_end_date: Optional[date] = None
    status: Optional[ProjectStatus] = None
    description: Optional[str] = None


class MilestoneCreate(BaseModel):
    project_id: str
    name: str
    description: str = ''
    planned_date: Optional[date] = None


class MilestoneUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    planned_date: Optional[date] = None
    actual_date: Optional[date] = None
    status: Optional[MilestoneStatus] = None


class InspectionCreate(BaseModel):
    project_id: str
    title: str
    content: str
    inspection_date: Optional[date] = None
    inspector: str = ''
    issues_found: str = ''
    status: str = 'pending'


class InspectionUpdate(BaseModel):
    title: Optional[str] = None
    content: Optional[str] = None
    inspection_date: Optional[date] = None
    inspector: Optional[str] = None
    issues_found: Optional[str] = None
    status: Optional[str] = None


class MonitoringPointCreate(BaseModel):
    project_id: str
    name: str
    location: str = ''
    description: str = ''


class MonitoringDataCreate(BaseModel):
    point_id: str
    monitoring_date: Optional[datetime] = None
    indicator_1: float = 0.0
    indicator_2: float = 0.0
    indicator_3: float = 0.0
    indicator_4: float = 0.0


class AcceptanceCreate(BaseModel):
    project_id: str
    score_1: float = Field(0.0, ge=0, le=100)
    score_2: float = Field(0.0, ge=0, le=100)
    score_3: float = Field(0.0, ge=0, le=100)
    score_4: float = Field(0.0, ge=0, le=100)
    comments: str = ''


class RemediationUpdate(BaseModel):
    remediation_status: RemediationStatus
