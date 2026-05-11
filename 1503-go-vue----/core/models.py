from datetime import datetime, date
from enum import Enum
from typing import List, Optional
from pydantic import BaseModel, Field


class Role(str, Enum):
    ADMIN = "admin"
    AREA_MANAGER = "area_manager"
    RANGER = "ranger"


class AnomalyType(str, Enum):
    PEST = "pest"
    FIRE_RISK = "fire_risk"


class Severity(str, Enum):
    MILD = "mild"
    MODERATE = "moderate"
    SEVERE = "severe"


class FireRiskType(str, Enum):
    DRY_VEGETATION = "dry_vegetation"
    DEBRIS = "debris"
    OPEN_FIRE = "open_fire"
    ELECTRICAL_HAZARD = "electrical_hazard"


class FireRiskStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    RESOLVED = "resolved"


class TodoStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    RESOLVED = "resolved"
    OVERDUE = "overdue"


class ReportStatus(str, Enum):
    PENDING = "pending"
    QUALIFIED = "qualified"
    UNQUALIFIED = "unqualified"


class Location(BaseModel):
    latitude: float = Field(..., ge=-90, le=90)
    longitude: float = Field(..., ge=-180, le=180)


class User(BaseModel):
    id: Optional[int] = None
    name: str
    role: Role
    phone: Optional[str] = None


class ForestArea(BaseModel):
    id: Optional[int] = None
    name: str
    area_km2: float = Field(..., gt=0)
    main_tree_species: str
    ranger_id: Optional[int] = None


class PatrolTask(BaseModel):
    id: Optional[int] = None
    area_id: int
    ranger_id: int
    patrol_date: date
    route_keypoints: List[Location]
    created_at: Optional[datetime] = None


class PestAnomaly(BaseModel):
    id: Optional[int] = None
    type: str
    severity: Severity
    trees_affected: int = Field(..., ge=0)
    location: Location


class FireRiskAnomaly(BaseModel):
    id: Optional[int] = None
    type: FireRiskType
    status: FireRiskStatus
    location: Location


class Anomaly(BaseModel):
    id: Optional[int] = None
    report_id: int
    anomaly_type: AnomalyType
    pest: Optional[PestAnomaly] = None
    fire_risk: Optional[FireRiskAnomaly] = None
    location: Location
    is_duplicate: bool = False
    created_at: Optional[datetime] = None


class PatrolReport(BaseModel):
    id: Optional[int] = None
    task_id: int
    submitted_at: Optional[datetime] = None
    actual_route: List[Location]
    anomalies: List[Anomaly] = []
    status: ReportStatus = ReportStatus.PENDING


class Todo(BaseModel):
    id: Optional[int] = None
    anomaly_id: int
    anomaly_type: AnomalyType
    area_id: int
    assigned_user_id: int
    status: TodoStatus = TodoStatus.PENDING
    created_at: Optional[datetime] = None
    resolved_at: Optional[datetime] = None


class MonthlyPestStats(BaseModel):
    year: int
    month: int
    new_discovery_count: int
    trees_affected_total: int
    severe_percentage: float


class AreaHealthScore(BaseModel):
    area_id: int
    area_name: str
    score: float
    anomaly_count: int
    severe_anomaly_count: int
