from abc import ABC, abstractmethod
from typing import List, Optional, Type, TypeVar, Dict, Any
from datetime import datetime, date
import uuid

from .models import (
    HazardWaste,
    WasteProducer,
    DisposalCompany,
    Ledger,
    TransferDocument,
    DisposalRecord,
    Alert,
    LedgerSummary,
    ExceptionRecord
)


T = TypeVar('T')


class Database(ABC):
    @abstractmethod
    def create(self, obj: T) -> T:
        pass

    @abstractmethod
    def get(self, model: Type[T], obj_id: str) -> Optional[T]:
        pass

    @abstractmethod
    def update(self, obj: T) -> T:
        pass

    @abstractmethod
    def list(self, model: Type[T], filters: Optional[Dict[str, Any]] = None) -> List[T]:
        pass

    @abstractmethod
    def delete(self, model: Type[T], obj_id: str) -> bool:
        pass


class InMemoryDatabase(Database):
    def __init__(self):
        self._storage: Dict[Type, Dict[str, Any]] = {
            WasteProducer: {},
            DisposalCompany: {},
            HazardWaste: {},
            Ledger: {},
            TransferDocument: {},
            DisposalRecord: {},
            Alert: {},
            LedgerSummary: {},
            ExceptionRecord: {}
        }

    def create(self, obj: T) -> T:
        model_type = type(obj)
        if model_type not in self._storage:
            self._storage[model_type] = {}
        self._storage[model_type][obj.id] = obj
        return obj

    def get(self, model: Type[T], obj_id: str) -> Optional[T]:
        return self._storage.get(model, {}).get(obj_id)

    def update(self, obj: T) -> T:
        model_type = type(obj)
        if model_type in self._storage and obj.id in self._storage[model_type]:
            self._storage[model_type][obj.id] = obj
            return obj
        raise ValueError(f"Object with id {obj.id} not found")

    def list(self, model: Type[T], filters: Optional[Dict[str, Any]] = None) -> List[T]:
        all_items = list(self._storage.get(model, {}).values())
        if not filters:
            return all_items
        return [
            item for item in all_items
            if all(getattr(item, k, None) == v for k, v in filters.items())
        ]

    def delete(self, model: Type[T], obj_id: str) -> bool:
        if model in self._storage and obj_id in self._storage[model]:
            del self._storage[model][obj_id]
            return True
        return False


_db: Optional[InMemoryDatabase] = None


def get_db() -> InMemoryDatabase:
    global _db
    if _db is None:
        _db = InMemoryDatabase()
    return _db
