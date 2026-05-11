from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from pydantic import BaseModel
from server.database import get_db
from core.models import TodoResponse
from core import services

router = APIRouter(prefix="/todos", tags=["todos"])


class TodoStatusUpdate(BaseModel):
    status: str


@router.post("/generate-maintenance")
def generate_maintenance(db: Session = Depends(get_db)):
    services.generate_maintenance_todos(db)
    return {"message": "Maintenance todos generated"}


@router.post("/generate-compliance")
def generate_compliance(db: Session = Depends(get_db)):
    services.generate_daily_compliance_todos(db)
    return {"message": "Compliance todos generated"}


@router.get("", response_model=List[TodoResponse])
def list_all(
    facility_id: Optional[int] = None,
    status: Optional[str] = None,
    db: Session = Depends(get_db)
):
    return services.list_todos(db, facility_id, status)


@router.patch("/{todo_id}/status", response_model=TodoResponse)
def update_status(
    todo_id: int,
    update: TodoStatusUpdate,
    db: Session = Depends(get_db)
):
    todo = services.update_todo_status(db, todo_id, update.status)
    if not todo:
        raise HTTPException(status_code=404, detail="Todo not found")
    return todo
