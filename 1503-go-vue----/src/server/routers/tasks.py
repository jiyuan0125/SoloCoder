from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from src.core import schemas, services
from src.core.database import get_db
from src.core.services import parse_waypoints

router = APIRouter(prefix="/tasks", tags=["tasks"])


@router.post("/", response_model=schemas.PatrolTaskRead)
def create_task(task: schemas.PatrolTaskCreate, db: Session = Depends(get_db)):
    try:
        t = services.create_patrol_task(db, task)
        return schemas.PatrolTaskRead(
            id=t.id,
            area_id=t.area_id,
            ranger_id=t.ranger_id,
            patrol_date=t.patrol_date,
            route_waypoints=parse_waypoints(t.route_waypoints),
            created_at=t.created_at
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/", response_model=List[schemas.PatrolTaskDetail])
def list_tasks(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    tasks = services.get_patrol_tasks(db, skip=skip, limit=limit)
    result = []
    for task in tasks:
        result.append(schemas.PatrolTaskDetail(
            id=task.id,
            area_id=task.area_id,
            ranger_id=task.ranger_id,
            patrol_date=task.patrol_date,
            route_waypoints=parse_waypoints(task.route_waypoints),
            created_at=task.created_at,
            area=schemas.AreaRead.from_orm(task.area) if task.area else None,
            ranger=schemas.UserRead.from_orm(task.ranger) if task.ranger else None
        ))
    return result


@router.get("/{task_id}", response_model=schemas.PatrolTaskDetail)
def get_task(task_id: int, db: Session = Depends(get_db)):
    task = services.get_patrol_task_by_id(db, task_id)
    if not task:
        raise HTTPException(status_code=404, detail="任务不存在")
    return schemas.PatrolTaskDetail(
        id=task.id,
        area_id=task.area_id,
        ranger_id=task.ranger_id,
        patrol_date=task.patrol_date,
        route_waypoints=parse_waypoints(task.route_waypoints),
        created_at=task.created_at,
        area=schemas.AreaRead.from_orm(task.area) if task.area else None,
        ranger=schemas.UserRead.from_orm(task.ranger) if task.ranger else None
    )
