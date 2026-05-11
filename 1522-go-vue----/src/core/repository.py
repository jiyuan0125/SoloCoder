from typing import Dict, List, Optional, TypeVar, Generic
from abc import ABC, abstractmethod
from .models import Project, Borehole, Sample, TodoItem


T = TypeVar('T')


class BaseRepository(ABC, Generic[T]):
    @abstractmethod
    def create(self, entity: T) -> T:
        pass

    @abstractmethod
    def get_by_id(self, entity_id: str) -> Optional[T]:
        pass

    @abstractmethod
    def update(self, entity_id: str, entity: T) -> T:
        pass

    @abstractmethod
    def delete(self, entity_id: str) -> bool:
        pass

    @abstractmethod
    def list_all(self) -> List[T]:
        pass


class InMemoryRepository(BaseRepository[T]):
    def __init__(self):
        self._storage: Dict[str, T] = {}
        self._next_id = 1

    def _generate_id(self) -> str:
        entity_id = str(self._next_id)
        self._next_id += 1
        return entity_id

    def create(self, entity: T) -> T:
        entity_id = self._generate_id()
        entity_dict = entity.model_dump()
        entity_dict['id'] = entity_id
        entity = entity.__class__(**entity_dict)
        self._storage[entity_id] = entity
        return entity

    def get_by_id(self, entity_id: str) -> Optional[T]:
        return self._storage.get(entity_id)

    def update(self, entity_id: str, entity: T) -> T:
        self._storage[entity_id] = entity
        return entity

    def delete(self, entity_id: str) -> bool:
        if entity_id in self._storage:
            del self._storage[entity_id]
            return True
        return False

    def list_all(self) -> List[T]:
        return list(self._storage.values())


class ProjectRepository(InMemoryRepository[Project]):
    pass


class BoreholeRepository(InMemoryRepository[Borehole]):
    def list_by_project(self, project_id: str) -> List[Borehole]:
        return [
            borehole for borehole in self.list_all()
            if borehole.project_id == project_id
        ]

    def get_incomplete_by_project(self, project_id: str) -> List[Borehole]:
        return [
            borehole for borehole in self.list_all()
            if borehole.project_id == project_id and not borehole.is_completed
        ]


class SampleRepository(InMemoryRepository[Sample]):
    def list_by_borehole(self, borehole_id: str) -> List[Sample]:
        return [
            sample for sample in self.list_all()
            if sample.borehole_id == borehole_id
        ]

    def list_by_borehole_sorted(self, borehole_id: str) -> List[Sample]:
        samples = self.list_by_borehole(borehole_id)
        return sorted(samples, key=lambda s: s.start_depth)


class TodoRepository(InMemoryRepository[TodoItem]):
    def list_by_project(self, project_id: str) -> List[TodoItem]:
        return [
            todo for todo in self.list_all()
            if todo.project_id == project_id
        ]

    def list_incomplete(self) -> List[TodoItem]:
        return [
            todo for todo in self.list_all()
            if not todo.is_completed
        ]
