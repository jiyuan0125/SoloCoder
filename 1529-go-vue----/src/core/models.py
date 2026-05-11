from __future__ import annotations

from dataclasses import dataclass, field
from datetime import date, datetime
from enum import Enum
from typing import List, Optional
from uuid import UUID, uuid4


class WellStatus(str, Enum):
    NOT_STARTED = 'not_started'
    IN_PROGRESS = 'in_progress'
    COMPLETED = 'completed'


class PhaseName(str, Enum):
    PHASE_1 = 'PHASE_1'
    PHASE_2 = 'PHASE_2'
    PHASE_3 = 'PHASE_3'
    PHASE_4 = 'PHASE_4'
    PHASE_5 = 'PHASE_5'
    PHASE_6 = 'PHASE_6'
    PHASE_7 = 'PHASE_7'
    PHASE_8 = 'PHASE_8'


PHASE_ORDER: List[PhaseName] = [
    PhaseName.PHASE_1,
    PhaseName.PHASE_2,
    PhaseName.PHASE_3,
    PhaseName.PHASE_4,
    PhaseName.PHASE_5,
    PhaseName.PHASE_6,
    PhaseName.PHASE_7,
    PhaseName.PHASE_8,
]


class PhaseStatus(str, Enum):
    PENDING = 'pending'
    IN_PROGRESS = 'in_progress'
    COMPLETED = 'completed'
    ACCEPTED = 'accepted'


class TodoType(str, Enum):
    ACCEPTANCE = 'acceptance'
    ANOMALY = 'anomaly'


class TodoStatus(str, Enum):
    PENDING = 'pending'
    COMPLETED = 'completed'


class AccidentType(str, Enum):
    WELL_KICK = 'well_kick'
    WELL_LOSS = 'well_loss'
    STUCK_PIPE = 'stuck_pipe'
    WELL_COLLAPSE = 'well_collapse'


@dataclass
class Well:
    id: UUID = field(default_factory=uuid4)
    name: str = ''
    block: str = ''
    status: WellStatus = WellStatus.NOT_STARTED
    planned_total_days: int = 0
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class WellPhase:
    id: UUID = field(default_factory=uuid4)
    well_id: UUID = field(default_factory=uuid4)
    phase_name: PhaseName = PhaseName.PHASE_1
    status: PhaseStatus = PhaseStatus.PENDING
    planned_days: int = 0
    actual_days: Optional[int] = None
    start_date: Optional[date] = None
    end_date: Optional[date] = None


@dataclass
class Todo:
    id: UUID = field(default_factory=uuid4)
    well_id: UUID = field(default_factory=uuid4)
    phase_id: Optional[UUID] = None
    todo_type: TodoType = TodoType.ACCEPTANCE
    status: TodoStatus = TodoStatus.PENDING
    description: str = ''
    created_at: datetime = field(default_factory=datetime.utcnow)


@dataclass
class ComplexAccident:
    id: UUID = field(default_factory=uuid4)
    well_id: UUID = field(default_factory=uuid4)
    accident_type: AccidentType = AccidentType.WELL_KICK
    start_time: datetime = field(default_factory=datetime.utcnow)
    end_time: Optional[datetime] = None
    has_loss: bool = False
    cause_analysis: Optional[str] = None
    preventive_measures: Optional[str] = None
    responsible_person: Optional[str] = None
    closed: bool = False


@dataclass
class BlockStats:
    block: str = ''
    in_progress_count: int = 0
    completed_count: int = 0
    avg_drilling_cycle_days: float = 0.0
    complex_accident_rate: float = 0.0
