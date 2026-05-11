from datetime import datetime, date
from typing import Optional, List, Any
from pydantic import BaseModel, Field
from src.core.models import (
    Severity, FireRiskType, PestType, ProcessingStatus, TodoStatus, UserRole
)


class UserBase(BaseModel):
    name: str
    phone: str
    role: UserRole


class UserCreate(UserBase):
    pass


class UserRead(UserBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class Waypoint(BaseModel):
    lat: float
    lng: float


class AreaBase(BaseModel):
    name: str
    area_km2: float
    main_tree_species: str


class AreaCreate(AreaBase):
    ranger_id: Optional[int] = None
    manager_id: Optional[int] = None


class AreaRead(AreaBase):
    id: int
    ranger_id: Optional[int] = None
    manager_id: Optional[int] = None
    created_at: datetime

    class Config:
        from_attributes = True


class AreaDetail(AreaRead):
    ranger: Optional[UserRead] = None
    manager: Optional[UserRead] = None


class PatrolTaskBase(BaseModel):
    area_id: int
    ranger_id: int
    patrol_date: date
    route_waypoints: List[Waypoint]


class PatrolTaskCreate(PatrolTaskBase):
    pass


class PatrolTaskRead(BaseModel):
    id: int
    area_id: int
    ranger_id: int
    patrol_date: date
    route_waypoints: List[Waypoint]
    created_at: datetime

    class Config:
        from_attributes = True


class PatrolTaskDetail(PatrolTaskRead):
    area: AreaRead
    ranger: UserRead


class PatrolReportBase(BaseModel):
    task_id: int
    actual_route: List[Waypoint]


class PatrolReportCreate(PatrolReportBase):
    pass


class PatrolReportRead(BaseModel):
    id: int
    task_id: int
    area_id: int
    ranger_id: int
    actual_route: List[Waypoint]
    is_qualified: int
    submitted_at: datetime

    class Config:
        from_attributes = True


class PestAnomalyBase(BaseModel):
    pest_type: PestType
    severity: Severity
    trees_affected: int = 1
    notes: Optional[str] = None


class FireRiskAnomalyBase(BaseModel):
    risk_type: FireRiskType
    status: ProcessingStatus = ProcessingStatus.PENDING
    notes: Optional[str] = None


class AnomalyBase(BaseModel):
    location_lat: float
    location_lng: float
    description: Optional[str] = None


class PestAnomalyCreate(PestAnomalyBase):
    pass


class FireRiskAnomalyCreate(FireRiskAnomalyBase):
    pass


class AnomalyCreate(AnomalyBase):
    pest_data: Optional[PestAnomalyCreate] = None
    fire_risk_data: Optional[FireRiskAnomalyCreate] = None


class AnomalyRead(BaseModel):
    id: int
    report_id: int
    anomaly_type: str
    location_lat: float
    location_lng: float
    description: Optional[str] = None
    is_duplicate: int
    duplicate_of_id: Optional[int] = None
    created_at: datetime

    class Config:
        from_attributes = True


class PestAnomalyRead(PestAnomalyBase):
    id: int
    anomaly_id: int

    class Config:
        from_attributes = True


class FireRiskAnomalyRead(FireRiskAnomalyBase):
    id: int
    anomaly_id: int

    class Config:
        from_attributes = True


class AnomalyDetail(AnomalyRead):
    pest_data: Optional[PestAnomalyRead] = None
    fire_risk_data: Optional[FireRiskAnomalyRead] = None


class ReportWithAnomalies(PatrolReportRead):
    anomalies: List[AnomalyDetail] = []


class TodoBase(BaseModel):
    pass


class TodoRead(BaseModel):
    id: int
    anomaly_id: int
    assignee_id: int
    status: TodoStatus
    priority: int
    due_date: date
    notes: Optional[str] = None
    created_at: datetime
    completed_at: Optional[datetime] = None

    class Config:
        from_attributes = True


class TodoDetail(TodoRead):
    anomaly: Optional[AnomalyDetail] = None
    assignee: Optional[UserRead] = None


class TodoUpdate(BaseModel):
    status: Optional[TodoStatus] = None
    notes: Optional[str] = None


class MonthlyPestStats(BaseModel):
    year: int
    month: int
    total_anomalies: int
    trees_affected: int
    severe_ratio: float
    severe_count: int


class AreaHealthScore(BaseModel):
    area_id: int
    area_name: str
    score: float
    total_anomalies: int
    severe_anomalies: int
    fire_risk_count: int
    pest_count: int


class MonthlyPestStatsResponse(BaseModel):
    stats: List[MonthlyPestStats]


class AreaHealthResponse(BaseModel):
    scores: List[AreaHealthScore]
