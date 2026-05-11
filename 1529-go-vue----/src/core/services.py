from __future__ import annotations

from datetime import date, datetime
from typing import List, Optional, Set
from uuid import UUID

from .models import (
    BlockStats, ComplexAccident, PHASE_ORDER,
    PhaseStatus, PhaseName, Todo, TodoStatus, TodoType,
    Well, WellPhase, WellStatus, AccidentType,
)
from .storage import storage


ACCIDENT_TYPE_MINOR_HOURS = 4


class WellService:
    def __init__(self) -> None:
        self.storage = storage

    def create_well(self, name: str, block: str, planned_total_days: int) -> Well:
        well = Well(name=name, block=block, planned_total_days=planned_total_days)
        self.storage.add_well(well)
        self._init_phases(well.id)
        return well

    def _init_phases(self, well_id: UUID) -> None:
        for phase_name in PHASE_ORDER:
            phase = WellPhase(
                well_id=well_id,
                phase_name=phase_name,
                status=PhaseStatus.PENDING,
                planned_days=0,
            )
            self.storage.add_phase(phase)

    def start_phase(self, well_id: UUID, phase_name: PhaseName, start_date: date, planned_days: int) -> WellPhase:
        well = self.storage.get_well(well_id)
        if well is None:
            raise ValueError(f'Well {well_id} not found')

        phases = self.storage.list_phases_by_well(well_id)
        in_progress = [p for p in phases if p.status == PhaseStatus.IN_PROGRESS]
        if in_progress:
            raise ValueError(f'Well {well_id} already has a phase in progress')

        current_idx = PHASE_ORDER.index(phase_name)
        if current_idx > 0:
            prev_phase_name = PHASE_ORDER[current_idx - 1]
            prev_phase = next((p for p in phases if p.phase_name == prev_phase_name), None)
            if prev_phase is None or prev_phase.status != PhaseStatus.ACCEPTED:
                raise ValueError(f'Previous phase {prev_phase_name.value} must be accepted first')

        phase = next((p for p in phases if p.phase_name == phase_name), None)
        if phase is None:
            raise ValueError(f'Phase {phase_name.value} not found for well {well_id}')

        phase.status = PhaseStatus.IN_PROGRESS
        phase.start_date = start_date
        phase.planned_days = planned_days

        if well.status == WellStatus.NOT_STARTED:
            well.status = WellStatus.IN_PROGRESS

        return phase

    def complete_phase(self, phase_id: UUID, end_date: date) -> WellPhase:
        phase = self.storage.get_phase(phase_id)
        if phase is None:
            raise ValueError(f'Phase {phase_id} not found')

        if phase.status != PhaseStatus.IN_PROGRESS:
            raise ValueError(f'Phase {phase_id} is not in progress')

        if phase.start_date is None:
            raise ValueError(f'Phase {phase_id} has no start date')

        actual_days = (end_date - phase.start_date).days + 1
        if actual_days < 1:
            actual_days = 1

        phase.status = PhaseStatus.COMPLETED
        phase.end_date = end_date
        phase.actual_days = actual_days

        acceptance_todo = Todo(
            well_id=phase.well_id,
            phase_id=phase.id,
            todo_type=TodoType.ACCEPTANCE,
            status=TodoStatus.PENDING,
            description=f'Acceptance for phase {phase.phase_name.value}',
        )
        self.storage.add_todo(acceptance_todo)

        if phase.planned_days > 0:
            threshold = phase.planned_days * 1.2
            if actual_days > threshold:
                anomaly_todo = Todo(
                    well_id=phase.well_id,
                    phase_id=phase.id,
                    todo_type=TodoType.ANOMALY,
                    status=TodoStatus.PENDING,
                    description=f'Anomaly: actual {actual_days} days > planned {phase.planned_days} days * 120%',
                )
                self.storage.add_todo(anomaly_todo)

        return phase

    def accept_phase(self, todo_id: UUID) -> WellPhase:
        todo = self.storage.get_todo(todo_id)
        if todo is None:
            raise ValueError(f'Todo {todo_id} not found')

        if todo.todo_type != TodoType.ACCEPTANCE:
            raise ValueError(f'Todo {todo_id} is not an acceptance todo')

        if todo.phase_id is None:
            raise ValueError(f'Todo {todo_id} has no phase')

        phase = self.storage.get_phase(todo.phase_id)
        if phase is None:
            raise ValueError(f'Phase {todo.phase_id} not found')

        todo.status = TodoStatus.COMPLETED
        phase.status = PhaseStatus.ACCEPTED

        phases = self.storage.list_phases_by_well(phase.well_id)
        all_accepted = all(p.status == PhaseStatus.ACCEPTED for p in phases)
        if all_accepted:
            well = self.storage.get_well(phase.well_id)
            if well is not None:
                well.status = WellStatus.COMPLETED

        return phase

    def get_well(self, well_id: UUID) -> Optional[Well]:
        return self.storage.get_well(well_id)

    def list_wells(self) -> List[Well]:
        return self.storage.list_wells()

    def get_phases(self, well_id: UUID) -> List[WellPhase]:
        return self.storage.list_phases_by_well(well_id)

    def list_todos(self, well_id: Optional[UUID] = None) -> List[Todo]:
        if well_id is not None:
            return self.storage.list_todos_by_well(well_id)
        return self.storage.list_todos()

    def get_todo(self, todo_id: UUID) -> Optional[Todo]:
        return self.storage.get_todo(todo_id)


class AccidentService:
    def __init__(self) -> None:
        self.storage = storage

    def record_accident(
        self,
        well_id: UUID,
        accident_type: AccidentType,
        start_time: datetime,
        end_time: Optional[datetime] = None,
        has_loss: bool = False,
        responsible_person: Optional[str] = None,
    ) -> ComplexAccident:
        accident = ComplexAccident(
            well_id=well_id,
            accident_type=accident_type,
            start_time=start_time,
            end_time=end_time,
            has_loss=has_loss,
            responsible_person=responsible_person,
        )
        self.storage.add_accident(accident)
        return accident

    def update_accident(
        self,
        accident_id: UUID,
        cause_analysis: Optional[str] = None,
        preventive_measures: Optional[str] = None,
        end_time: Optional[datetime] = None,
        has_loss: Optional[bool] = None,
        responsible_person: Optional[str] = None,
    ) -> ComplexAccident:
        accident = self.storage.get_accident(accident_id)
        if accident is None:
            raise ValueError(f'Accident {accident_id} not found')

        if cause_analysis is not None:
            accident.cause_analysis = cause_analysis
        if preventive_measures is not None:
            accident.preventive_measures = preventive_measures
        if end_time is not None:
            accident.end_time = end_time
        if has_loss is not None:
            accident.has_loss = has_loss
        if responsible_person is not None:
            accident.responsible_person = responsible_person

        if accident.cause_analysis and accident.preventive_measures:
            accident.closed = True

        return accident

    def get_accident(self, accident_id: UUID) -> Optional[ComplexAccident]:
        return self.storage.get_accident(accident_id)

    def list_accidents(self, well_id: Optional[UUID] = None) -> List[ComplexAccident]:
        if well_id is not None:
            return self.storage.list_accidents_by_well(well_id)
        return self.storage.list_accidents()

    def is_minor_accident(self, accident: ComplexAccident) -> bool:
        if accident.has_loss:
            return False
        if accident.end_time is None:
            return False
        duration_hours = (accident.end_time - accident.start_time).total_seconds() / 3600
        return duration_hours < ACCIDENT_TYPE_MINOR_HOURS


class StatsService:
    def __init__(self) -> None:
        self.storage = storage
        self.well_service = WellService()
        self.accident_service = AccidentService()

    def _get_well_cycle_days(self, well: Well) -> Optional[int]:
        phases = self.storage.list_phases_by_well(well.id)
        if not phases:
            return None
        actual_days_list = [p.actual_days for p in phases if p.actual_days is not None]
        if not actual_days_list:
            return None
        return sum(actual_days_list)

    def _is_complex_accident_well(self, well_id: UUID) -> bool:
        accidents = self.storage.list_accidents_by_well(well_id)
        for acc in accidents:
            if not self.accident_service.is_minor_accident(acc):
                return True
        return False

    def get_block_stats(self, block: str) -> BlockStats:
        wells = [w for w in self.storage.list_wells() if w.block == block]

        in_progress = [w for w in wells if w.status == WellStatus.IN_PROGRESS]
        completed = [w for w in wells if w.status == WellStatus.COMPLETED]

        total_completed_cycles: List[int] = []
        for w in completed:
            cycle = self._get_well_cycle_days(w)
            if cycle is not None:
                total_completed_cycles.append(cycle)

        avg_cycle = sum(total_completed_cycles) / len(total_completed_cycles) if total_completed_cycles else 0.0

        well_with_complex_accident: Set[UUID] = set()
        for w in wells:
            if self._is_complex_accident_well(w.id):
                well_with_complex_accident.add(w.id)

        total_wells = len(wells)
        accident_rate = len(well_with_complex_accident) / total_wells if total_wells > 0 else 0.0

        return BlockStats(
            block=block,
            in_progress_count=len(in_progress),
            completed_count=len(completed),
            avg_drilling_cycle_days=avg_cycle,
            complex_accident_rate=accident_rate,
        )

    def list_blocks(self) -> List[str]:
        blocks: Set[str] = set()
        for well in self.storage.list_wells():
            if well.block:
                blocks.add(well.block)
        return sorted(list(blocks))


well_service = WellService()
accident_service = AccidentService()
stats_service = StatsService()
