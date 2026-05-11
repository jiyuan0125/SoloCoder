from datetime import datetime
from typing import List, Optional

from .models import (
    DispositionType,
    ProductionBatch,
    QCResult,
    QualityCheck,
    Recipe,
    RawMaterial,
    RawMaterialItem,
    Todo,
    TodoStatus,
)
from .storage import Storage, get_storage


class BusinessError(Exception):
    pass


class RecipeService:
    def __init__(self, storage: Optional[Storage] = None):
        self.storage = storage or get_storage()

    def create_recipe(self, name: str, category: str, raw_materials: List[RawMaterialItem]) -> Recipe:
        if len(raw_materials) < 2:
            raise BusinessError("配方原材料清单至少需要两条")
        recipe = Recipe(name=name, category=category, raw_materials=raw_materials)
        self.storage.save_recipe(recipe)
        return recipe

    def get_recipe(self, recipe_id: str) -> Optional[Recipe]:
        return self.storage.get_recipe(recipe_id)

    def list_recipes(self) -> List[Recipe]:
        return self.storage.list_recipes()


class RawMaterialService:
    def __init__(self, storage: Optional[Storage] = None):
        self.storage = storage or get_storage()

    def create_or_update_material(self, name: str, current_stock: float, safety_stock: float) -> RawMaterial:
        existing = self.storage.get_raw_material_by_name(name)
        if existing:
            existing.current_stock = current_stock
            existing.safety_stock = safety_stock
            existing.updated_at = datetime.utcnow()
            return self.storage.save_raw_material(existing)
        material = RawMaterial(name=name, current_stock=current_stock, safety_stock=safety_stock)
        return self.storage.save_raw_material(material)

    def get_material(self, material_id: str) -> Optional[RawMaterial]:
        return self.storage.get_raw_material(material_id)

    def list_materials(self) -> List[RawMaterial]:
        return self.storage.list_raw_materials()

    def list_low_stock_materials(self) -> List[RawMaterial]:
        return [m for m in self.storage.list_raw_materials() if m.is_below_safety]


class ProductionService:
    MAX_ACTUAL_RATIO = 1.10

    def __init__(self, storage: Optional[Storage] = None):
        self.storage = storage or get_storage()

    def create_batch(self, recipe_id: str, plan_quantity: int, actual_quantity: int) -> ProductionBatch:
        recipe = self.storage.get_recipe(recipe_id)
        if not recipe:
            raise BusinessError(f"配方不存在: {recipe_id}")
        if plan_quantity <= 0:
            raise BusinessError("计划数量必须大于0")
        max_allowed = int(plan_quantity * self.MAX_ACTUAL_RATIO)
        if actual_quantity > max_allowed:
            raise BusinessError(f"实际成品数量不能超过计划的110%（最大允许: {max_allowed}）")
        if actual_quantity < 0:
            raise BusinessError("实际数量不能为负数")
        batch = ProductionBatch(
            recipe_id=recipe_id,
            plan_quantity=plan_quantity,
            actual_quantity=actual_quantity,
        )
        self.storage.save_batch(batch)
        return batch

    def get_batch(self, batch_id: str) -> Optional[ProductionBatch]:
        return self.storage.get_batch(batch_id)

    def list_batches(self, start: Optional[datetime] = None, end: Optional[datetime] = None) -> List[ProductionBatch]:
        return self.storage.list_batches(start, end)

    def add_quality_check(
        self,
        batch_id: str,
        inspector: str,
        result: QCResult,
        measured_value: Optional[float],
        notes: str = "",
    ) -> ProductionBatch:
        batch = self.storage.get_batch(batch_id)
        if not batch:
            raise BusinessError(f"批次不存在: {batch_id}")
        if batch.has_passed_qc:
            raise BusinessError("已有合格质检的批次不能再加新质检记录")
        if result == QCResult.PASS and measured_value is None:
            raise BusinessError("检测数值为空但结果填合格的是矛盾数据")
        qc = QualityCheck(
            batch_id=batch_id,
            inspector=inspector,
            result=result,
            measured_value=measured_value,
            notes=notes,
        )
        self.storage.add_quality_check(batch_id, qc)
        if result == QCResult.FAIL and not batch.has_pending_todo:
            todo = Todo(batch_id=batch_id, status=TodoStatus.PENDING)
            self.storage.save_todo(todo)
            batch.todos.append(todo)
        return batch

    def resolve_todo(
        self,
        todo_id: str,
        disposition: DispositionType,
        resolved_by: str,
    ) -> Todo:
        todo = self.storage.get_todo(todo_id)
        if not todo:
            raise BusinessError(f"待办不存在: {todo_id}")
        if todo.status == TodoStatus.RESOLVED:
            raise BusinessError("待办已处理")
        todo.status = TodoStatus.RESOLVED
        todo.disposition = disposition
        todo.resolved_by = resolved_by
        todo.resolved_at = datetime.utcnow()
        return self.storage.save_todo(todo)

    def list_todos(self) -> List[Todo]:
        return self.storage.list_todos()

    def get_todo(self, todo_id: str) -> Optional[Todo]:
        return self.storage.get_todo(todo_id)


class ExportService:
    def __init__(self, storage: Optional[Storage] = None):
        self.storage = storage or get_storage()

    def export_batches(self, start: Optional[datetime] = None, end: Optional[datetime] = None) -> str:
        batches = self.storage.list_batches(start, end)
        lines = []
        header = [
            "批次ID",
            "配方ID",
            "计划数量",
            "实际数量",
            "合格数量",
            "不合格数量",
            "待办数量",
            "创建时间",
        ]
        lines.append(self._format_row(header))
        lines.append(self._format_separator(len(header)))
        for batch in batches:
            pass_count = sum(1 for qc in batch.quality_checks if qc.result == QCResult.PASS)
            fail_count = sum(1 for qc in batch.quality_checks if qc.result == QCResult.FAIL)
            pending_todos = sum(1 for t in batch.todos if t.status == TodoStatus.PENDING)
            row = [
                batch.id[:8],
                batch.recipe_id[:8],
                str(batch.plan_quantity),
                str(batch.actual_quantity),
                str(pass_count),
                str(fail_count),
                str(pending_todos),
                batch.created_at.strftime("%Y-%m-%d %H:%M:%S"),
            ]
            lines.append(self._format_row(row))
        return "\n".join(lines)

    def _format_row(self, cells: List[str], width: int = 14) -> str:
        return "| " + " | ".join(c.ljust(width) for c in cells) + " |"

    def _format_separator(self, count: int, width: int = 14) -> str:
        return "+" + "+".join("-" * (width + 2) for _ in range(count)) + "+"
