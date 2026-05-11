from datetime import date, datetime
from enum import Enum
from typing import Optional, List
from pydantic import BaseModel, Field
import uuid


class OperationType(str, Enum):
    FERTILIZATION = "fertilization"
    PESTICIDE = "pesticide"
    IRRIGATION = "irrigation"


class HarvestStatus(str, Enum):
    PLANNED = "planned"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    CANCELLED = "cancelled"


class TodoType(str, Enum):
    SAFETY_INTERVAL = "safety_interval"
    GAP_INSPECTION = "gap_inspection"


class TodoStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"
    OVERDUE = "overdue"


class Plot(BaseModel):
    id: str = Field(default_factory=lambda: uuid.uuid4().hex[:12])
    name: str
    area: float
    variety: str
    planting_date: date
    expected_harvest_date: date
    created_at: datetime = Field(default_factory=datetime.now)
    notes: Optional[str] = None


class FarmOperation(BaseModel):
    id: str = Field(default_factory=lambda: uuid.uuid4().hex[:12])
    plot_id: str
    operation_type: OperationType
    operation_date: date
    details: str
    quantity: Optional[str] = None
    safety_interval_days: Optional[int] = None
    created_at: datetime = Field(default_factory=datetime.now)


class Harvest(BaseModel):
    id: str = Field(default_factory=lambda: uuid.uuid4().hex[:12])
    plot_id: str
    planned_date: date
    status: HarvestStatus = HarvestStatus.PLANNED
    actual_date: Optional[date] = None
    quantity: Optional[float] = None
    unit: Optional[str] = None
    quality_status: Optional[str] = None
    notes: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class Processing(BaseModel):
    id: str = Field(default_factory=lambda: uuid.uuid4().hex[:12])
    harvest_id: str
    batch_number: str
    input_quantity: float
    output_quantity: Optional[float] = None
    unit: Optional[str] = None
    processing_date: date
    details: str
    created_at: datetime = Field(default_factory=datetime.now)


class Todo(BaseModel):
    id: str = Field(default_factory=lambda: uuid.uuid4().hex[:12])
    todo_type: TodoType
    title: str
    description: str
    due_date: date
    related_plot_id: Optional[str] = None
    related_operation_id: Optional[str] = None
    status: TodoStatus = TodoStatus.PENDING
    created_at: datetime = Field(default_factory=datetime.now)


class PlotCreate(BaseModel):
    name: str
    area: float
    variety: str
    planting_date: date
    expected_harvest_date: date
    notes: Optional[str] = None


class PlotUpdate(BaseModel):
    name: Optional[str] = None
    area: Optional[float] = None
    variety: Optional[str] = None
    planting_date: Optional[date] = None
    expected_harvest_date: Optional[date] = None
    notes: Optional[str] = None


class FarmOperationCreate(BaseModel):
    plot_id: str
    operation_type: OperationType
    operation_date: date
    details: str
    quantity: Optional[str] = None
    safety_interval_days: Optional[int] = None


class HarvestCreate(BaseModel):
    plot_id: str
    planned_date: date
    quantity: Optional[float] = None
    unit: Optional[str] = None
    notes: Optional[str] = None


class HarvestUpdate(BaseModel):
    status: Optional[HarvestStatus] = None
    actual_date: Optional[date] = None
    quantity: Optional[float] = None
    unit: Optional[str] = None
    quality_status: Optional[str] = None
    notes: Optional[str] = None


class ProcessingCreate(BaseModel):
    harvest_id: str
    batch_number: str
    input_quantity: float
    output_quantity: Optional[float] = None
    unit: Optional[str] = None
    processing_date: date
    details: str


class ProcessingUpdate(BaseModel):
    output_quantity: Optional[float] = None
    details: Optional[str] = None


class TodoUpdate(BaseModel):
    status: TodoStatus
