from typing import List
from fastapi import APIRouter, Depends, HTTPException
from src.core.models import TodoItem
from src.core.services import TodoService
from src.core.exceptions import NotFoundError
from src.server.dependencies import get_todo_service


router = APIRouter(prefix="/todos", tags=["todos"])


@router.get("/", response_model=List[TodoItem])
def list_todos(
    service: TodoService = Depends(get_todo_service)
):
    return service.list_todos()


@router.get("/incomplete", response_model=List[TodoItem])
def list_incomplete_todos(
    service: TodoService = Depends(get_todo_service)
):
    return service.list_incomplete_todos()


@router.get("/{todo_id}", response_model=TodoItem)
def get_todo(
    todo_id: str,
    service: TodoService = Depends(get_todo_service)
):
    try:
        return service.get_todo(todo_id)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.patch("/{todo_id}/complete", response_model=TodoItem)
def complete_todo(
    todo_id: str,
    service: TodoService = Depends(get_todo_service)
):
    try:
        return service.update_todo_status(todo_id, True)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))


@router.patch("/{todo_id}/incomplete", response_model=TodoItem)
def reopen_todo(
    todo_id: str,
    service: TodoService = Depends(get_todo_service)
):
    try:
        return service.update_todo_status(todo_id, False)
    except NotFoundError as e:
        raise HTTPException(status_code=404, detail=str(e))
