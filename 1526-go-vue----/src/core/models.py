from dataclasses import dataclass, field
from datetime import date, datetime
from enum import Enum
from typing import List, Optional
from uuid import uuid4


class ProjectStatus(str, Enum):
    PLANNED = 'planned'
    IN_PROGRESS = 'in_progress'
    PENDING_ACCEPTANCE = 'pending_acceptance'
    ACCEPTED = 'accepted'
    NEEDS_REMEDIATION = 'needs_remediation'
    REDO_REQUIRED = 'redo_required'


class MilestoneStatus(str, Enum):
    NOT_STARTED = 'not_started'
    IN_PROGRESS = 'in_progress'
    COMPLETED = 'completed'
    OVERDUE = 'overdue'


class AcceptanceResult(str, Enum):
    PASS = 'pass'
    REMEDIATION_REQUIRED = 'remediation_required'
    FAIL = 'fail'


class AcceptanceStatus(str, Enum):
    INITIAL = 'initial'
    REMEDIATION_1 = 'remediation_1'
    REMEDIATION_2 = 'remediation_2'
    FINAL = 'final'


class RemediationStatus(str, Enum):
    NOT_REQUIRED = 'not_required'
    PENDING = 'pending'
    IN_PROGRESS = 'in_progress'
    COMPLETED = 'completed'
    FAILED = 'failed'


@dataclass
class BaseEntity:
    id: str = field(default_factory=lambda: str(uuid4()))
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)


@dataclass
class Project(BaseEntity):
    name: str = ''
    area: float = 0.0
    remediation_type: str = ''
    planned_start_date: Optional[date] = None
    planned_end_date: Optional[date] = None
    actual_start_date: Optional[date] = None
    actual_end_date: Optional[date] = None
    status: ProjectStatus = ProjectStatus.PLANNED
    description: str = ''


@dataclass
class Milestone(BaseEntity):
    project_id: str = ''
    name: str = ''
    description: str = ''
    planned_date: Optional[date] = None
    actual_date: Optional[date] = None
    status: MilestoneStatus = MilestoneStatus.NOT_STARTED
    overdue_days: int = 0


@dataclass
class Inspection(BaseEntity):
    project_id: str = ''
    title: str = ''
    content: str = ''
    inspection_date: Optional[date] = None
    inspector: str = ''
    issues_found: str = ''
    status: str = 'pending'


@dataclass
class MonitoringPoint(BaseEntity):
    project_id: str = ''
    name: str = ''
    location: str = ''
    description: str = ''


@dataclass
class MonitoringData(BaseEntity):
    point_id: str = ''
    monitoring_date: datetime = field(default_factory=datetime.now)
    indicator_1: float = 0.0
    indicator_2: float = 0.0
    indicator_3: float = 0.0
    indicator_4: float = 0.0


@dataclass
class AcceptanceScore:
    score_1: float = 0.0
    score_2: float = 0.0
    score_3: float = 0.0
    score_4: float = 0.0


@dataclass
class Acceptance(BaseEntity):
    project_id: str = ''
    status: AcceptanceStatus = AcceptanceStatus.INITIAL
    scores: AcceptanceScore = field(default_factory=AcceptanceScore)
    weighted_score: float = 0.0
    result: AcceptanceResult = AcceptanceResult.FAIL
    acceptance_date: Optional[date] = None
    comments: str = ''
    remediation_count: int = 0
    remediation_status: RemediationStatus = RemediationStatus.NOT_REQUIRED


@dataclass
class MonitoringAggregation:
    point_id: str = ''
    point_name: str = ''
    year: int = 0
    month: int = 0
    data_count: int = 0
    avg_indicator_1: float = 0.0
    avg_indicator_2: float = 0.0
    avg_indicator_3: float = 0.0
    avg_indicator_4: float = 0.0
    excluded: bool = False
    exclusion_reason: str = ''
