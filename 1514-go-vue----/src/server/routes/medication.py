from datetime import date
from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException

from src.core.models import MedicationTodo
from src.server.deps import get_medication_service

router = APIRouter(prefix="/api", tags=["medication", "todos"])


@router.get("/owners/{owner_id}/medication-todos", response_model=List[MedicationTodo])
def get_pending_todos(
    owner_id: str,
    medication_service=Depends(get_medication_service),
):
    return medication_service.get_pending_todos(owner_id)


@router.post("/medication-todos/{todo_id}/complete", response_model=MedicationTodo)
def complete_todo(
    todo_id: str,
    medication_service=Depends(get_medication_service),
):
    try:
        return medication_service.complete_todo(todo_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/medication-todos/by-date", response_model=List[MedicationTodo])
def get_todos_by_date(
    target_date: date,
    medication_service=Depends(get_medication_service),
):
    return medication_service.get_todos_by_date(target_date)
