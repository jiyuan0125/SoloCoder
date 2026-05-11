from datetime import datetime
from typing import Dict, List, Optional, Type, TypeVar, Generic
from uuid import UUID

from .models import (
    Storehouse, WarehouseArea, StockInRecord, StockOutRecord,
    TemperatureRecord, PestInspectionRecord, TodoItem
)

T = TypeVar('T')


class GenericRepository(Generic[T]):
    def __init__(self):
        self._storage: Dict[UUID, T] = {}

    def add(self, item: T) -> T:
        self._storage[item.id] = item
        return item

    def get(self, item_id: UUID) -> Optional[T]:
        return self._storage.get(item_id)

    def list(self) -> List[T]:
        return list(self._storage.values())

    def delete(self, item_id: UUID) -> bool:
        if item_id in self._storage:
            del self._storage[item_id]
            return True
        return False

    def update(self, item: T) -> Optional[T]:
        if item.id in self._storage:
            self._storage[item.id] = item
            return item
        return None


class StorehouseRepository(GenericRepository[Storehouse]):
    pass


class WarehouseAreaRepository(GenericRepository[WarehouseArea]):
    def list_by_storehouse(self, storehouse_id: UUID) -> List[WarehouseArea]:
        return [a for a in self.list() if a.storehouse_id == storehouse_id]


class StockInRecordRepository(GenericRepository[StockInRecord]):
    def list_by_area(self, area_id: UUID) -> List[StockInRecord]:
        return [s for s in self.list() if s.area_id == area_id]

    def list_active(self, area_id: Optional[UUID] = None, variety: Optional[str] = None) -> List[StockInRecord]:
        result = [s for s in self.list() if s.remaining_kg > 0]
        if area_id:
            result = [s for s in result if s.area_id == area_id]
        if variety:
            result = [s for s in result if s.variety == variety]
        return result

    def list_for_fifo(self, area_id: UUID, variety: str) -> List[StockInRecord]:
        records = [s for s in self.list_active() if s.area_id == area_id and s.variety == variety]
        return sorted(records, key=lambda x: x.in_date)


class StockOutRecordRepository(GenericRepository[StockOutRecord]):
    def list_by_area(self, area_id: UUID) -> List[StockOutRecord]:
        return [s for s in self.list() if s.area_id == area_id]


class TemperatureRecordRepository(GenericRepository[TemperatureRecord]):
    def list_by_area(self, area_id: UUID) -> List[TemperatureRecord]:
        return [t for t in self.list() if t.area_id == area_id]

    def get_by_area_and_minute(self, area_id: UUID, record_time: datetime) -> Optional[TemperatureRecord]:
        target_minute = record_time.replace(second=0, microsecond=0)
        for record in self.list():
            rec_minute = record.record_time.replace(second=0, microsecond=0)
            if record.area_id == area_id and rec_minute == target_minute:
                return record
        return None


class PestInspectionRecordRepository(GenericRepository[PestInspectionRecord]):
    def list_by_area(self, area_id: UUID) -> List[PestInspectionRecord]:
        return [p for p in self.list() if p.area_id == area_id]


class TodoRepository(GenericRepository[TodoItem]):
    def list_by_area(self, area_id: UUID) -> List[TodoItem]:
        return [t for t in self.list() if t.area_id == area_id]

    def list_pending(self) -> List[TodoItem]:
        return [t for t in self.list() if not t.is_completed]


class UnitOfWork:
    def __init__(self):
        self.storehouses = StorehouseRepository()
        self.areas = WarehouseAreaRepository()
        self.stock_ins = StockInRecordRepository()
        self.stock_outs = StockOutRecordRepository()
        self.temperatures = TemperatureRecordRepository()
        self.pest_inspections = PestInspectionRecordRepository()
        self.todos = TodoRepository()


uow = UnitOfWork()
