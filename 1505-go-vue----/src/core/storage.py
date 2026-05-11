from abc import ABC, abstractmethod
from datetime import datetime
from typing import Dict, List, Optional

from .models import ProductionBatch, QualityCheck, Recipe, RawMaterial, Todo


class Storage(ABC):
    @abstractmethod
    def save_recipe(self, recipe: Recipe) -> Recipe:
        ...

    @abstractmethod
    def get_recipe(self, recipe_id: str) -> Optional[Recipe]:
        ...

    @abstractmethod
    def list_recipes(self) -> List[Recipe]:
        ...

    @abstractmethod
    def save_raw_material(self, material: RawMaterial) -> RawMaterial:
        ...

    @abstractmethod
    def get_raw_material(self, material_id: str) -> Optional[RawMaterial]:
        ...

    @abstractmethod
    def get_raw_material_by_name(self, name: str) -> Optional[RawMaterial]:
        ...

    @abstractmethod
    def list_raw_materials(self) -> List[RawMaterial]:
        ...

    @abstractmethod
    def save_batch(self, batch: ProductionBatch) -> ProductionBatch:
        ...

    @abstractmethod
    def get_batch(self, batch_id: str) -> Optional[ProductionBatch]:
        ...

    @abstractmethod
    def list_batches(self, start: Optional[datetime] = None, end: Optional[datetime] = None) -> List[ProductionBatch]:
        ...

    @abstractmethod
    def add_quality_check(self, batch_id: str, qc: QualityCheck) -> Optional[ProductionBatch]:
        ...

    @abstractmethod
    def save_todo(self, todo: Todo) -> Todo:
        ...

    @abstractmethod
    def get_todo(self, todo_id: str) -> Optional[Todo]:
        ...

    @abstractmethod
    def list_todos(self) -> List[Todo]:
        ...


class InMemoryStorage(Storage):
    def __init__(self):
        self._recipes: Dict[str, Recipe] = {}
        self._raw_materials: Dict[str, RawMaterial] = {}
        self._batches: Dict[str, ProductionBatch] = {}
        self._todos: Dict[str, Todo] = {}

    def save_recipe(self, recipe: Recipe) -> Recipe:
        self._recipes[recipe.id] = recipe
        return recipe

    def get_recipe(self, recipe_id: str) -> Optional[Recipe]:
        return self._recipes.get(recipe_id)

    def list_recipes(self) -> List[Recipe]:
        return list(self._recipes.values())

    def save_raw_material(self, material: RawMaterial) -> RawMaterial:
        self._raw_materials[material.id] = material
        return material

    def get_raw_material(self, material_id: str) -> Optional[RawMaterial]:
        return self._raw_materials.get(material_id)

    def get_raw_material_by_name(self, name: str) -> Optional[RawMaterial]:
        for material in self._raw_materials.values():
            if material.name == name:
                return material
        return None

    def list_raw_materials(self) -> List[RawMaterial]:
        return list(self._raw_materials.values())

    def save_batch(self, batch: ProductionBatch) -> ProductionBatch:
        self._batches[batch.id] = batch
        for todo in batch.todos:
            self._todos[todo.id] = todo
        return batch

    def get_batch(self, batch_id: str) -> Optional[ProductionBatch]:
        return self._batches.get(batch_id)

    def list_batches(self, start: Optional[datetime] = None, end: Optional[datetime] = None) -> List[ProductionBatch]:
        batches = list(self._batches.values())
        if start:
            batches = [b for b in batches if b.created_at >= start]
        if end:
            batches = [b for b in batches if b.created_at <= end]
        return batches

    def add_quality_check(self, batch_id: str, qc: QualityCheck) -> Optional[ProductionBatch]:
        batch = self._batches.get(batch_id)
        if not batch:
            return None
        batch.quality_checks.append(qc)
        return batch

    def save_todo(self, todo: Todo) -> Todo:
        self._todos[todo.id] = todo
        batch = self._batches.get(todo.batch_id)
        if batch:
            for i, existing in enumerate(batch.todos):
                if existing.id == todo.id:
                    batch.todos[i] = todo
                    break
            else:
                batch.todos.append(todo)
        return todo

    def get_todo(self, todo_id: str) -> Optional[Todo]:
        return self._todos.get(todo_id)

    def list_todos(self) -> List[Todo]:
        return list(self._todos.values())


_storage_instance: Optional[Storage] = None


def get_storage() -> Storage:
    global _storage_instance
    if _storage_instance is None:
        _storage_instance = InMemoryStorage()
    return _storage_instance


def set_storage(storage: Storage) -> None:
    global _storage_instance
    _storage_instance = storage
