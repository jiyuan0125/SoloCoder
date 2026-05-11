from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date, datetime
from enum import Enum
from typing import List, Optional
from uuid import UUID, uuid4


class SafetyLevel(str, Enum):
    ONE = "one"
    TWO = "two"
    THREE = "three"
    FOUR = "four"
    FIVE = "five"


class InspectionStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"
    OVERDUE = "overdue"


class MaterialStatus(str, Enum):
    NOT_SUBMITTED = "not_submitted"
    SUBMITTED = "submitted"
    UNDER_REVIEW = "under_review"
    APPROVED = "approved"
    REJECTED = "rejected"


class AcceptanceStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    APPROVED = "approved"


@dataclass
class ReviewRecord:
    reviewer: str
    opinion: str
    is_approved: bool
    id: UUID = field(default_factory=uuid4)
    review_date: datetime = field(default_factory=datetime.now)


@dataclass
class AcceptanceMaterial:
    name: str
    status: MaterialStatus = MaterialStatus.NOT_SUBMITTED
    content: Optional[str] = None
    submission_date: Optional[datetime] = None
    id: UUID = field(default_factory=uuid4)
    review_records: List[ReviewRecord] = field(default_factory=list)


@dataclass
class ClosureAcceptance:
    pond_id: UUID
    status: AcceptanceStatus = AcceptanceStatus.PENDING
    id: UUID = field(default_factory=uuid4)
    materials: List[AcceptanceMaterial] = field(default_factory=list)
    created_at: datetime = field(default_factory=datetime.now)
    approval_date: Optional[datetime] = None


@dataclass
class TailingPond:
    name: str
    capacity: float
    dam_height: float
    safety_level: SafetyLevel
    id: UUID = field(default_factory=uuid4)
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)


@dataclass
class MonitoringSection:
    pond_id: UUID
    name: str
    id: UUID = field(default_factory=uuid4)
    created_at: datetime = field(default_factory=datetime.now)


@dataclass
class MonitoringData:
    section_id: UUID
    timestamp: datetime
    dry_beach_length: Optional[float] = None
    phreatic_line: Optional[float] = None
    dam_displacement: Optional[float] = None
    water_level: Optional[float] = None
    id: UUID = field(default_factory=uuid4)


@dataclass
class AggregatedMonitoringData:
    section_id: UUID
    data_type: str
    period_start: datetime
    period_end: datetime
    value: float
    id: UUID = field(default_factory=uuid4)
    created_at: datetime = field(default_factory=datetime.now)


@dataclass
class InspectionTask:
    pond_id: UUID
    planned_date: date
    actual_date: Optional[date] = None
    inspector: Optional[str] = None
    remarks: Optional[str] = None
    status: InspectionStatus = InspectionStatus.PENDING
    id: UUID = field(default_factory=uuid4)
    created_at: datetime = field(default_factory=datetime.now)


@dataclass
class Warning:
    pond_id: UUID
    section_id: UUID
    warning_type: str
    message: str
    value: float
    threshold: float
    id: UUID = field(default_factory=uuid4)
    timestamp: datetime = field(default_factory=datetime.now)
    is_acknowledged: bool = False
