from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from src.core.database import get_db
from src.core.services import IntegrityService
from src.server.schemas import (
    IntegrityRecordCreate,
    IntegrityRecordResponse,
    WallThicknessCheckResponse,
    MaintenanceTaskResponse,
    MaintenanceTaskUpdate
)


router = APIRouter(prefix="/integrity", tags=["integrity"])


@router.post("/records/", response_model=IntegrityRecordResponse)
def add_integrity_record(
    record: IntegrityRecordCreate,
    db: Session = Depends(get_db)
):
    service = IntegrityService(db)
    new_record = service.add_integrity_record(record.model_dump())

    check = service.check_wall_thickness(new_record)
    if check["needs_maintenance"]:
        service.create_maintenance_task(new_record)

    return new_record


@router.get("/records/", response_model=List[IntegrityRecordResponse])
def list_records(
    pipeline_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    service = IntegrityService(db)
    return service.list_records(pipeline_id=pipeline_id)


@router.get("/records/{record_id}/check", response_model=WallThicknessCheckResponse)
def check_wall_thickness(record_id: int, db: Session = Depends(get_db)):
    service = IntegrityService(db)
    records = service.list_records()
    record = next((r for r in records if r.id == record_id), None)
    if not record:
        raise HTTPException(status_code=404, detail="Record not found")
    return service.check_wall_thickness(record)


@router.get("/maintenance/", response_model=List[MaintenanceTaskResponse])
def list_maintenance_tasks(
    status: Optional[str] = None,
    db: Session = Depends(get_db)
):
    service = IntegrityService(db)
    return service.list_maintenance_tasks(status=status)


@router.put("/maintenance/{task_id}", response_model=MaintenanceTaskResponse)
def update_maintenance_task(
    task_id: int,
    data: MaintenanceTaskUpdate,
    db: Session = Depends(get_db)
):
    service = IntegrityService(db)
    task = service.update_maintenance_task(task_id, data.model_dump(exclude_unset=True))
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    return task
