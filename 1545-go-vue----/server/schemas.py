from datetime import datetime, date
from typing import Optional
from pydantic import BaseModel, Field

from .models import ControlLevel, UserRole, EvaluationStatus


class UserBase(BaseModel):
    username: str
    full_name: Optional[str] = None


class UserCreate(UserBase):
    password: str
    role: UserRole = UserRole.USER


class UserLogin(BaseModel):
    username: str
    password: str


class UserResponse(UserBase):
    id: int
    role: UserRole
    is_active: bool
    created_at: datetime

    class Config:
        from_attributes = True


class Token(BaseModel):
    access_token: str
    token_type: str = "bearer"


class SectionBase(BaseModel):
    name: str
    code: str
    river: Optional[str] = None
    control_level: ControlLevel
    description: Optional[str] = None


class SectionCreate(SectionBase):
    pass


class SectionUpdate(BaseModel):
    name: Optional[str] = None
    code: Optional[str] = None
    river: Optional[str] = None
    control_level: Optional[ControlLevel] = None
    description: Optional[str] = None


class StandardLimitBase(BaseModel):
    ph_min: Optional[float] = None
    ph_max: Optional[float] = None
    do_limit: Optional[float] = None
    cod_limit: Optional[float] = None
    nh3n_limit: Optional[float] = None
    tp_limit: Optional[float] = None
    tn_limit: Optional[float] = None
    ph_class: Optional[str] = None
    do_class: Optional[str] = None
    cod_class: Optional[str] = None
    nh3n_class: Optional[str] = None
    tp_class: Optional[str] = None
    tn_class: Optional[str] = None
    overall_class: Optional[str] = None
    effective_date: Optional[date] = None
    description: Optional[str] = None


class StandardLimitCreate(StandardLimitBase):
    section_id: int


class StandardLimitUpdate(BaseModel):
    ph_min: Optional[float] = None
    ph_max: Optional[float] = None
    do_limit: Optional[float] = None
    cod_limit: Optional[float] = None
    nh3n_limit: Optional[float] = None
    tp_limit: Optional[float] = None
    tn_limit: Optional[float] = None
    ph_class: Optional[str] = None
    do_class: Optional[str] = None
    cod_class: Optional[str] = None
    nh3n_class: Optional[str] = None
    tp_class: Optional[str] = None
    tn_class: Optional[str] = None
    overall_class: Optional[str] = None
    effective_date: Optional[date] = None
    description: Optional[str] = None


class StandardLimitResponse(StandardLimitBase):
    id: int
    section_id: int
    is_current: bool
    created_at: datetime

    class Config:
        from_attributes = True


class SectionResponse(SectionBase):
    id: int
    created_at: datetime
    current_standard: Optional[StandardLimitResponse] = None

    class Config:
        from_attributes = True


class MonitoringDataBase(BaseModel):
    monitoring_date: date
    ph_value: Optional[float] = None
    do_value: Optional[float] = None
    cod_value: Optional[float] = None
    nh3n_value: Optional[float] = None
    tp_value: Optional[float] = None
    tn_value: Optional[float] = None
    sampler_name: Optional[str] = None
    sampler_id: Optional[str] = None
    remarks: Optional[str] = None


class MonitoringDataCreate(MonitoringDataBase):
    section_id: int
    task_id: Optional[int] = None


class MonitoringDataResponse(MonitoringDataBase):
    id: int
    section_id: int
    task_id: Optional[int] = None
    created_at: datetime

    class Config:
        from_attributes = True


class MonthlyEvaluationBase(BaseModel):
    section_id: int
    year: int
    month: int


class MonthlyEvaluationResponse(MonthlyEvaluationBase):
    id: int
    total_count: int
    pass_count: int
    pass_rate: Optional[float] = None
    is_meet_standard: Optional[bool] = None
    ph_pass_count: int
    do_pass_count: int
    cod_pass_count: int
    nh3n_pass_count: int
    tp_pass_count: int
    tn_pass_count: int
    worst_indicator: Optional[str] = None
    worst_value: Optional[float] = None
    status: EvaluationStatus
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class MonitoringTaskBase(BaseModel):
    section_id: int
    scheduled_date: date
    is_flood_season: bool = False


class MonitoringTaskResponse(MonitoringTaskBase):
    id: int
    status: str
    created_at: datetime

    class Config:
        from_attributes = True


class AuditLogBase(BaseModel):
    user_id: int
    action: str
    target_type: str
    target_id: Optional[int] = None
    details: str


class AuditLogResponse(AuditLogBase):
    id: int
    created_at: datetime
    operator: Optional[UserResponse] = None

    class Config:
        from_attributes = True


class SectionDetailResponse(BaseModel):
    section: SectionResponse
    monitoring_data: list[MonitoringDataResponse]
    evaluations: list[MonthlyEvaluationResponse]
    standards_history: list[StandardLimitResponse]


class TaskGenerationRequest(BaseModel):
    year: int
    month: int
