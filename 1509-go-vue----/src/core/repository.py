from abc import ABC, abstractmethod
from datetime import datetime
from typing import List, Optional, Dict
import uuid

from .models import (
    SupplyChainRecord,
    InspectionRecord,
    RecallRecord,
    TodoItem,
    SupplyChainCreate,
    InspectionCreate,
    RecallCreate,
    TodoUpdate,
    RecallStatus,
    TodoStatus,
)


class Repository(ABC):
    @abstractmethod
    def create_supply_chain(self, data: SupplyChainCreate) -> SupplyChainRecord:
        pass

    @abstractmethod
    def get_supply_chain_by_batch(self, batch_number: str) -> List[SupplyChainRecord]:
        pass

    @abstractmethod
    def get_all_supply_chain(self) -> List[SupplyChainRecord]:
        pass

    @abstractmethod
    def create_inspection(self, data: InspectionCreate) -> InspectionRecord:
        pass

    @abstractmethod
    def get_inspection_by_batch(self, batch_number: str) -> List[InspectionRecord]:
        pass

    @abstractmethod
    def get_all_inspections(self) -> List[InspectionRecord]:
        pass

    @abstractmethod
    def create_recall(self, data: RecallCreate, can_track: bool, affected_stages: List[str]) -> RecallRecord:
        pass

    @abstractmethod
    def get_active_recall_by_batch(self, batch_number: str) -> Optional[RecallRecord]:
        pass

    @abstractmethod
    def get_recall_by_id(self, recall_id: str) -> Optional[RecallRecord]:
        pass

    @abstractmethod
    def get_all_recalls(self) -> List[RecallRecord]:
        pass

    @abstractmethod
    def update_recall_status(self, recall_id: str, status: RecallStatus) -> Optional[RecallRecord]:
        pass

    @abstractmethod
    def create_todo(self, recall_id: str, batch_number: str, stage: str, handler: str) -> TodoItem:
        pass

    @abstractmethod
    def get_todos_by_recall(self, recall_id: str) -> List[TodoItem]:
        pass

    @abstractmethod
    def get_todo_by_id(self, todo_id: str) -> Optional[TodoItem]:
        pass

    @abstractmethod
    def update_todo(self, todo_id: str, data: TodoUpdate) -> Optional[TodoItem]:
        pass

    @abstractmethod
    def get_all_todos(self) -> List[TodoItem]:
        pass


class InMemoryRepository(Repository):
    def __init__(self):
        self._supply_chain: Dict[str, SupplyChainRecord] = {}
        self._inspections: Dict[str, InspectionRecord] = {}
        self._recalls: Dict[str, RecallRecord] = {}
        self._todos: Dict[str, TodoItem] = {}

    def create_supply_chain(self, data: SupplyChainCreate) -> SupplyChainRecord:
        now = datetime.now()
        record = SupplyChainRecord(
            id=str(uuid.uuid4()),
            **data.dict(),
            created_at=now,
        )
        self._supply_chain[record.id] = record
        return record

    def get_supply_chain_by_batch(self, batch_number: str) -> List[SupplyChainRecord]:
        records = [
            r for r in self._supply_chain.values()
            if r.batch_number == batch_number
        ]
        return sorted(records, key=lambda r: r.operation_time)

    def get_all_supply_chain(self) -> List[SupplyChainRecord]:
        return sorted(self._supply_chain.values(), key=lambda r: r.created_at)

    def create_inspection(self, data: InspectionCreate) -> InspectionRecord:
        now = datetime.now()
        record = InspectionRecord(
            id=str(uuid.uuid4()),
            **data.dict(),
            created_at=now,
        )
        self._inspections[record.id] = record
        return record

    def get_inspection_by_batch(self, batch_number: str) -> List[InspectionRecord]:
        records = [
            r for r in self._inspections.values()
            if r.batch_number == batch_number
        ]
        return sorted(records, key=lambda r: r.inspection_time)

    def get_all_inspections(self) -> List[InspectionRecord]:
        return sorted(self._inspections.values(), key=lambda r: r.created_at)

    def create_recall(self, data: RecallCreate, can_track: bool, affected_stages: List[str]) -> RecallRecord:
        now = datetime.now()
        record = RecallRecord(
            id=str(uuid.uuid4()),
            batch_number=data.batch_number,
            reason=data.reason,
            initiator=data.initiator,
            status=RecallStatus.ACTIVE,
            start_time=now,
            can_track=can_track,
            affected_stages=affected_stages,
            created_at=now,
        )
        self._recalls[record.id] = record
        return record

    def get_active_recall_by_batch(self, batch_number: str) -> Optional[RecallRecord]:
        for r in self._recalls.values():
            if r.batch_number == batch_number and r.status == RecallStatus.ACTIVE:
                return r
        return None

    def get_recall_by_id(self, recall_id: str) -> Optional[RecallRecord]:
        return self._recalls.get(recall_id)

    def get_all_recalls(self) -> List[RecallRecord]:
        return sorted(self._recalls.values(), key=lambda r: r.created_at, reverse=True)

    def update_recall_status(self, recall_id: str, status: RecallStatus) -> Optional[RecallRecord]:
        if recall_id not in self._recalls:
            return None
        recall = self._recalls[recall_id]
        recall.status = status
        if status == RecallStatus.COMPLETED:
            recall.end_time = datetime.now()
        self._recalls[recall_id] = recall
        return recall

    def create_todo(self, recall_id: str, batch_number: str, stage: str, handler: str) -> TodoItem:
        now = datetime.now()
        todo = TodoItem(
            id=str(uuid.uuid4()),
            recall_id=recall_id,
            batch_number=batch_number,
            stage=stage,
            handler=handler,
            status="pending",
            created_at=now,
        )
        self._todos[todo.id] = todo
        return todo

    def get_todos_by_recall(self, recall_id: str) -> List[TodoItem]:
        return [
            t for t in self._todos.values()
            if t.recall_id == recall_id
        ]

    def get_todo_by_id(self, todo_id: str) -> Optional[TodoItem]:
        return self._todos.get(todo_id)

    def update_todo(self, todo_id: str, data: TodoUpdate) -> Optional[TodoItem]:
        if todo_id not in self._todos:
            return None
        todo = self._todos[todo_id]
        todo.status = data.status
        if data.status == TodoStatus.COMPLETED:
            todo.processed_at = datetime.now()
            todo.processed_by = data.processed_by
            todo.notes = data.notes
        elif data.notes:
            todo.notes = data.notes
        self._todos[todo_id] = todo
        return todo

    def get_all_todos(self) -> List[TodoItem]:
        return sorted(self._todos.values(), key=lambda t: t.created_at, reverse=True)
