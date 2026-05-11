from datetime import datetime
from typing import Optional, List
from pydantic import BaseModel, Field
from .models import StageType, SolutionType, SolutionStatus, AuditStatus


class AuditCreate(BaseModel):
    company_name: str = Field(..., min_length=1, max_length=255)
    company_id: Optional[str] = None
    start_date: Optional[datetime] = None


class StageResponse(BaseModel):
    id: int
    audit_id: int
    stage_type: StageType
    start_date: datetime
    end_date: Optional[datetime] = None
    is_completed: bool
    notes: Optional[str] = None

    class Config:
        from_attributes = True


class DimensionCreate(BaseModel):
    dimension_name: str
    score: int = Field(..., ge=0, le=10)


class DimensionResponse(BaseModel):
    id: int
    dimension_name: str
    score: int

    class Config:
        from_attributes = True


class SolutionCreate(BaseModel):
    name: str
    description: Optional[str] = None
    solution_type: SolutionType
    expected_energy_saving: Optional[float] = None
    expected_investment: Optional[float] = None


class SolutionResponse(BaseModel):
    id: int
    audit_id: int
    name: str
    description: Optional[str] = None
    solution_type: SolutionType
    status: SolutionStatus
    expected_energy_saving: Optional[float] = None
    expected_investment: Optional[float] = None
    actual_energy_saving: Optional[float] = None
    actual_investment: Optional[float] = None
    implementation_date: Optional[datetime] = None
    is_effective: Optional[bool] = None
    dimensions: List[DimensionResponse] = []

    class Config:
        from_attributes = True


class SolutionFilterResponse(BaseModel):
    solution_id: int
    solution_name: str
    passed: bool
    reason: Optional[str] = None


class AuditResponse(BaseModel):
    id: int
    company_name: str
    company_id: Optional[str] = None
    start_date: datetime
    acceptance_date: Optional[datetime] = None
    status: AuditStatus
    acceptance_score: Optional[float] = None
    is_overdue: bool
    stages: List[StageResponse] = []
    solutions: List[SolutionResponse] = []

    class Config:
        from_attributes = True


class SolutionImplementRequest(BaseModel):
    actual_energy_saving: float
    actual_investment: float


class AcceptanceRequest(BaseModel):
    score: float = Field(..., ge=0, le=100)
    notes: Optional[str] = None


class AuditLogResponse(BaseModel):
    id: int
    audit_id: Optional[int] = None
    user_id: Optional[str] = None
    user_name: Optional[str] = None
    action: str
    table_name: Optional[str] = None
    record_id: Optional[int] = None
    old_value: Optional[str] = None
    new_value: Optional[str] = None
    timestamp: datetime

    class Config:
        from_attributes = True


class AuditSummary(BaseModel):
    total_solutions: int
    passed_filter: int
    implemented: int
    effective_rate: float
    stages_completed: int
    total_stages: int
