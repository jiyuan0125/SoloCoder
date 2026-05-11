from datetime import datetime
from typing import Dict, List, Optional

from .models import (
    Batch,
    BatchCreate,
    BatchSummary,
    InspectionResult,
    MaterialInventory,
    QualityInspection,
    QualityInspectionCreate,
    Recipe,
    RecipeCreate,
    Todo,
    TodoCreate,
    TodoStatus,
)


class RecipeRepository:
    def __init__(self):
        self._recipes: Dict[int, Recipe] = {}
        self._next_id: int = 1

    def create(self, data: RecipeCreate) -> Recipe:
        now = datetime.now()
        recipe = Recipe(
            id=self._next_id,
            **data.model_dump(),
            created_at=now,
        )
        self._recipes[recipe.id] = recipe
        self._next_id += 1
        return recipe

    def get_by_id(self, recipe_id: int) -> Optional[Recipe]:
        return self._recipes.get(recipe_id)

    def list_all(self) -> List[Recipe]:
        return list(self._recipes.values())

    def get_name_by_id(self, recipe_id: int) -> Optional[str]:
        recipe = self.get_by_id(recipe_id)
        return recipe.name if recipe else None


class BatchRepository:
    def __init__(self):
        self._batches: Dict[int, Batch] = {}
        self._next_id: int = 1

    def create(self, data: BatchCreate) -> Batch:
        now = datetime.now()
        batch = Batch(
            id=self._next_id,
            **data.model_dump(),
            created_at=now,
            has_passed_inspection=False,
        )
        self._batches[batch.id] = batch
        self._next_id += 1
        return batch

    def get_by_id(self, batch_id: int) -> Optional[Batch]:
        return self._batches.get(batch_id)

    def list_all(self) -> List[Batch]:
        return list(self._batches.values())

    def update(self, batch: Batch) -> Batch:
        self._batches[batch.id] = batch
        return batch

    def set_passed_inspection(self, batch_id: int) -> None:
        if batch_id in self._batches:
            self._batches[batch_id].has_passed_inspection = True

    def find_by_date_range(
        self,
        start_date: datetime,
        end_date: datetime,
    ) -> List[Batch]:
        result = []
        for batch in self._batches.values():
            if batch.production_date:
                if start_date <= batch.production_date <= end_date:
                    result.append(batch)
            else:
                if start_date <= batch.created_at <= end_date:
                    result.append(batch)
        return result


class QualityInspectionRepository:
    def __init__(self):
        self._inspections: Dict[int, QualityInspection] = {}
        self._next_id: int = 1

    def create(self, data: QualityInspectionCreate) -> QualityInspection:
        now = datetime.now()
        inspection = QualityInspection(
            id=self._next_id,
            **data.model_dump(),
            created_at=now,
        )
        self._inspections[inspection.id] = inspection
        self._next_id += 1
        return inspection

    def get_by_batch_id(self, batch_id: int) -> List[QualityInspection]:
        return [
            insp for insp in self._inspections.values()
            if insp.batch_id == batch_id
        ]

    def get_last_result(self, batch_id: int) -> Optional[InspectionResult]:
        inspections = self.get_by_batch_id(batch_id)
        if not inspections:
            return None
        inspections.sort(key=lambda x: x.created_at, reverse=True)
        return inspections[0].result


class TodoRepository:
    def __init__(self):
        self._todos: Dict[int, Todo] = {}
        self._next_id: int = 1

    def create(self, data: TodoCreate) -> Todo:
        now = datetime.now()
        todo = Todo(
            id=self._next_id,
            **data.model_dump(),
            created_at=now,
            updated_at=now,
        )
        self._todos[todo.id] = todo
        self._next_id += 1
        return todo

    def get_by_id(self, todo_id: int) -> Optional[Todo]:
        return self._todos.get(todo_id)

    def get_by_batch_id(self, batch_id: int) -> Optional[Todo]:
        for todo in self._todos.values():
            if todo.batch_id == batch_id:
                return todo
        return None

    def list_all(self) -> List[Todo]:
        return list(self._todos.values())

    def update_status(
        self,
        todo_id: int,
        status: TodoStatus,
        notes: Optional[str] = None,
    ) -> Optional[Todo]:
        if todo_id not in self._todos:
            return None
        todo = self._todos[todo_id]
        todo.status = status
        if notes:
            todo.notes = notes
        todo.updated_at = datetime.now()
        return todo


class MaterialInventoryRepository:
    def __init__(self):
        self._materials: Dict[int, MaterialInventory] = {}
        self._next_id: int = 1

    def create(
        self,
        name: str,
        current_stock: float,
        safety_stock: float,
    ) -> MaterialInventory:
        material = MaterialInventory(
            id=self._next_id,
            name=name,
            current_stock=current_stock,
            safety_stock=safety_stock,
        )
        self._materials[material.id] = material
        self._next_id += 1
        return material

    def get_by_id(self, material_id: int) -> Optional[MaterialInventory]:
        return self._materials.get(material_id)

    def get_by_name(self, name: str) -> Optional[MaterialInventory]:
        for m in self._materials.values():
            if m.name == name:
                return m
        return None

    def list_all(self) -> List[MaterialInventory]:
        return list(self._materials.values())

    def update_stock(self, material_id: int, current_stock: float) -> Optional[MaterialInventory]:
        if material_id not in self._materials:
            return None
        self._materials[material_id].current_stock = current_stock
        return self._materials[material_id]

    def get_low_stock(self) -> List[MaterialInventory]:
        return [
            m for m in self._materials.values()
            if m.current_stock < m.safety_stock
        ]


class SummaryRepository:
    def __init__(
        self,
        batch_repo: BatchRepository,
        recipe_repo: RecipeRepository,
        inspection_repo: QualityInspectionRepository,
        todo_repo: TodoRepository,
    ):
        self._batch_repo = batch_repo
        self._recipe_repo = recipe_repo
        self._inspection_repo = inspection_repo
        self._todo_repo = todo_repo

    def get_batch_summaries(
        self,
        start_date: datetime,
        end_date: datetime,
    ) -> List[BatchSummary]:
        batches = self._batch_repo.find_by_date_range(start_date, end_date)
        summaries = []
        for batch in batches:
            recipe = self._recipe_repo.get_by_id(batch.recipe_id)
            inspection_result = self._inspection_repo.get_last_result(batch.id)
            todo = self._todo_repo.get_by_batch_id(batch.id)
            
            summary = BatchSummary(
                batch_id=batch.id,
                recipe_name=recipe.name if recipe else "未知配方",
                product_category=recipe.product_category if recipe else "",
                planned_quantity=batch.planned_quantity,
                actual_quantity=batch.actual_quantity,
                production_date=batch.production_date or batch.created_at,
                inspection_result=inspection_result.value if inspection_result else None,
                todo_status=todo.status.value if todo else None,
            )
            summaries.append(summary)
        return summaries
