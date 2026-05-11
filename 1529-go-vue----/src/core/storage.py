from __future__ import annotations

from typing import Dict, List, Optional
from uuid import UUID

from .models import (
    ComplexAccident, Todo, Well, WellPhase,
)


class InMemoryStorage:
    def __init__(self) -> None:
        self.wells: Dict[UUID, Well] = {}
        self.phases: Dict[UUID, WellPhase] = {}
        self.todos: Dict[UUID, Todo] = {}
        self.accidents: Dict[UUID, ComplexAccident] = {}

    def add_well(self, well: Well) -> Well:
        self.wells[well.id] = well
        return well

    def get_well(self, well_id: UUID) -> Optional[Well]:
        return self.wells.get(well_id)

    def list_wells(self) -> List[Well]:
        return list(self.wells.values())

    def add_phase(self, phase: WellPhase) -> WellPhase:
        self.phases[phase.id] = phase
        return phase

    def get_phase(self, phase_id: UUID) -> Optional[WellPhase]:
        return self.phases.get(phase_id)

    def list_phases_by_well(self, well_id: UUID) -> List[WellPhase]:
        return [p for p in self.phases.values() if p.well_id == well_id]

    def add_todo(self, todo: Todo) -> Todo:
        self.todos[todo.id] = todo
        return todo

    def get_todo(self, todo_id: UUID) -> Optional[Todo]:
        return self.todos.get(todo_id)

    def list_todos(self) -> List[Todo]:
        return list(self.todos.values())

    def list_todos_by_well(self, well_id: UUID) -> List[Todo]:
        return [t for t in self.todos.values() if t.well_id == well_id]

    def add_accident(self, accident: ComplexAccident) -> ComplexAccident:
        self.accidents[accident.id] = accident
        return accident

    def get_accident(self, accident_id: UUID) -> Optional[ComplexAccident]:
        return self.accidents.get(accident_id)

    def list_accidents(self) -> List[ComplexAccident]:
        return list(self.accidents.values())

    def list_accidents_by_well(self, well_id: UUID) -> List[ComplexAccident]:
        return [a for a in self.accidents.values() if a.well_id == well_id]


storage = InMemoryStorage()
