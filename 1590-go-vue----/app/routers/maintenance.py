from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from datetime import datetime
from app.database import get_db
from app.models import MaintenanceTask, MaintenanceStatus
from app.schemas import (
    MaintenanceStartResponse, MaintenanceCompleteRequest, MaintenanceCompleteResponse
)

router = APIRouter(prefix="/maintenance", tags=["维护管理"])


def get_task_or_404(db: Session, task_id: int) -> MaintenanceTask:
    task = db.query(MaintenanceTask).filter(MaintenanceTask.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail=f"维护任务 {task_id} 不存在")
    return task


@router.post("/{task_id}/start", response_model=MaintenanceStartResponse)
def start_maintenance(
    task_id: int,
    db: Session = Depends(get_db)
):
    task = get_task_or_404(db, task_id)
    
    if task.status == MaintenanceStatus.COMPLETED:
        raise HTTPException(status_code=400, detail="该维护任务已完成")
    
    if task.status == MaintenanceStatus.IN_PROGRESS:
        return MaintenanceStartResponse(
            task_id=task_id,
            status=MaintenanceStatus.IN_PROGRESS,
            start_time=task.start_time,
            message="维护任务已在进行中"
        )
    
    now = datetime.utcnow()
    task.status = MaintenanceStatus.IN_PROGRESS
    task.start_time = now
    
    db.commit()
    db.refresh(task)
    
    return MaintenanceStartResponse(
        task_id=task_id,
        status=MaintenanceStatus.IN_PROGRESS,
        start_time=now,
        message=f"开始执行 {task.task_type.value} 维护任务"
    )


@router.post("/{task_id}/complete", response_model=MaintenanceCompleteResponse)
def complete_maintenance(
    task_id: int,
    request: MaintenanceCompleteRequest,
    db: Session = Depends(get_db)
):
    task = get_task_or_404(db, task_id)
    
    if task.status == MaintenanceStatus.COMPLETED:
        raise HTTPException(status_code=400, detail="该维护任务已完成")
    
    if task.status not in [MaintenanceStatus.IN_PROGRESS, MaintenanceStatus.OVERDUE, MaintenanceStatus.PENDING]:
        raise HTTPException(status_code=400, detail=f"当前状态 {task.status.value} 无法完成")
    
    now = datetime.utcnow()
    task.status = MaintenanceStatus.COMPLETED
    task.complete_time = now
    
    if request.notes:
        task.notes = request.notes
    
    db.commit()
    db.refresh(task)
    
    return MaintenanceCompleteResponse(
        task_id=task_id,
        status=MaintenanceStatus.COMPLETED,
        complete_time=now,
        message=f"{task.task_type.value} 维护任务完成"
    )
