from typing import List, Optional, Dict
from collections import defaultdict
import uuid
from datetime import datetime

from core.models import (
    SupplyChainRecord,
    InspectionRecord,
    Recall,
    TodoItem,
    RecallStatus,
    TodoStatus
)


class InMemoryStorage:
    def __init__(self):
        self._supply_chain: Dict[str, SupplyChainRecord] = {}
        self._supply_chain_by_batch: Dict[str, List[str]] = defaultdict(list)
        
        self._inspections: Dict[str, InspectionRecord] = {}
        self._inspections_by_batch: Dict[str, List[str]] = defaultdict(list)
        
        self._recalls: Dict[str, Recall] = {}
        self._recalls_by_batch: Dict[str, List[str]] = defaultdict(list)
        self._active_recalls_by_batch: Dict[str, str] = {}
        
        self._todos: Dict[str, TodoItem] = {}
        self._todos_by_recall: Dict[str, List[str]] = defaultdict(list)

    def add_supply_chain_record(self, record: SupplyChainRecord) -> SupplyChainRecord:
        if record.id is None:
            record.id = str(uuid.uuid4())
        self._supply_chain[record.id] = record
        self._supply_chain_by_batch[record.batch_number].append(record.id)
        return record

    def get_supply_chain_by_batch(self, batch_number: str) -> List[SupplyChainRecord]:
        record_ids = self._supply_chain_by_batch.get(batch_number, [])
        records = [self._supply_chain[rid] for rid in record_ids]
        return sorted(records, key=lambda r: r.operation_time)

    def get_all_supply_chain(self) -> List[SupplyChainRecord]:
        return list(self._supply_chain.values())

    def add_inspection_record(self, record: InspectionRecord) -> InspectionRecord:
        if record.id is None:
            record.id = str(uuid.uuid4())
        self._inspections[record.id] = record
        self._inspections_by_batch[record.batch_number].append(record.id)
        return record

    def get_inspections_by_batch(self, batch_number: str) -> List[InspectionRecord]:
        record_ids = self._inspections_by_batch.get(batch_number, [])
        return [self._inspections[rid] for rid in record_ids]

    def get_all_inspections(self) -> List[InspectionRecord]:
        return list(self._inspections.values())

    def has_active_recall(self, batch_number: str) -> bool:
        return batch_number in self._active_recalls_by_batch

    def add_recall(self, recall: Recall) -> Recall:
        if recall.id is None:
            recall.id = str(uuid.uuid4())
        self._recalls[recall.id] = recall
        self._recalls_by_batch[recall.batch_number].append(recall.id)
        if recall.status == RecallStatus.ACTIVE:
            self._active_recalls_by_batch[recall.batch_number] = recall.id
        return recall

    def update_recall_status(self, recall_id: str, status: 'RecallStatus', 
                           completed_by: Optional[str] = None) -> Optional[Recall]:
        recall = self._recalls.get(recall_id)
        if recall is None:
            return None
        recall.status = status
        if status != RecallStatus.ACTIVE:
            if self._active_recalls_by_batch.get(recall.batch_number) == recall_id:
                del self._active_recalls_by_batch[recall.batch_number]
        if status == RecallStatus.COMPLETED:
            recall.completed_at = datetime.now()
            recall.completed_by = completed_by
        return recall

    def get_recall(self, recall_id: str) -> Optional[Recall]:
        return self._recalls.get(recall_id)

    def get_recalls_by_batch(self, batch_number: str) -> List[Recall]:
        recall_ids = self._recalls_by_batch.get(batch_number, [])
        return [self._recalls[rid] for rid in recall_ids]

    def get_all_recalls(self) -> List[Recall]:
        return list(self._recalls.values())

    def add_todo(self, todo: TodoItem) -> TodoItem:
        if todo.id is None:
            todo.id = str(uuid.uuid4())
        self._todos[todo.id] = todo
        self._todos_by_recall[todo.recall_id].append(todo.id)
        return todo

    def get_todo(self, todo_id: str) -> Optional[TodoItem]:
        return self._todos.get(todo_id)

    def get_todos_by_recall(self, recall_id: str) -> List[TodoItem]:
        todo_ids = self._todos_by_recall.get(recall_id, [])
        return [self._todos[tid] for tid in todo_ids]

    def get_all_todos(self) -> List[TodoItem]:
        return list(self._todos.values())

    def update_todo_status(self, todo_id: str, status: 'TodoStatus', 
                          confirmed_by: Optional[str] = None,
                          remarks: Optional[str] = None) -> Optional[TodoItem]:
        todo = self._todos.get(todo_id)
        if todo is None:
            return None
        todo.status = status
        if status == TodoStatus.CONFIRMED:
            todo.confirmed_at = datetime.now()
            todo.confirmed_by = confirmed_by
        if remarks is not None:
            todo.remarks = remarks
        return todo
