from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from sqlalchemy import select
from typing import List
from server.database import get_db
from server.models import PilotTask, TaskStatus
from server.schemas import (
    PilotTaskResponse, PilotTaskDetailResponse,
    SuspensionRequest
)
from server.services import (
    create_pilot_task, start_task, suspend_task, complete_task,
    get_current_weather, update_weather
)

router = APIRouter(
    prefix="/tasks",
    tags=["tasks"]
)


@router.post("/dispatch", response_model=PilotTaskResponse)
def dispatch_task(application_id: int, pilot_id: int, db: Session = Depends(get_db)):
    task = create_pilot_task(db, application_id, pilot_id)
    
    if not task:
        raise HTTPException(status_code=400, detail="调度失败，请检查申请和引航员是否存在")
    
    return task


@router.get("", response_model=List[PilotTaskResponse])
def get_tasks(status: TaskStatus = None, db: Session = Depends(get_db)):
    query = select(PilotTask)
    if status:
        query = query.where(PilotTask.status == status)
    
    return db.execute(query).scalars().all()


@router.get("/{task_id}", response_model=PilotTaskDetailResponse)
def get_task(task_id: int, db: Session = Depends(get_db)):
    task = db.execute(
        select(PilotTask).where(PilotTask.id == task_id)
    ).scalar()
    
    if not task:
        raise HTTPException(status_code=404, detail="任务不存在")
    
    return task


@router.post("/{task_id}/start", response_model=PilotTaskResponse)
def start_pilot_task(task_id: int, db: Session = Depends(get_db)):
    weather = get_current_weather(db)
    if weather and weather.condition == "bad":
        raise HTTPException(status_code=400, detail="当前天气恶劣，无法开始任务")
    
    task = start_task(db, task_id)
    
    if not task:
        raise HTTPException(status_code=400, detail="无法开始任务，请检查任务状态")
    
    return task


@router.post("/{task_id}/suspend", response_model=PilotTaskResponse)
def suspend_pilot_task(task_id: int, request: SuspensionRequest, db: Session = Depends(get_db)):
    task = suspend_task(db, task_id, request.reason)
    
    if not task:
        raise HTTPException(status_code=400, detail="无法中止任务，请检查任务状态")
    
    return task


@router.post("/{task_id}/resume", response_model=PilotTaskResponse)
def resume_pilot_task(task_id: int, db: Session = Depends(get_db)):
    weather = get_current_weather(db)
    if weather and weather.condition == "bad":
        raise HTTPException(status_code=400, detail="当前天气恶劣，无法恢复任务")
    
    task = start_task(db, task_id)
    
    if not task:
        raise HTTPException(status_code=400, detail="无法恢复任务，请检查任务状态")
    
    return task


@router.post("/{task_id}/complete", response_model=PilotTaskResponse)
def complete_pilot_task(task_id: int, db: Session = Depends(get_db)):
    task = complete_task(db, task_id)
    
    if not task:
        raise HTTPException(status_code=400, detail="无法完成任务，请检查任务状态")
    
    return task
