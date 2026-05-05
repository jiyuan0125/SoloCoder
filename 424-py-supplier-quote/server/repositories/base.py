from abc import ABC, abstractmethod
from typing import Any, Generic, Optional, TypeVar
from uuid import UUID

from shared.models.base import BaseModel, UUIDMixin

T = TypeVar("T", bound=UUIDMixin)


class BaseRepository(Generic[T], ABC):
    def __init__(self) -> None:
        self._storage: dict[UUID, T] = {}

    def create(self, entity: T) -> T:
        if entity.id in self._storage:
            raise ValueError(f"Entity with id {entity.id} already exists")
        self._storage[entity.id] = entity
        return entity

    def get_by_id(self, entity_id: UUID) -> Optional[T]:
        return self._storage.get(entity_id)

    def update(self, entity: T) -> T:
        if entity.id not in self._storage:
            raise ValueError(f"Entity with id {entity.id} does not exist")
        self._storage[entity.id] = entity
        return entity

    def delete(self, entity_id: UUID) -> bool:
        if entity_id in self._storage:
            del self._storage[entity_id]
            return True
        return False

    def list_all(self) -> list[T]:
        return list(self._storage.values())

    def count(self) -> int:
        return len(self._storage)

    def find(self, predicate: Any) -> list[T]:
        return [entity for entity in self._storage.values() if predicate(entity)]

    def find_one(self, predicate: Any) -> Optional[T]:
        for entity in self._storage.values():
            if predicate(entity):
                return entity
        return None
