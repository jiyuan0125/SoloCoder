from datetime import date
from typing import List, Optional, Dict
from pydantic import BaseModel, Field, field_validator


class Coordinates(BaseModel):
    latitude: float
    longitude: float
    elevation: Optional[float] = None


class ProjectStatus(str):
    PLANNING = "planning"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class Project(BaseModel):
    id: Optional[str] = None
    name: str
    exploration_area: str
    mineral_targets: List[str]
    start_date: date
    planned_end_date: date
    actual_end_date: Optional[date] = None
    status: str = ProjectStatus.PLANNING
    description: Optional[str] = None

    @field_validator('mineral_targets')
    def validate_mineral_targets(cls, v):
        if not v or len(v) == 0:
            raise ValueError('必须至少指定一个矿种目标')
        return v

    @field_validator('status')
    def validate_status(cls, v):
        valid_statuses = [
            ProjectStatus.PLANNING,
            ProjectStatus.IN_PROGRESS,
            ProjectStatus.COMPLETED,
            ProjectStatus.CANCELLED
        ]
        if v not in valid_statuses:
            raise ValueError(f'无效的项目状态: {v}')
        return v


class Borehole(BaseModel):
    id: Optional[str] = None
    project_id: str
    name: str
    code: str
    coordinates: Coordinates
    designed_depth: float
    actual_depth: Optional[float] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None
    is_completed: bool = False
    remarks: Optional[str] = None

    @field_validator('designed_depth')
    def validate_designed_depth(cls, v):
        if v <= 0:
            raise ValueError('设计孔深必须大于0')
        return v


class Sample(BaseModel):
    id: Optional[str] = None
    borehole_id: str
    sample_number: str
    start_depth: float
    end_depth: float
    lithology: str
    sampling_date: Optional[date] = None
    lab_received_date: Optional[date] = None
    analysis_results: Dict[str, float] = {}
    remarks: Optional[str] = None

    @field_validator('start_depth', 'end_depth')
    def validate_depths(cls, v):
        if v < 0:
            raise ValueError('深度不能为负数')
        return v

    @field_validator('analysis_results')
    def validate_analysis_results(cls, v):
        for element, value in v.items():
            if value < 0:
                raise ValueError(f'元素 {element} 的含量不能为负数')
        return v


class TodoItem(BaseModel):
    id: Optional[str] = None
    project_id: str
    borehole_id: str
    message: str
    due_date: date
    is_completed: bool = False
    created_at: date
