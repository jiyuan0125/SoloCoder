from datetime import datetime, date
from typing import Optional, List
from pydantic import BaseModel, Field


class PipelineBase(BaseModel):
    name: str
    code: str
    description: Optional[str] = None
    total_length: float
    design_wall_thickness: float
    operating_pressure: float
    material: Optional[str] = None
    diameter: Optional[float] = None
    installation_date: Optional[date] = None
    risk_level: str = "medium"
    status: str = "active"


class PipelineCreate(PipelineBase):
    pass


class PipelineUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    total_length: Optional[float] = None
    design_wall_thickness: Optional[float] = None
    operating_pressure: Optional[float] = None
    material: Optional[str] = None
    diameter: Optional[float] = None
    risk_level: Optional[str] = None
    status: Optional[str] = None


class PipelineResponse(PipelineBase):
    id: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class SegmentBase(BaseModel):
    pipeline_id: int
    name: str
    start_position: float
    end_position: float
    length: float
    description: Optional[str] = None
    risk_level: str = "medium"
    status: str = "operational"


class SegmentCreate(SegmentBase):
    pass


class SegmentUpdate(BaseModel):
    name: Optional[str] = None
    start_position: Optional[float] = None
    end_position: Optional[float] = None
    length: Optional[float] = None
    description: Optional[str] = None
    risk_level: Optional[str] = None
    status: Optional[str] = None


class SegmentResponse(SegmentBase):
    id: int
    patrol_frequency_days: int
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class PressurePointBase(BaseModel):
    pipeline_id: int
    name: str
    position: float


class PressurePointCreate(PressurePointBase):
    pass


class PressurePointResponse(PressurePointBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class PressureReadingCreate(BaseModel):
    pressure_point_id: int
    pressure: float


class PressureReadingResponse(BaseModel):
    id: int
    pressure_point_id: int
    pressure: float
    timestamp: datetime
    is_aggregated: bool

    class Config:
        from_attributes = True


class PressureStatsResponse(BaseModel):
    mean: float
    stddev: float
    count: int


class LeakCheckResponse(BaseModel):
    is_suspected: bool
    pressure_diff: Optional[float] = None
    mean: Optional[float] = None
    stddev: Optional[float] = None
    threshold: Optional[float] = None
    data_count: Optional[int] = None


class AlarmBase(BaseModel):
    alarm_type: str
    severity: str = "medium"
    pipeline_id: int
    segment_id: Optional[int] = None
    title: str
    description: Optional[str] = None


class AlarmCreate(AlarmBase):
    pass


class AlarmUpdateStatus(BaseModel):
    status: str
    handled_by: Optional[str] = None
    false_alarm_reason: Optional[str] = None


class AlarmResponse(AlarmBase):
    id: int
    status: str
    created_at: datetime
    updated_at: datetime
    resolved_at: Optional[datetime] = None
    handled_by: Optional[str] = None
    escalated: bool
    reminders: int
    false_alarm_reason: Optional[str] = None

    class Config:
        from_attributes = True


class PatrolPlanItemResponse(BaseModel):
    id: int
    segment_id: int
    scheduled_time: Optional[datetime] = None
    assigned_to: Optional[str] = None
    completed: bool

    class Config:
        from_attributes = True


class PatrolPlanResponse(BaseModel):
    id: int
    pipeline_id: int
    plan_date: date
    status: str
    created_at: datetime
    plan_items: List[PatrolPlanItemResponse] = []

    class Config:
        from_attributes = True


class PatrolRecordCreate(BaseModel):
    segment_id: int
    patrol_date: date
    inspector: Optional[str] = None
    findings: Optional[str] = None
    status: str = "normal"


class PatrolRecordResponse(BaseModel):
    id: int
    segment_id: int
    patrol_date: date
    inspector: Optional[str] = None
    findings: Optional[str] = None
    status: str
    created_at: datetime

    class Config:
        from_attributes = True


class IntegrityRecordCreate(BaseModel):
    pipeline_id: int
    segment_id: Optional[int] = None
    inspection_date: date
    wall_thickness: float
    location: Optional[str] = None
    notes: Optional[str] = None


class IntegrityRecordResponse(BaseModel):
    id: int
    pipeline_id: int
    segment_id: Optional[int] = None
    inspection_date: date
    wall_thickness: float
    location: Optional[str] = None
    notes: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class WallThicknessCheckResponse(BaseModel):
    needs_maintenance: bool
    design_thickness: Optional[float] = None
    threshold: Optional[float] = None
    current_thickness: Optional[float] = None
    ratio: Optional[float] = None


class MaintenanceTaskResponse(BaseModel):
    id: int
    task_type: str
    pipeline_id: int
    segment_id: Optional[int] = None
    title: str
    description: Optional[str] = None
    priority: str
    status: str
    due_date: Optional[date] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class MaintenanceTaskUpdate(BaseModel):
    status: Optional[str] = None
    priority: Optional[str] = None


class DailyReportResponse(BaseModel):
    id: int
    report_date: date
    total_pipelines: int
    active_pipelines: int
    total_alarms: int
    new_alarms: int
    resolved_alarms: int
    pressure_readings_count: int
    patrols_completed: int
    maintenance_tasks: int
    content: str
    created_at: datetime

    class Config:
        from_attributes = True
