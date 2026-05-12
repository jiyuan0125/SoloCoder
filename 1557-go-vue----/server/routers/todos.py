from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from .. import models, schemas
from ..database import get_db

router = APIRouter()


@router.get("/", response_model=List[schemas.TodoItem])
def list_todos(
    vessel_id: Optional[int] = Query(None),
    status: Optional[str] = Query(None),
    todo_type: Optional[str] = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(models.TodoItem)

    if vessel_id:
        query = query.filter(models.TodoItem.vessel_id == vessel_id)
    if status:
        query = query.filter(models.TodoItem.status == status)
    if todo_type:
        query = query.filter(models.TodoItem.todo_type == todo_type)

    todos = query.order_by(models.TodoItem.due_date.asc()).offset(skip).limit(limit).all()
    return todos


@router.get("/{todo_id}", response_model=schemas.TodoItem)
def get_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = (
        db.query(models.TodoItem)
        .filter(models.TodoItem.id == todo_id)
        .first()
    )
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    return todo


@router.put("/{todo_id}", response_model=schemas.TodoItem)
def update_todo(
    todo_id: int,
    todo_update: schemas.TodoItemUpdate,
    db: Session = Depends(get_db),
):
    todo = (
        db.query(models.TodoItem)
        .filter(models.TodoItem.id == todo_id)
        .first()
    )
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")

    update_data = todo_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(todo, key, value)

    db.commit()
    db.refresh(todo)
    return todo


@router.get("/reminders/")
def list_reminders(
    vessel_id: Optional[int] = Query(None),
    is_read: Optional[bool] = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(models.Reminder)

    if vessel_id:
        query = query.filter(models.Reminder.vessel_id == vessel_id)
    if is_read is not None:
        query = query.filter(models.Reminder.is_read == is_read)

    reminders = query.order_by(models.Reminder.created_at.desc()).offset(skip).limit(limit).all()
    return reminders


@router.put("/reminders/{reminder_id}/read")
def mark_reminder_read(reminder_id: int, db: Session = Depends(get_db)):
    reminder = (
        db.query(models.Reminder)
        .filter(models.Reminder.id == reminder_id)
        .first()
    )
    if not reminder:
        raise HTTPException(status_code=404, detail="提醒不存在")

    reminder.is_read = True
    db.commit()
    db.refresh(reminder)
    return reminder
