from typing import List, Optional
from datetime import datetime, timedelta
from decimal import Decimal
from sqlalchemy.orm import Session

from src.core.models import (
    RecipeCreate, RecipeUpdate, Recipe, RawMaterial,
    ProductionBatchCreate, ProductionBatchUpdate, ProductionBatch,
    QualityInspectionCreate, QualityInspection,
    TodoItemUpdate, TodoItem,
    RecipeStatus, BatchStatus, InspectionStatus, TodoStatus, TodoResolution
)
from src.core.utils import check_overdue, split_date_range, format_as_text_table, decimal_to_str
from src.server.database import (
    DBRecipe, DBRawMaterial, DBProductionBatch, DBQualityInspection, DBTodoItem
)


def create_recipe(db: Session, recipe_data: RecipeCreate) -> Recipe:
    db_recipe = DBRecipe(
        name=recipe_data.name,
        description=recipe_data.description,
        status=RecipeStatus.ACTIVE
    )
    db.add(db_recipe)
    db.flush()

    for material in recipe_data.raw_materials:
        db_material = DBRawMaterial(
            recipe_id=db_recipe.id,
            name=material.name,
            percentage=material.percentage
        )
        db.add(db_material)

    db.commit()
    db.refresh(db_recipe)
    return _recipe_to_pydantic(db_recipe)


def get_recipe(db: Session, recipe_id: int) -> Optional[Recipe]:
    db_recipe = db.query(DBRecipe).filter(DBRecipe.id == recipe_id).first()
    if db_recipe:
        return _recipe_to_pydantic(db_recipe)
    return None


def get_all_recipes(db: Session) -> List[Recipe]:
    db_recipes = db.query(DBRecipe).all()
    return [_recipe_to_pydantic(r) for r in db_recipes]


def update_recipe(db: Session, recipe_id: int, recipe_data: RecipeUpdate) -> Optional[Recipe]:
    db_recipe = db.query(DBRecipe).filter(DBRecipe.id == recipe_id).first()
    if not db_recipe:
        return None

    if recipe_data.name is not None:
        db_recipe.name = recipe_data.name
    if recipe_data.description is not None:
        db_recipe.description = recipe_data.description

    db.commit()
    db.refresh(db_recipe)
    return _recipe_to_pydantic(db_recipe)


def deactivate_recipe(db: Session, recipe_id: int) -> tuple[bool, str]:
    db_recipe = db.query(DBRecipe).filter(DBRecipe.id == recipe_id).first()
    if not db_recipe:
        return False, "配方不存在"

    if db_recipe.status == RecipeStatus.INACTIVE:
        return False, "配方已停用"

    active_batches = db.query(DBProductionBatch).filter(
        DBProductionBatch.recipe_id == recipe_id,
        DBProductionBatch.status.in_([BatchStatus.PENDING, BatchStatus.IN_PROGRESS])
    ).first()

    if active_batches:
        return False, "配方有进行中批次，无法停用"

    db_recipe.status = RecipeStatus.INACTIVE
    db.commit()
    return True, "配方已停用"


def activate_recipe(db: Session, recipe_id: int) -> tuple[bool, str]:
    db_recipe = db.query(DBRecipe).filter(DBRecipe.id == recipe_id).first()
    if not db_recipe:
        return False, "配方不存在"

    if db_recipe.status == RecipeStatus.ACTIVE:
        return False, "配方已启用"

    db_recipe.status = RecipeStatus.ACTIVE
    db.commit()
    return True, "配方已启用"


def create_batch(db: Session, batch_data: ProductionBatchCreate) -> tuple[Optional[ProductionBatch], str]:
    db_recipe = db.query(DBRecipe).filter(
        DBRecipe.id == batch_data.recipe_id,
        DBRecipe.status == RecipeStatus.ACTIVE
    ).first()
    if not db_recipe:
        return None, "配方不存在或未启用"

    db_batch = DBProductionBatch(
        recipe_id=batch_data.recipe_id,
        planned_quantity=batch_data.planned_quantity,
        actual_quantity=batch_data.actual_quantity,
        status=BatchStatus.PENDING
    )
    db.add(db_batch)
    db.commit()
    db.refresh(db_batch)
    return _batch_to_pydantic(db_batch), "批次创建成功"


def get_batch(db: Session, batch_id: int) -> Optional[ProductionBatch]:
    db_batch = db.query(DBProductionBatch).filter(DBProductionBatch.id == batch_id).first()
    if db_batch:
        return _batch_to_pydantic(db_batch)
    return None


def get_all_batches(db: Session) -> List[ProductionBatch]:
    db_batches = db.query(DBProductionBatch).all()
    return [_batch_to_pydantic(b) for b in db_batches]


def update_batch(db: Session, batch_id: int, batch_data: ProductionBatchUpdate) -> tuple[Optional[ProductionBatch], str]:
    db_batch = db.query(DBProductionBatch).filter(DBProductionBatch.id == batch_id).first()
    if not db_batch:
        return None, "批次不存在"

    if batch_data.actual_quantity is not None:
        max_allowed = db_batch.planned_quantity * Decimal("1.05")
        if batch_data.actual_quantity > max_allowed:
            return None, f"实际产量不能超过计划的105%（最大{max_allowed}）"
        db_batch.actual_quantity = batch_data.actual_quantity

    if batch_data.status is not None:
        db_batch.status = batch_data.status

    db.commit()
    db.refresh(db_batch)
    return _batch_to_pydantic(db_batch), "批次更新成功"


def create_inspection(db: Session, inspection_data: QualityInspectionCreate) -> tuple[Optional[QualityInspection], str]:
    db_batch = db.query(DBProductionBatch).filter(
        DBProductionBatch.id == inspection_data.batch_id
    ).first()
    if not db_batch:
        return None, "批次不存在"

    has_pass = db.query(DBQualityInspection).filter(
        DBQualityInspection.batch_id == inspection_data.batch_id,
        DBQualityInspection.status == InspectionStatus.PASS
    ).first()
    if has_pass:
        return None, "批次已有合格质检，不能添加新质检"

    db_inspection = DBQualityInspection(
        batch_id=inspection_data.batch_id,
        status=inspection_data.status,
        notes=inspection_data.notes
    )
    db.add(db_inspection)
    db.flush()

    if inspection_data.status == InspectionStatus.FAIL:
        existing_todo = db.query(DBTodoItem).filter(
            DBTodoItem.batch_id == inspection_data.batch_id,
            DBTodoItem.status.in_([TodoStatus.PENDING, TodoStatus.IN_PROGRESS])
        ).first()
        if not existing_todo:
            db_todo = DBTodoItem(
                batch_id=inspection_data.batch_id,
                status=TodoStatus.PENDING
            )
            db.add(db_todo)

    db.commit()
    db.refresh(db_inspection)
    return _inspection_to_pydantic(db_inspection), "质检记录创建成功"


def get_batch_inspections(db: Session, batch_id: int) -> List[QualityInspection]:
    inspections = db.query(DBQualityInspection).filter(
        DBQualityInspection.batch_id == batch_id
    ).all()
    return [_inspection_to_pydantic(i) for i in inspections]


def get_all_todos(db: Session) -> List[TodoItem]:
    db_todos = db.query(DBTodoItem).all()
    now = datetime.utcnow()
    result = []
    for todo in db_todos:
        is_overdue = check_overdue(todo.created_at, now)
        if is_overdue and todo.status == TodoStatus.PENDING:
            todo.status = TodoStatus.OVERDUE
            db.commit()
            db.refresh(todo)
        pydantic_todo = _todo_to_pydantic(todo)
        pydantic_todo.is_overdue = is_overdue or todo.status == TodoStatus.OVERDUE
        result.append(pydantic_todo)
    return result


def get_todo(db: Session, todo_id: int) -> Optional[TodoItem]:
    db_todo = db.query(DBTodoItem).filter(DBTodoItem.id == todo_id).first()
    if db_todo:
        now = datetime.utcnow()
        is_overdue = check_overdue(db_todo.created_at, now)
        if is_overdue and db_todo.status == TodoStatus.PENDING:
            db_todo.status = TodoStatus.OVERDUE
            db.commit()
            db.refresh(db_todo)
        pydantic_todo = _todo_to_pydantic(db_todo)
        pydantic_todo.is_overdue = is_overdue or db_todo.status == TodoStatus.OVERDUE
        return pydantic_todo
    return None


def update_todo(db: Session, todo_id: int, todo_data: TodoItemUpdate) -> tuple[Optional[TodoItem], str]:
    db_todo = db.query(DBTodoItem).filter(DBTodoItem.id == todo_id).first()
    if not db_todo:
        return None, "待办不存在"

    if db_todo.status == TodoStatus.RESOLVED:
        return None, "待办已处理"

    db_todo.status = todo_data.status
    if todo_data.notes is not None:
        db_todo.notes = todo_data.notes

    if todo_data.status == TodoStatus.RESOLVED:
        if todo_data.resolution is None:
            return None, "处理待办时必须指定处理方式"
        db_todo.resolution = todo_data.resolution
        db_todo.resolved_at = datetime.utcnow()

    db.commit()
    db.refresh(db_todo)
    return _todo_to_pydantic(db_todo), "待办更新成功"


def export_batches(db: Session, start_date: datetime, end_date: datetime) -> str:
    segments = split_date_range(start_date, end_date)
    all_lines = []

    for seg_start, seg_end in segments:
        segment_header = f"\n{'=' * 60}\n"
        if len(segments) > 1:
            segment_header += f"时段: {seg_start.strftime('%Y-%m-%d')} 至 {seg_end.strftime('%Y-%m-%d')}\n"
            segment_header += f"{'=' * 60}\n"

        batches = db.query(DBProductionBatch).filter(
            DBProductionBatch.created_at >= seg_start,
            DBProductionBatch.created_at < seg_end
        ).all()

        if not batches:
            segment_header += "该时段无生产批次记录\n"
            all_lines.append(segment_header)
            continue

        rows = []
        for batch in batches:
            recipe = db.query(DBRecipe).filter(DBRecipe.id == batch.recipe_id).first()
            recipe_name = recipe.name if recipe else "未知配方"

            actual_qty = decimal_to_str(batch.actual_quantity) if batch.actual_quantity else "-"
            yield_rate = "-"
            if batch.actual_quantity and batch.planned_quantity > 0:
                yield_rate = f"{(batch.actual_quantity / batch.planned_quantity * 100):.1f}%"

            inspections = db.query(DBQualityInspection).filter(
                DBQualityInspection.batch_id == batch.id
            ).all()
            has_pass = any(i.status == InspectionStatus.PASS for i in inspections)
            has_fail = any(i.status == InspectionStatus.FAIL for i in inspections)

            if has_pass:
                quality_status = "合格"
            elif has_fail:
                quality_status = "不合格"
            else:
                quality_status = "未质检"

            rows.append([
                str(batch.id),
                recipe_name,
                batch.status.value,
                decimal_to_str(batch.planned_quantity),
                actual_qty,
                yield_rate,
                quality_status,
                batch.created_at.strftime("%Y-%m-%d %H:%M")
            ])

        headers = ["批次ID", "配方名称", "状态", "计划产量", "实际产量", "良率", "质检状态", "创建时间"]
        table_text = format_as_text_table(headers, rows)
        all_lines.append(segment_header + table_text)

    return "\n".join(all_lines).strip()


def _recipe_to_pydantic(db_recipe: DBRecipe) -> Recipe:
    return Recipe(
        id=db_recipe.id,
        name=db_recipe.name,
        description=db_recipe.description,
        status=db_recipe.status,
        created_at=db_recipe.created_at,
        raw_materials=[
            RawMaterial(
                id=m.id,
                recipe_id=m.recipe_id,
                name=m.name,
                percentage=m.percentage
            ) for m in db_recipe.raw_materials
        ]
    )


def _batch_to_pydantic(db_batch: DBProductionBatch) -> ProductionBatch:
    return ProductionBatch(
        id=db_batch.id,
        recipe_id=db_batch.recipe_id,
        planned_quantity=db_batch.planned_quantity,
        actual_quantity=db_batch.actual_quantity,
        status=db_batch.status,
        created_at=db_batch.created_at,
        updated_at=db_batch.updated_at
    )


def _inspection_to_pydantic(db_inspection: DBQualityInspection) -> QualityInspection:
    return QualityInspection(
        id=db_inspection.id,
        batch_id=db_inspection.batch_id,
        status=db_inspection.status,
        notes=db_inspection.notes,
        created_at=db_inspection.created_at
    )


def _todo_to_pydantic(db_todo: DBTodoItem) -> TodoItem:
    return TodoItem(
        id=db_todo.id,
        batch_id=db_todo.batch_id,
        status=db_todo.status,
        resolution=db_todo.resolution,
        notes=db_todo.notes,
        created_at=db_todo.created_at,
        resolved_at=db_todo.resolved_at,
        is_overdue=False
    )
