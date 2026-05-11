from datetime import date, datetime
from typing import List, Optional
from enum import Enum

from pydantic import BaseModel, Field


class TodoPriority(str, Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    URGENT = "urgent"


class TodoStatus(str, Enum):
    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"


class SafetyCheckStatus(str, Enum):
    QUALIFIED = "qualified"
    UNQUALIFIED = "unqualified"


class SafetyCheckItem(BaseModel):
    name: str
    is_abnormal: bool


class Mine(BaseModel):
    id: Optional[int] = None
    name: str
    mineral_type: str
    annual_capacity: float = Field(gt=0, description="年产能（吨）")


class MiningOperation(BaseModel):
    id: Optional[int] = None
    mine_id: int
    operation_date: date
    planned_output: float = Field(gt=0, description="计划产量（吨）")
    actual_output: float = Field(ge=0, description="实际产量（吨）")


class Transport(BaseModel):
    id: Optional[int] = None
    mine_id: int
    vehicle_number: str
    transport_date: date
    departure_time: datetime
    arrival_time: datetime
    transport_volume: float = Field(gt=0, description="运输量（吨）")


class SafetyCheck(BaseModel):
    id: Optional[int] = None
    mine_id: int
    check_date: date
    items: List[SafetyCheckItem]
    status: SafetyCheckStatus = SafetyCheckStatus.QUALIFIED


class Todo(BaseModel):
    id: Optional[int] = None
    safety_check_id: Optional[int] = None
    mine_id: int
    description: str
    responsible_person: str
    deadline: date
    priority: TodoPriority = TodoPriority.MEDIUM
    status: TodoStatus = TodoStatus.PENDING
    created_at: datetime = Field(default_factory=datetime.now)


class Statistics(BaseModel):
    month: str
    total_output: float
    completion_rate: float
    transport_count: int
    avg_transport_duration: float
    safety_check_count: int
    risk_remediation_rate: float
