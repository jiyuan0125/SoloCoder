from typing import List
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import Todo, TodoCreate
from server import services

router = APIRouter(prefix="/api/agent-services/{service_id}/todos", tags=["todos"])


@router.post("", response_model=Todo)
def create_todo(
    service_id: int,
    todo_in: TodoCreate,
    db: Session = Depends(get_db),
):
    return services.create_todo(db, service_id, todo_in)


@router.get("", response_model=List[Todo])
def list_todos(service_id: int, db: Session = Depends(get_db)):
    return services.get_todos(db, service_id)


@router.post("/{todo_id}/complete", response_model=Todo)
def complete_todo(
    service_id: int,
    todo_id: int,
    db: Session = Depends(get_db),
):
    return services.complete_todo(db, todo_id)
