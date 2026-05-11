from datetime import datetime
from typing import List, Optional

from .models import (
    Batch,
    BatchCreate,
    InspectionResult,
    MaterialInventory,
    QualityInspection,
    QualityInspectionCreate,
    Recipe,
    RecipeCreate,
    Todo,
    TodoCreate,
    TodoStatus,
    ValidationError,
)
from .repositories import (
    BatchRepository,
    MaterialInventoryRepository,
    QualityInspectionRepository,
    RecipeRepository,
    SummaryRepository,
    TodoRepository,
)
from .validators import (
    validate_batch,
    validate_batch_actual_quantity,
    validate_quality_inspection,
    validate_recipe,
)


class ProductionService:
    def __init__(self):
        self.recipe_repo = RecipeRepository()
        self.batch_repo = BatchRepository()
        self.inspection_repo = QualityInspectionRepository()
        self.todo_repo = TodoRepository()
        self.material_repo = MaterialInventoryRepository()
        self.summary_repo = SummaryRepository(
            batch_repo=self.batch_repo,
            recipe_repo=self.recipe_repo,
            inspection_repo=self.inspection_repo,
            todo_repo=self.todo_repo,
        )

    def create_recipe(self, data: RecipeCreate) -> Recipe:
        validate_recipe(data)
        return self.recipe_repo.create(data)

    def get_recipe(self, recipe_id: int) -> Optional[Recipe]:
        return self.recipe_repo.get_by_id(recipe_id)

    def list_recipes(self) -> List[Recipe]:
        return self.recipe_repo.list_all()

    def create_batch(self, data: BatchCreate) -> Batch:
        if not self.recipe_repo.get_by_id(data.recipe_id):
            raise ValidationError(f"配方ID {data.recipe_id} 不存在")
        
        validate_batch(data)
        return self.batch_repo.create(data)

    def get_batch(self, batch_id: int) -> Optional[Batch]:
        return self.batch_repo.get_by_id(batch_id)

    def list_batches(self) -> List[Batch]:
        return self.batch_repo.list_all()

    def update_batch_actual_quantity(
        self,
        batch_id: int,
        actual_quantity: float,
    ) -> Optional[Batch]:
        batch = self.batch_repo.get_by_id(batch_id)
        if not batch:
            return None
        
        validate_batch_actual_quantity(actual_quantity, batch.planned_quantity)
        batch.actual_quantity = actual_quantity
        return self.batch_repo.update(batch)

    def create_quality_inspection(
        self,
        data: QualityInspectionCreate,
    ) -> QualityInspection:
        batch = self.batch_repo.get_by_id(data.batch_id)
        if not batch:
            raise ValidationError(f"批次ID {data.batch_id} 不存在")
        
        existing_inspections = self.inspection_repo.get_by_batch_id(data.batch_id)
        validate_quality_inspection(data, batch, existing_inspections)
        
        inspection = self.inspection_repo.create(data)
        
        if data.result == InspectionResult.PASS:
            self.batch_repo.set_passed_inspection(data.batch_id)
        
        if data.result == InspectionResult.FAIL:
            existing_todo = self.todo_repo.get_by_batch_id(data.batch_id)
            if not existing_todo:
                self.todo_repo.create(TodoCreate(
                    batch_id=data.batch_id,
                    status=TodoStatus.PENDING,
                    notes="质检不合格，需品控主管处理",
                ))
        
        return inspection

    def get_inspections_by_batch(self, batch_id: int) -> List[QualityInspection]:
        return self.inspection_repo.get_by_batch_id(batch_id)

    def get_todo(self, todo_id: int) -> Optional[Todo]:
        return self.todo_repo.get_by_id(todo_id)

    def list_todos(self) -> List[Todo]:
        return self.todo_repo.list_all()

    def process_todo(
        self,
        todo_id: int,
        status: TodoStatus,
        notes: Optional[str] = None,
    ) -> Optional[Todo]:
        return self.todo_repo.update_status(todo_id, status, notes)

    def create_material_inventory(
        self,
        name: str,
        current_stock: float,
        safety_stock: float,
    ) -> MaterialInventory:
        if current_stock < 0:
            raise ValidationError("当前库存不能为负数")
        if safety_stock < 0:
            raise ValidationError("安全库存不能为负数")
        
        existing = self.material_repo.get_by_name(name)
        if existing:
            raise ValidationError(f"原材料 {name} 已存在")
        
        return self.material_repo.create(name, current_stock, safety_stock)

    def get_material(self, material_id: int) -> Optional[MaterialInventory]:
        return self.material_repo.get_by_id(material_id)

    def list_materials(self) -> List[MaterialInventory]:
        return self.material_repo.list_all()

    def update_material_stock(
        self,
        material_id: int,
        current_stock: float,
    ) -> Optional[MaterialInventory]:
        if current_stock < 0:
            raise ValidationError("当前库存不能为负数")
        return self.material_repo.update_stock(material_id, current_stock)

    def get_low_stock_materials(self) -> List[MaterialInventory]:
        return self.material_repo.get_low_stock()

    def get_batch_summaries(
        self,
        start_date: datetime,
        end_date: datetime,
    ):
        return self.summary_repo.get_batch_summaries(start_date, end_date)
