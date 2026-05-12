from fastapi import APIRouter, HTTPException, status
from typing import List
from .models import (
    TaskCreate,
    Task,
    TaskStatusResponse,
    TaskType,
    ExecutionHistoryResponse,
    TaskStatus,
)
from .storage import storage
from .scheduler import calculate_next_run


router = APIRouter()


@router.post("/tasks", response_model=TaskStatusResponse, status_code=status.HTTP_201_CREATED)
def create_task(task_in: TaskCreate):
    if task_in.task_type == TaskType.CRON:
        if not task_in.cron_expression:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Cron expression is required for cron tasks",
            )
        try:
            from crontab import CronTab
            CronTab(task_in.cron_expression)
        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Invalid cron expression: {str(e)}",
            )

    if task_in.task_type == TaskType.DELAYED:
        if not task_in.execute_at:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail="Execute_at is required for delayed tasks",
            )

    task = Task(
        name=task_in.name,
        task_type=task_in.task_type,
        callback_url=task_in.callback_url,
        payload=task_in.payload,
        cron_expression=task_in.cron_expression,
        execute_at=task_in.execute_at,
    )

    task.next_run_at = calculate_next_run(task)

    if not storage.add_task(task):
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=f"Task with name '{task_in.name}' already exists",
        )

    return TaskStatusResponse(**task.dict())


@router.get("/tasks/{task_name}", response_model=TaskStatusResponse)
def get_task_status(task_name: str):
    task = storage.get_task_by_name(task_name)
    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Task '{task_name}' not found",
        )
    return TaskStatusResponse(**task.dict())


@router.get("/tasks/{task_name}/history", response_model=ExecutionHistoryResponse)
def get_task_execution_history(task_name: str):
    task = storage.get_task_by_name(task_name)
    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Task '{task_name}' not found",
        )
    records = storage.get_execution_history(task.id)
    return ExecutionHistoryResponse(
        task_name=task_name,
        total_count=len(records),
        records=records,
    )


@router.post("/tasks/{task_name}/pause", response_model=TaskStatusResponse)
def pause_task(task_name: str):
    task = storage.get_task_by_name(task_name)
    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Task '{task_name}' not found",
        )
    paused = storage.pause_task(task.id)
    if not paused:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to pause task",
        )
    return TaskStatusResponse(**paused.dict())


@router.post("/tasks/{task_name}/resume", response_model=TaskStatusResponse)
def resume_task(task_name: str):
    task = storage.get_task_by_name(task_name)
    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Task '{task_name}' not found",
        )
    if task.status == TaskStatus.DEAD_LETTER:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Cannot resume a dead letter task",
        )
    resumed = storage.resume_task(task.id)
    if not resumed:
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail="Failed to resume task",
        )
    return TaskStatusResponse(**resumed.dict())


@router.get("/tasks", response_model=List[TaskStatusResponse])
def list_tasks():
    tasks = storage.get_all_tasks()
    return [TaskStatusResponse(**task.dict()) for task in tasks]
