from datetime import date, datetime
from typing import List, Optional, Dict, Any
from pydantic import BaseModel, Field

from server.models import (
    ProjectStatus,
    MeasureType,
    TodoType,
    AcceptanceResult,
)


class ProjectCreate(BaseModel):
    name: str
    code: str
    description: Optional[str] = None
    location: Optional[str] = None
    start_date: date
    end_date: date
    total_budget: float = 0.0


class ProjectUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    location: Optional[str] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None
    status: Optional[ProjectStatus] = None


class BudgetAdjustment(BaseModel):
    new_total_budget: float


class BudgetItemCreate(BaseModel):
    measure_type: MeasureType
    budget_amount: float
    budget_ratio: float = 0.0


class MeasureCreate(BaseModel):
    measure_type: MeasureType
    name: str
    description: Optional[str] = None
    planned_quantity: float = 0.0
    unit: Optional[str] = None


class MeasureProgressCreate(BaseModel):
    year: int
    month: int
    completed_quantity: float = 0.0
    notes: Optional[str] = None


class ExpenditureCreate(BaseModel):
    budget_item_id: int
    amount: float
    description: Optional[str] = None
    expenditure_date: Optional[date] = None


class MonitoringCreate(BaseModel):
    year: int
    month: int
    erosion_modulus: float = 0.0
    vegetation_coverage: float = 0.0
    notes: Optional[str] = None


class AcceptanceCreate(BaseModel):
    engineering_score: float = 0.0
    plant_score: float = 0.0
    farming_score: float = 0.0
    temporary_score: float = 0.0
    notes: Optional[str] = None
    acceptance_date: Optional[date] = None


class ReinspectionCreate(BaseModel):
    engineering_score: float = 0.0
    plant_score: float = 0.0
    farming_score: float = 0.0
    temporary_score: float = 0.0
    notes: Optional[str] = None
    acceptance_date: Optional[date] = None


class TodoResolve(BaseModel):
    resolved: bool = True


class MeasureProgressOut(BaseModel):
    id: int
    measure_id: int
    year: int
    month: int
    completed_quantity: float
    progress_percent: float
    recorded_at: date
    notes: Optional[str]

    class Config:
        from_attributes = True


class MeasureOut(BaseModel):
    id: int
    project_id: int
    measure_type: MeasureType
    name: str
    description: Optional[str]
    planned_quantity: float
    unit: Optional[str]
    total_progress: float = 0.0
    progress_records: List[MeasureProgressOut] = []

    class Config:
        from_attributes = True


class ExpenditureOut(BaseModel):
    id: int
    amount: float
    description: Optional[str]
    expenditure_date: date
    created_at: datetime

    class Config:
        from_attributes = True


class BudgetItemOut(BaseModel):
    id: int
    measure_type: MeasureType
    budget_amount: float
    original_budget: float
    budget_ratio: float
    is_completed: bool
    total_spent: float = 0.0
    remaining: float = 0.0
    expenditures: List[ExpenditureOut] = []

    class Config:
        from_attributes = True


class MonitoringOut(BaseModel):
    id: int
    year: int
    month: int
    quarter: int
    erosion_modulus: float
    vegetation_coverage: float
    notes: Optional[str]
    recorded_at: date

    class Config:
        from_attributes = True


class AcceptanceOut(BaseModel):
    id: int
    project_id: int
    engineering_score: float
    plant_score: float
    farming_score: float
    temporary_score: float
    weighted_score: float
    result: Optional[AcceptanceResult]
    reinspection_count: int
    is_final: bool
    notes: Optional[str]
    acceptance_date: date
    created_at: datetime

    class Config:
        from_attributes = True


class TodoOut(BaseModel):
    id: int
    project_id: int
    todo_type: TodoType
    title: str
    description: Optional[str]
    is_resolved: bool
    resolved_at: Optional[datetime]
    related_measure_type: Optional[MeasureType]
    lag_amount: Optional[float]
    overrun_amount: Optional[float]
    created_at: datetime

    class Config:
        from_attributes = True


class ProjectOut(BaseModel):
    id: int
    name: str
    code: str
    description: Optional[str]
    location: Optional[str]
    status: ProjectStatus
    start_date: date
    end_date: date
    total_budget: float
    initial_total_budget: float
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class ProjectDetailOut(BaseModel):
    id: int
    name: str
    code: str
    description: Optional[str]
    location: Optional[str]
    status: ProjectStatus
    start_date: date
    end_date: date
    total_budget: float
    initial_total_budget: float
    progress: Dict[str, Any] = {}
    measures: List[MeasureOut] = []
    budgets: List[BudgetItemOut] = []
    latest_acceptance: Optional[AcceptanceOut] = None

    class Config:
        from_attributes = True


class TodoSummaryOut(BaseModel):
    project_id: int
    project_name: str
    status: str
    progress: Dict[str, Any]
    expenditure: Dict[str, Any]
    acceptance: Optional[Dict[str, Any]]
    unresolved_todos: List[Dict[str, Any]]


class YearlyAggregateOut(BaseModel):
    year: int
    avg_erosion_modulus: float
    avg_vegetation_coverage: float
    coverage_change_rate: Optional[float]
