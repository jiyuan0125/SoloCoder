from __future__ import annotations

from datetime import date, datetime
from typing import List, Optional
from uuid import UUID

from pydantic import BaseModel

from ..core.models import (
    AccidentType, PhaseName, PhaseStatus, TodoStatus, TodoType, WellStatus,
)


class WellCreate(BaseModel):
    name: str
    block: str
    planned_total_days: int


class WellResponse(BaseModel):
    id: UUID
    name: str
    block: str
    status: WellStatus
    planned_total_days: int
    created_at: datetime


class PhaseStart(BaseModel):
    phase_name: PhaseName
    start_date: date
    planned_days: int


class PhaseComplete(BaseModel):
    end_date: date


class PhaseResponse(BaseModel):
    id: UUID
    well_id: UUID
    phase_name: PhaseName
    status: PhaseStatus
    planned_days: int
    actual_days: Optional[int]
    start_date: Optional[date]
    end_date: Optional[date]


class TodoResponse(BaseModel):
    id: UUID
    well_id: UUID
    phase_id: Optional[UUID]
    todo_type: TodoType
    status: TodoStatus
    description: str
    created_at: datetime


class TodoAccept(BaseModel):
    pass


class AccidentCreate(BaseModel):
    accident_type: AccidentType
    start_time: datetime
    end_time: Optional[datetime] = None
    has_loss: bool = False
    responsible_person: Optional[str] = None


class AccidentUpdate(BaseModel):
    cause_analysis: Optional[str] = None
    preventive_measures: Optional[str] = None
    end_time: Optional[datetime] = None
    has_loss: Optional[bool] = None
    responsible_person: Optional[str] = None


class AccidentResponse(BaseModel):
    id: UUID
    well_id: UUID
    accident_type: AccidentType
    start_time: datetime
    end_time: Optional[datetime]
    has_loss: bool
    cause_analysis: Optional[str]
    preventive_measures: Optional[str]
    responsible_person: Optional[str]
    closed: bool


class BlockStatsResponse(BaseModel):
    block: str
    in_progress_count: int
    completed_count: int
    avg_drilling_cycle_days: float
    complex_accident_rate: float


class BlocksResponse(BaseModel):
    blocks: List[str]


class ErrorResponse(BaseModel):
    detail: str
