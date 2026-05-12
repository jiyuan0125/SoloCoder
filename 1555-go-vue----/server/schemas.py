from pydantic import BaseModel, Field
from datetime import datetime
from typing import Optional, List
from server.models import PilotLevel, ShipType, TaskStatus, WeatherCondition


class PilotCreate(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    level: PilotLevel
    is_on_duty: bool = False


class PilotResponse(BaseModel):
    id: int
    name: str
    level: PilotLevel
    is_on_duty: bool
    total_work_hours: float
    monthly_work_hours: float
    monthly_task_count: int
    created_at: datetime
    
    class Config:
        from_attributes = True


class PilotApplicationCreate(BaseModel):
    ship_name: str = Field(..., min_length=1, max_length=100)
    ship_type: ShipType
    from_location: str = Field(..., min_length=1, max_length=100)
    to_location: str = Field(..., min_length=1, max_length=100)
    requested_start_time: datetime
    estimated_duration_hours: float = Field(..., gt=0)
    remarks: Optional[str] = None


class PilotApplicationResponse(BaseModel):
    id: int
    ship_name: str
    ship_type: ShipType
    from_location: str
    to_location: str
    requested_start_time: datetime
    estimated_duration_hours: float
    remarks: Optional[str]
    status: str
    created_at: datetime
    
    class Config:
        from_attributes = True


class PilotTaskResponse(BaseModel):
    id: int
    application_id: int
    pilot_id: int
    status: TaskStatus
    assigned_at: datetime
    started_at: Optional[datetime]
    completed_at: Optional[datetime]
    suspended_at: Optional[datetime]
    actual_duration_hours: Optional[float]
    is_abnormal: bool
    suspension_reason: Optional[str]
    progress_percent: int
    
    class Config:
        from_attributes = True


class PilotTaskDetailResponse(PilotTaskResponse):
    application: Optional[PilotApplicationResponse] = None
    pilot: Optional[PilotResponse] = None


class DispatchRecommendation(BaseModel):
    pilot_id: int
    pilot_name: str
    level: PilotLevel
    is_on_duty: bool
    match_score: int
    reason: str


class SuspensionRequest(BaseModel):
    reason: str


class WeatherUpdate(BaseModel):
    condition: WeatherCondition
    description: Optional[str] = None


class WeatherResponse(BaseModel):
    id: int
    condition: WeatherCondition
    description: Optional[str]
    updated_at: datetime
    
    class Config:
        from_attributes = True


class TimeWindowConfigResponse(BaseModel):
    id: int
    name: str
    start_time: str
    end_time: str
    description: Optional[str]
    is_active: bool
    rule_type: str
    
    class Config:
        from_attributes = True


class DailyStats(BaseModel):
    date: str
    total_tasks: int
    completed_tasks: int
    pending_tasks: int
    suspended_tasks: int


class PilotWorkload(BaseModel):
    pilot_id: int
    pilot_name: str
    level: PilotLevel
    today_tasks: int
    today_hours: float
    monthly_hours: float
    monthly_tasks: int
    approaching_monthly_limit: bool


class StatisticsResponse(BaseModel):
    daily_stats: DailyStats
    pilot_workloads: List[PilotWorkload]
    average_waiting_time_hours: float
