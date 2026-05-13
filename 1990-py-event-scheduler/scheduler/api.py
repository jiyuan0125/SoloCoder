from typing import List

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from .models import (
    ExecutionHistory,
    HistoryResponse,
    Task,
    TaskCreate,
    TaskResponse,
    TaskStatus,
    TaskType,
    TaskUpdate,
    get_db,
)
from .scheduler import scheduler

router = APIRouter(prefix="/api", tags=["tasks"])


@router.post("/tasks", response_model=TaskResponse)
def create_task(task_data: TaskCreate, db: Session = Depends(get_db)):
    if task_data.task_type == TaskType.CRON and not task_data.cron_expression:
        raise HTTPException(status_code=400, detail="Cron 类型任务需要 cron_expression")
    if task_data.task_type == TaskType.DELAYED and not task_data.delay_seconds:
        raise HTTPException(status_code=400, detail="延迟任务需要 delay_seconds")

    task = Task(
        name=task_data.name,
        task_type=task_data.task_type,
        cron_expression=task_data.cron_expression,
        delay_seconds=task_data.delay_seconds,
        callback_url=task_data.callback_url,
        callback_payload=task_data.callback_payload,
        timeout_seconds=task_data.timeout_seconds,
        max_retries=task_data.max_retries,
        max_queue_size=task_data.max_queue_size,
    )

    task.next_run_at = scheduler.calculate_next_run(task)
    if task.task_type == TaskType.DELAYED:
        task.status = TaskStatus.PENDING

    db.add(task)
    db.commit()
    db.refresh(task)
    return task


@router.get("/tasks", response_model=List[TaskResponse])
def list_tasks(db: Session = Depends(get_db)):
    tasks = db.query(Task).order_by(Task.id.desc()).all()
    return tasks


@router.get("/tasks/{task_id}", response_model=TaskResponse)
def get_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="任务不存在")
    return task


@router.put("/tasks/{task_id}", response_model=TaskResponse)
def update_task(task_id: int, task_data: TaskUpdate, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="任务不存在")

    update_data = task_data.model_dump(exclude_unset=True)

    for key, value in update_data.items():
        setattr(task, key, value)

    if task.task_type == TaskType.CRON and "cron_expression" in update_data:
        task.next_run_at = scheduler.calculate_next_run(task)

    db.commit()
    db.refresh(task)
    return task


@router.delete("/tasks/{task_id}")
def delete_task(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="任务不存在")

    db.delete(task)
    db.commit()
    return {"message": "任务已删除"}


@router.post("/tasks/{task_id}/run")
def run_task_now(task_id: int, db: Session = Depends(get_db)):
    task = db.query(Task).filter(Task.id == task_id).first()
    if not task:
        raise HTTPException(status_code=404, detail="任务不存在")

    import asyncio

    loop = asyncio.get_event_loop()
    loop.create_task(scheduler._execute_task(task_id, 0))
    return {"message": "任务已触发执行"}


@router.get("/tasks/{task_id}/history", response_model=List[HistoryResponse])
def get_task_history(task_id: int, limit: int = 50, db: Session = Depends(get_db)):
    history = (
        db.query(ExecutionHistory)
        .filter(ExecutionHistory.task_id == task_id)
        .order_by(ExecutionHistory.id.desc())
        .limit(limit)
        .all()
    )
    return history


@router.get("/history", response_model=List[HistoryResponse])
def list_all_history(limit: int = 100, db: Session = Depends(get_db)):
    history = (
        db.query(ExecutionHistory)
        .order_by(ExecutionHistory.id.desc())
        .limit(limit)
        .all()
    )
    return history
