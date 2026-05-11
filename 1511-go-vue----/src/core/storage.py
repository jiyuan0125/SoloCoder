from typing import Dict, List, Optional, TypeVar, Generic
from datetime import date, datetime

from .models import (
    Plot,
    HarvestPlan,
    HarvestRecord,
    ProcessingBatch,
    QualityRating,
    Alert,
    Todo,
)


T = TypeVar("T")


class InMemoryStore(Generic[T]):
    def __init__(self):
        self._items: Dict[str, T] = {}

    def add(self, item: T, item_id: str) -> T:
        self._items[item_id] = item
        return item

    def get(self, item_id: str) -> Optional[T]:
        return self._items.get(item_id)

    def list(self) -> List[T]:
        return list(self._items.values())

    def update(self, item_id: str, item: T) -> Optional[T]:
        if item_id in self._items:
            self._items[item_id] = item
            return item
        return None

    def delete(self, item_id: str) -> bool:
        if item_id in self._items:
            del self._items[item_id]
            return True
        return False


class Storage:
    def __init__(self):
        self.plots: InMemoryStore[Plot] = InMemoryStore()
        self.harvest_plans: InMemoryStore[HarvestPlan] = InMemoryStore()
        self.harvest_records: InMemoryStore[HarvestRecord] = InMemoryStore()
        self.processing_batches: InMemoryStore[ProcessingBatch] = InMemoryStore()
        self.quality_ratings: InMemoryStore[QualityRating] = InMemoryStore()
        self.alerts: InMemoryStore[Alert] = InMemoryStore()
        self.todos: InMemoryStore[Todo] = InMemoryStore()

    def get_plans_by_date(self, plan_date: date) -> List[HarvestPlan]:
        return [
            plan for plan in self.harvest_plans.list()
            if plan.plan_date == plan_date
        ]

    def get_records_without_batch(self) -> List[HarvestRecord]:
        batch_record_ids = {
            batch.harvest_record_id for batch in self.processing_batches.list()
        }
        return [
            record for record in self.harvest_records.list()
            if record.id not in batch_record_ids
        ]

    def get_record_by_plan(self, plan_id: str) -> Optional[HarvestRecord]:
        for record in self.harvest_records.list():
            if record.plan_id == plan_id:
                return record
        return None

    def get_batch_by_harvest_record(self, harvest_record_id: str) -> Optional[ProcessingBatch]:
        for batch in self.processing_batches.list():
            if batch.harvest_record_id == harvest_record_id:
                return batch
        return None

    def get_rating_by_batch(self, batch_id: str) -> Optional[QualityRating]:
        for rating in self.quality_ratings.list():
            if rating.batch_id == batch_id:
                return rating
        return None

    def get_todos_by_plan(self, plan_id: str) -> List[Todo]:
        return [
            todo for todo in self.todos.list()
            if todo.plan_id == plan_id
        ]

    def get_alerts_by_record(self, harvest_record_id: str) -> List[Alert]:
        return [
            alert for alert in self.alerts.list()
            if alert.harvest_record_id == harvest_record_id
        ]
