from datetime import date
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from core import schemas
from core.services import TodoService
from server.dependencies import get_db_session

router = APIRouter(prefix="/todos", tags=["todos"])


@router.get("", response_model=List[schemas.Todo])
def list_todos(
    todo_date: Optional[date] = Query(None),
    pending: bool = Query(False),
    db: Session = Depends(get_db_session)
):
    if pending:
        return TodoService.get_all_pending(db)
    if todo_date:
        return TodoService.get_by_date(db, todo_date)
    return []


@router.get("/{todo_id}", response_model=schemas.Todo)
def get_todo(
    todo_id: int,
    db: Session = Depends(get_db_session)
):
    todo = TodoService.get_by_id(db, todo_id)
    if not todo:
        raise HTTPException(status_code=404, detail="Todo not found")
    return todo


@router.put("/{todo_id}", response_model=schemas.Todo)
def update_todo(
    todo_id: int,
    todo_in: schemas.TodoUpdate,
    db: Session = Depends(get_db_session)
):
    todo = TodoService.update(db, todo_id, todo_in)
    if not todo:
        raise HTTPException(status_code=404, detail="Todo not found")
    return todo
