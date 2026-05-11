from typing import List

from fastapi import APIRouter, Depends, HTTPException, status

from core.models import Todo, TodoStatus, TodoUpdate, ValidationError
from core.services import ProductionService
from server.deps import get_production_service

router = APIRouter(prefix="/todos", tags=["todos"])


@router.get("", response_model=List[Todo])
def list_todos(service: ProductionService = Depends(get_production_service)):
    return service.list_todos()


@router.get("/{todo_id}", response_model=Todo)
def get_todo(
    todo_id: int,
    service: ProductionService = Depends(get_production_service),
):
    todo = service.get_todo(todo_id)
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    return todo


@router.patch("/{todo_id}", response_model=Todo)
def process_todo(
    todo_id: int,
    data: TodoUpdate,
    service: ProductionService = Depends(get_production_service),
):
    todo = service.process_todo(todo_id, data.status, data.notes)
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")
    return todo
