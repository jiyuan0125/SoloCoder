from datetime import date, datetime
from enum import Enum
from typing import Optional, List
from uuid import uuid4

from pydantic import BaseModel, Field, validator


class FreshLeafGrade(str, Enum):
    A = "A"
    B = "B"
    C = "C"


class FinishedGrade(str, Enum):
    SPECIAL = "特级"
    FIRST = "一级"
    SECOND = "二级"
    THIRD = "三级"


class AlertType(str, Enum):
    OVERDUE_PROCESSING = "overdue_processing"


class AlertStatus(str, Enum):
    ACTIVE = "active"
    RESOLVED = "resolved"


class TodoStatus(str, Enum):
    PENDING = "pending"
    COMPLETED = "completed"


class Plot(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    name: str
    location: str
    area: float
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)


class HarvestPlan(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    plot_id: str
    plan_date: date
    expected_quantity: float
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)

    @validator("plan_date")
    def validate_plan_date_not_past(cls, v: date) -> date:
        if v < date.today():
            raise ValueError("采摘计划日期不能是过去的")
        return v


class HarvestRecord(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    plan_id: str
    actual_quantity: float
    fresh_leaf_grade: FreshLeafGrade
    harvest_time: datetime = Field(default_factory=datetime.now)
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)

    @validator("actual_quantity")
    def validate_quantity_positive(cls, v: float) -> float:
        if v <= 0:
            raise ValueError("实际采摘量必须大于0")
        return v


class ProcessingBatch(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    harvest_record_id: str
    input_quantity: float
    output_quantity: Optional[float] = None
    start_time: datetime = Field(default_factory=datetime.now)
    end_time: Optional[datetime] = None
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)

    @validator("input_quantity")
    def validate_input_quantity_positive(cls, v: float) -> float:
        if v <= 0:
            raise ValueError("投叶量必须大于0")
        return v

    @validator("output_quantity")
    def validate_output_quantity(cls, v: Optional[float], values) -> Optional[float]:
        if v is None:
            return v
        if v < 0:
            raise ValueError("成品量不能为负数")
        input_quantity = values.get("input_quantity")
        if input_quantity and v > input_quantity * 0.4:
            raise ValueError("成品量不能超过投叶量的40%")
        return v


class QualityRating(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    batch_id: str
    grade: FinishedGrade
    sensory_description: str
    rating_time: datetime = Field(default_factory=datetime.now)
    created_at: datetime = Field(default_factory=datetime.now)
    updated_at: datetime = Field(default_factory=datetime.now)

    @validator("grade")
    def validate_grade_not_empty(cls, v: FinishedGrade) -> FinishedGrade:
        if v is None:
            raise ValueError("等级不能为空")
        return v

    @validator("sensory_description")
    def validate_description_not_empty(cls, v: str) -> str:
        if not v or not v.strip():
            raise ValueError("感官描述不能为空")
        return v


class Alert(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    type: AlertType
    harvest_record_id: str
    message: str
    status: AlertStatus = AlertStatus.ACTIVE
    created_at: datetime = Field(default_factory=datetime.now)
    resolved_at: Optional[datetime] = None


class Todo(BaseModel):
    id: str = Field(default_factory=lambda: uuid4().hex)
    plan_id: str
    title: str
    description: str
    due_date: date
    status: TodoStatus = TodoStatus.PENDING
    created_at: datetime = Field(default_factory=datetime.now)
    completed_at: Optional[datetime] = None
