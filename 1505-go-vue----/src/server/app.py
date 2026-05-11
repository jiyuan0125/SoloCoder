from datetime import datetime
from typing import List, Optional

from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import PlainTextResponse

from core.models import (
    DispositionType,
    ProductionBatch,
    QCResult,
    Recipe,
    RawMaterial,
    RawMaterialItem,
)
from core.services import (
    BusinessError,
    ExportService,
    ProductionService,
    RawMaterialService,
    RecipeService,
)

from . import schemas


def recipe_to_out(recipe: Recipe) -> schemas.RecipeOut:
    return schemas.RecipeOut(
        id=recipe.id,
        name=recipe.name,
        category=recipe.category,
        raw_materials=[
            schemas.RawMaterialItemOut(name=item.name, quantity=item.quantity, is_key=item.is_key)
            for item in recipe.raw_materials
        ],
        created_at=recipe.created_at.isoformat(),
    )


def raw_material_to_out(material: RawMaterial) -> schemas.RawMaterialOut:
    return schemas.RawMaterialOut(
        id=material.id,
        name=material.name,
        current_stock=material.current_stock,
        safety_stock=material.safety_stock,
        is_below_safety=material.is_below_safety,
        created_at=material.created_at.isoformat(),
        updated_at=material.updated_at.isoformat(),
    )


def qc_to_out(qc) -> schemas.QualityCheckOut:
    return schemas.QualityCheckOut(
        id=qc.id,
        batch_id=qc.batch_id,
        inspector=qc.inspector,
        result=qc.result.value,
        measured_value=qc.measured_value,
        notes=qc.notes,
        created_at=qc.created_at.isoformat(),
    )


def todo_to_out(todo) -> schemas.TodoOut:
    return schemas.TodoOut(
        id=todo.id,
        batch_id=todo.batch_id,
        status=todo.status.value,
        disposition=todo.disposition.value if todo.disposition else None,
        resolved_by=todo.resolved_by,
        resolved_at=todo.resolved_at.isoformat() if todo.resolved_at else None,
        created_at=todo.created_at.isoformat(),
    )


def batch_to_out(batch: ProductionBatch) -> schemas.ProductionBatchOut:
    return schemas.ProductionBatchOut(
        id=batch.id,
        recipe_id=batch.recipe_id,
        plan_quantity=batch.plan_quantity,
        actual_quantity=batch.actual_quantity,
        has_passed_qc=batch.has_passed_qc,
        has_pending_todo=batch.has_pending_todo,
        quality_checks=[qc_to_out(qc) for qc in batch.quality_checks],
        todos=[todo_to_out(t) for t in batch.todos],
        created_at=batch.created_at.isoformat(),
    )


def create_app() -> FastAPI:
    app = FastAPI(title="食品加工厂生产管理系统")

    recipe_service = RecipeService()
    raw_material_service = RawMaterialService()
    production_service = ProductionService()
    export_service = ExportService()

    @app.post("/recipes", response_model=schemas.RecipeOut)
    def create_recipe(data: schemas.RecipeIn):
        try:
            raw_materials = [
                RawMaterialItem(name=item.name, quantity=item.quantity, is_key=item.is_key)
                for item in data.raw_materials
            ]
            recipe = recipe_service.create_recipe(data.name, data.category, raw_materials)
            return recipe_to_out(recipe)
        except BusinessError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.get("/recipes", response_model=List[schemas.RecipeOut])
    def list_recipes():
        return [recipe_to_out(r) for r in recipe_service.list_recipes()]

    @app.get("/recipes/{recipe_id}", response_model=schemas.RecipeOut)
    def get_recipe(recipe_id: str):
        recipe = recipe_service.get_recipe(recipe_id)
        if not recipe:
            raise HTTPException(status_code=404, detail="配方不存在")
        return recipe_to_out(recipe)

    @app.post("/raw-materials", response_model=schemas.RawMaterialOut)
    def create_or_update_material(data: schemas.RawMaterialIn):
        material = raw_material_service.create_or_update_material(
            data.name, data.current_stock, data.safety_stock
        )
        return raw_material_to_out(material)

    @app.get("/raw-materials", response_model=List[schemas.RawMaterialOut])
    def list_materials(low_stock: bool = False):
        if low_stock:
            materials = raw_material_service.list_low_stock_materials()
        else:
            materials = raw_material_service.list_materials()
        return [raw_material_to_out(m) for m in materials]

    @app.get("/raw-materials/{material_id}", response_model=schemas.RawMaterialOut)
    def get_material(material_id: str):
        material = raw_material_service.get_material(material_id)
        if not material:
            raise HTTPException(status_code=404, detail="原材料不存在")
        return raw_material_to_out(material)

    @app.post("/batches", response_model=schemas.ProductionBatchOut)
    def create_batch(data: schemas.ProductionBatchIn):
        try:
            batch = production_service.create_batch(
                data.recipe_id, data.plan_quantity, data.actual_quantity
            )
            return batch_to_out(batch)
        except BusinessError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.get("/batches", response_model=List[schemas.ProductionBatchOut])
    def list_batches(
        start: Optional[str] = Query(None, description="开始日期，格式 YYYY-MM-DD"),
        end: Optional[str] = Query(None, description="结束日期，格式 YYYY-MM-DD"),
    ):
        start_dt = None
        end_dt = None
        if start:
            start_dt = datetime.fromisoformat(start)
        if end:
            end_dt = datetime.fromisoformat(end)
        batches = production_service.list_batches(start_dt, end_dt)
        return [batch_to_out(b) for b in batches]

    @app.get("/batches/{batch_id}", response_model=schemas.ProductionBatchOut)
    def get_batch(batch_id: str):
        batch = production_service.get_batch(batch_id)
        if not batch:
            raise HTTPException(status_code=404, detail="批次不存在")
        return batch_to_out(batch)

    @app.post("/batches/{batch_id}/quality-checks", response_model=schemas.ProductionBatchOut)
    def add_quality_check(batch_id: str, data: schemas.QualityCheckIn):
        try:
            result = QCResult(data.result)
            batch = production_service.add_quality_check(
                batch_id, data.inspector, result, data.measured_value, data.notes
            )
            return batch_to_out(batch)
        except BusinessError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.get("/todos", response_model=List[schemas.TodoOut])
    def list_todos():
        return [todo_to_out(t) for t in production_service.list_todos()]

    @app.post("/todos/{todo_id}/resolve", response_model=schemas.TodoOut)
    def resolve_todo(todo_id: str, data: schemas.ResolveTodoIn):
        try:
            disposition = DispositionType(data.disposition)
            todo = production_service.resolve_todo(todo_id, disposition, data.resolved_by)
            return todo_to_out(todo)
        except BusinessError as e:
            raise HTTPException(status_code=400, detail=str(e))

    @app.get("/export/batches", response_class=PlainTextResponse)
    def export_batches(
        start: Optional[str] = Query(None, description="开始日期，格式 YYYY-MM-DD"),
        end: Optional[str] = Query(None, description="结束日期，格式 YYYY-MM-DD"),
    ):
        start_dt = None
        end_dt = None
        if start:
            start_dt = datetime.fromisoformat(start)
        if end:
            end_dt = datetime.fromisoformat(end)
        return export_service.export_batches(start_dt, end_dt)

    return app
