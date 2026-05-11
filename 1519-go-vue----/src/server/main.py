import os
from fastapi import FastAPI, Depends, HTTPException
from fastapi.responses import PlainTextResponse
from typing import List
from datetime import datetime, date
from sqlalchemy.orm import Session

from src.core.models import (
    RecipeCreate, RecipeUpdate, Recipe,
    ProductionBatchCreate, ProductionBatchUpdate, ProductionBatch,
    QualityInspectionCreate, QualityInspection,
    TodoItemUpdate, TodoItem
)
from src.server.database import get_db, init_db
from src.server import services


app = FastAPI(title="饲料加工厂管理系统", version="1.0.0")


@app.on_event("startup")
def on_startup():
    init_db()


@app.get("/")
def root():
    return {"message": "饲料加工厂管理系统 API", "version": "1.0.0"}


@app.post("/recipes/", response_model=Recipe, status_code=201)
def create_recipe(recipe: RecipeCreate, db: Session = Depends(get_db)):
    try:
        return services.create_recipe(db, recipe)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/recipes/", response_model=List[Recipe])
def list_recipes(db: Session = Depends(get_db)):
    return services.get_all_recipes(db)


@app.get("/recipes/{recipe_id}", response_model=Recipe)
def get_recipe(recipe_id: int, db: Session = Depends(get_db)):
    recipe = services.get_recipe(db, recipe_id)
    if not recipe:
        raise HTTPException(status_code=404, detail="配方不存在")
    return recipe


@app.put("/recipes/{recipe_id}", response_model=Recipe)
def update_recipe(recipe_id: int, recipe: RecipeUpdate, db: Session = Depends(get_db)):
    updated = services.update_recipe(db, recipe_id, recipe)
    if not updated:
        raise HTTPException(status_code=404, detail="配方不存在")
    return updated


@app.post("/recipes/{recipe_id}/deactivate")
def deactivate_recipe(recipe_id: int, db: Session = Depends(get_db)):
    success, message = services.deactivate_recipe(db, recipe_id)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"message": message}


@app.post("/recipes/{recipe_id}/activate")
def activate_recipe(recipe_id: int, db: Session = Depends(get_db)):
    success, message = services.activate_recipe(db, recipe_id)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"message": message}


@app.post("/batches/", response_model=ProductionBatch, status_code=201)
def create_batch(batch: ProductionBatchCreate, db: Session = Depends(get_db)):
    created, message = services.create_batch(db, batch)
    if not created:
        raise HTTPException(status_code=400, detail=message)
    return created


@app.get("/batches/", response_model=List[ProductionBatch])
def list_batches(db: Session = Depends(get_db)):
    return services.get_all_batches(db)


@app.get("/batches/{batch_id}", response_model=ProductionBatch)
def get_batch(batch_id: int, db: Session = Depends(get_db)):
    batch = services.get_batch(db, batch_id)
    if not batch:
        raise HTTPException(status_code=404, detail="批次不存在")
    return batch


@app.put("/batches/{batch_id}", response_model=ProductionBatch)
def update_batch(batch_id: int, batch: ProductionBatchUpdate, db: Session = Depends(get_db)):
    updated, message = services.update_batch(db, batch_id, batch)
    if not updated:
        raise HTTPException(status_code=400, detail=message)
    return updated


@app.post("/inspections/", response_model=QualityInspection, status_code=201)
def create_inspection(inspection: QualityInspectionCreate, db: Session = Depends(get_db)):
    created, message = services.create_inspection(db, inspection)
    if not created:
        raise HTTPException(status_code=400, detail=message)
    return created


@app.get("/batches/{batch_id}/inspections", response_model=List[QualityInspection])
def list_batch_inspections(batch_id: int, db: Session = Depends(get_db)):
    return services.get_batch_inspections(db, batch_id)


@app.get("/todos/", response_model=List[TodoItem])
def list_todos(db: Session = Depends(get_db)):
    return services.get_all_todos(db)


@app.get("/todos/{todo_id}", response_model=TodoItem)
def get_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = services.get_todo(db, todo_id)
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    return todo


@app.put("/todos/{todo_id}", response_model=TodoItem)
def update_todo(todo_id: int, todo: TodoItemUpdate, db: Session = Depends(get_db)):
    updated, message = services.update_todo(db, todo_id, todo)
    if not updated:
        raise HTTPException(status_code=400, detail=message)
    return updated


@app.get("/export/batches", response_class=PlainTextResponse)
def export_batches(start: date, end: date, db: Session = Depends(get_db)):
    start_dt = datetime.combine(start, datetime.min.time())
    end_dt = datetime.combine(end, datetime.max.time())
    return services.export_batches(db, start_dt, end_dt)
