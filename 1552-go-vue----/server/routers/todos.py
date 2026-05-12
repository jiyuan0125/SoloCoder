from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from .. import models, schemas, services
from ..database import get_db

router = APIRouter(prefix="/api/todos", tags=["todos"])


@router.post("/", response_model=schemas.Todo)
def create_todo(todo: schemas.TodoCreate, db: Session = Depends(get_db)):
    return services.create_todo(db, todo)


@router.get("/", response_model=List[schemas.Todo])
def list_todos(
    agent_service_id: int = None,
    status: models.TodoStatus = None,
    db: Session = Depends(get_db)
):
    query = db.query(models.Todo)
    if agent_service_id:
        query = query.filter(models.Todo.agent_service_id == agent_service_id)
    if status:
        query = query.filter(models.Todo.status == status)
    return query.order_by(models.Todo.due_time.asc()).all()


@router.get("/{todo_id}", response_model=schemas.Todo)
def get_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = db.query(models.Todo).filter(models.Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    return todo


@router.put("/{todo_id}", response_model=schemas.Todo)
def update_todo(todo_id: int, data: schemas.TodoUpdate, db: Session = Depends(get_db)):
    todo = db.query(models.Todo).filter(models.Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(todo, key, value)
    db.commit()
    db.refresh(todo)
    return todo


@router.post("/{todo_id}/complete", response_model=schemas.Todo)
def complete_todo(todo_id: int, db: Session = Depends(get_db)):
    return services.complete_todo(db, todo_id)


@router.delete("/{todo_id}")
def delete_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = db.query(models.Todo).filter(models.Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    db.delete(todo)
    db.commit()
    return {"message": "已删除"}
