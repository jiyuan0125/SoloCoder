from typing import List, Optional
from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.models import ReminderTodo, Borrowing, Collection
from app.schemas import ReminderTodoResponse

router = APIRouter()


@router.get("/", response_model=List[ReminderTodoResponse])
def list_todos(
    is_completed: Optional[bool] = Query(None, description="是否已完成"),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(ReminderTodo)
    if is_completed is not None:
        query = query.filter(ReminderTodo.is_completed == is_completed)

    todos = query.order_by(ReminderTodo.created_at.desc()).offset(skip).limit(limit).all()

    results = []
    for todo in todos:
        borrowing = db.query(Borrowing).filter(Borrowing.id == todo.borrowing_id).first()
        collection = None
        if borrowing:
            collection = db.query(Collection).filter(Collection.id == borrowing.collection_id).first()

        results.append(ReminderTodoResponse(
            id=todo.id,
            borrowing_id=todo.borrowing_id,
            borrower=borrowing.borrower if borrowing else None,
            collection_name=collection.name if collection else None,
            overdue_days=todo.overdue_days,
            message=todo.message,
            is_completed=todo.is_completed,
            created_at=todo.created_at,
            completed_at=todo.completed_at
        ))
    return results


@router.post("/{todo_id}/complete")
def complete_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = db.query(ReminderTodo).filter(ReminderTodo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")

    if todo.is_completed:
        return {"message": "该待办事项已完成", "todo_id": todo_id}

    todo.is_completed = True
    todo.completed_at = datetime.utcnow()
    db.commit()

    return {"message": "待办事项已完成", "todo_id": todo_id}


@router.delete("/{todo_id}", status_code=204)
def delete_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = db.query(ReminderTodo).filter(ReminderTodo.id == todo_id).first()
    if not todo:
        raise HTTPException(status_code=404, detail="待办事项不存在")
    db.delete(todo)
    db.commit()
