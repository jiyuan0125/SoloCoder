from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date, datetime
from database import get_db
from models import Todo, TodoStatus, TodoType
from schemas import TodoCreate, TodoUpdate, TodoResponse

router = APIRouter(prefix="/api/todos", tags=["todos"])


@router.post("", response_model=TodoResponse)
def create_todo(todo: TodoCreate, db: Session = Depends(get_db)):
    db_todo = Todo(**todo.dict())
    db.add(db_todo)
    db.commit()
    db.refresh(db_todo)
    return db_todo


@router.get("", response_model=List[TodoResponse])
def list_todos(
    todo_type: Optional[str] = None,
    status: Optional[str] = None,
    related_id: Optional[int] = None,
    related_type: Optional[str] = None,
    due_before: Optional[date] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Todo)
    if todo_type:
        try:
            query = query.filter(Todo.todo_type == TodoType(todo_type))
        except ValueError:
            raise HTTPException(status_code=400, detail=f"Invalid todo type: {todo_type}")
    if status:
        try:
            query = query.filter(Todo.status == TodoStatus(status))
        except ValueError:
            raise HTTPException(status_code=400, detail=f"Invalid status: {status}")
    if related_id:
        query = query.filter(Todo.related_id == related_id)
    if related_type:
        query = query.filter(Todo.related_type == related_type)
    if due_before:
        query = query.filter(Todo.due_date <= due_before)
    return query.order_by(Todo.due_date.asc(), Todo.created_at.desc()).all()


@router.get("/pending", response_model=List[TodoResponse])
def list_pending_todos(db: Session = Depends(get_db)):
    return db.query(Todo).filter(Todo.status == TodoStatus.PENDING).order_by(Todo.due_date.asc()).all()


@router.get("/{todo_id}", response_model=TodoResponse)
def get_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="Todo not found")
    return todo


@router.put("/{todo_id}", response_model=TodoResponse)
def update_todo(todo_id: int, todo: TodoUpdate, db: Session = Depends(get_db)):
    db_todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not db_todo:
        raise HTTPException(status_code=404, detail="Todo not found")
    for key, value in todo.dict(exclude_unset=True).items():
        setattr(db_todo, key, value)
    db.commit()
    db.refresh(db_todo)
    return db_todo


@router.post("/{todo_id}/complete", response_model=TodoResponse)
def complete_todo(todo_id: int, completed_by: str = "system", db: Session = Depends(get_db)):
    db_todo = db.query(Todo).filter(Todo.id == todo_id).first()
    if not db_todo:
        raise HTTPException(status_code=404, detail="Todo not found")
    
    db_todo.status = TodoStatus.COMPLETED
    db_todo.completed_at = datetime.utcnow()
    db_todo.completed_by = completed_by
    db.commit()
    db.refresh(db_todo)
    return db_todo


@router.get("/statistics")
def get_todo_statistics(db: Session = Depends(get_db)):
    pending = db.query(Todo).filter(Todo.status == TodoStatus.PENDING).count()
    completed = db.query(Todo).filter(Todo.status == TodoStatus.COMPLETED).count()
    
    stats = {
        "total_pending": pending,
        "total_completed": completed,
        "by_type": {}
    }
    
    for todo_type in TodoType:
        type_pending = db.query(Todo).filter(
            Todo.todo_type == todo_type,
            Todo.status == TodoStatus.PENDING
        ).count()
        stats["by_type"][todo_type.value] = {
            "pending": type_pending
        }
    
    return stats
