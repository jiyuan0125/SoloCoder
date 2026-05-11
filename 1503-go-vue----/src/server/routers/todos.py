from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from src.core import schemas, services
from src.core.database import get_db

router = APIRouter(prefix="/todos", tags=["todos"])


@router.get("/", response_model=List[schemas.TodoDetail])
def list_todos(skip: int = 0, limit: int = 100, overdue_only: bool = False, db: Session = Depends(get_db)):
    todos = services.get_todos(db, skip=skip, limit=limit, overdue_only=overdue_only)
    result = []
    for todo in todos:
        anom_data = None
        if todo.anomaly:
            anom_data = schemas.AnomalyDetail.from_orm(todo.anomaly)
            if todo.anomaly.pest_data:
                anom_data.pest_data = schemas.PestAnomalyRead.from_orm(todo.anomaly.pest_data)
            if todo.anomaly.fire_risk_data:
                anom_data.fire_risk_data = schemas.FireRiskAnomalyRead.from_orm(todo.anomaly.fire_risk_data)

        assignee_data = None
        if todo.assignee:
            assignee_data = schemas.UserRead.from_orm(todo.assignee)

        result.append(schemas.TodoDetail(
            id=todo.id,
            anomaly_id=todo.anomaly_id,
            assignee_id=todo.assignee_id,
            status=todo.status,
            priority=todo.priority,
            due_date=todo.due_date,
            notes=todo.notes,
            created_at=todo.created_at,
            completed_at=todo.completed_at,
            anomaly=anom_data,
            assignee=assignee_data
        ))
    return result


@router.get("/{todo_id}", response_model=schemas.TodoDetail)
def get_todo(todo_id: int, db: Session = Depends(get_db)):
    todo = services.get_todo_by_id(db, todo_id)
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")

    anom_data = None
    if todo.anomaly:
        anom_data = schemas.AnomalyDetail.from_orm(todo.anomaly)
        if todo.anomaly.pest_data:
            anom_data.pest_data = schemas.PestAnomalyRead.from_orm(todo.anomaly.pest_data)
        if todo.anomaly.fire_risk_data:
            anom_data.fire_risk_data = schemas.FireRiskAnomalyRead.from_orm(todo.anomaly.fire_risk_data)

    assignee_data = None
    if todo.assignee:
        assignee_data = schemas.UserRead.from_orm(todo.assignee)

    return schemas.TodoDetail(
        id=todo.id,
        anomaly_id=todo.anomaly_id,
        assignee_id=todo.assignee_id,
        status=todo.status,
        priority=todo.priority,
        due_date=todo.due_date,
        notes=todo.notes,
        created_at=todo.created_at,
        completed_at=todo.completed_at,
        anomaly=anom_data,
        assignee=assignee_data
    )


@router.patch("/{todo_id}", response_model=schemas.TodoDetail)
def update_todo(todo_id: int, update: schemas.TodoUpdate, db: Session = Depends(get_db)):
    todo = services.update_todo(db, todo_id, update)
    if not todo:
        raise HTTPException(status_code=404, detail="待办不存在")

    anom_data = None
    if todo.anomaly:
        anom_data = schemas.AnomalyDetail.from_orm(todo.anomaly)
        if todo.anomaly.pest_data:
            anom_data.pest_data = schemas.PestAnomalyRead.from_orm(todo.anomaly.pest_data)
        if todo.anomaly.fire_risk_data:
            anom_data.fire_risk_data = schemas.FireRiskAnomalyRead.from_orm(todo.anomaly.fire_risk_data)

    assignee_data = None
    if todo.assignee:
        assignee_data = schemas.UserRead.from_orm(todo.assignee)

    return schemas.TodoDetail(
        id=todo.id,
        anomaly_id=todo.anomaly_id,
        assignee_id=todo.assignee_id,
        status=todo.status,
        priority=todo.priority,
        due_date=todo.due_date,
        notes=todo.notes,
        created_at=todo.created_at,
        completed_at=todo.completed_at,
        anomaly=anom_data,
        assignee=assignee_data
    )
