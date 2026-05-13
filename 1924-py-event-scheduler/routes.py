import os
import uuid
from contextlib import asynccontextmanager
from datetime import datetime
from typing import List
from fastapi import FastAPI, HTTPException, status
from pydantic import BaseModel

from models import (
    Task, TaskStatus, ExecutionInstance, TaskCreate,
    ScheduleType, ExecutorType
)
from scheduler import TaskStore, Executor, Scheduler


store = TaskStore()
executor = Executor()
scheduler = Scheduler(store, executor)


@asynccontextmanager
async def lifespan(app: FastAPI):
    scheduler.start()
    yield
    await scheduler.stop()


app = FastAPI(title="Event Scheduler", lifespan=lifespan)


class TaskDetailResponse(BaseModel):
    task: Task
    executions: List[ExecutionInstance]


class TaskCancelResponse(BaseModel):
    success: bool
    message: str


@app.post("/tasks", response_model=Task, status_code=status.HTTP_201_CREATED)
async def create_task(task_create: TaskCreate):
    for dep_id in task_create.depends_on:
        if not store.get_task(dep_id):
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Dependency task '{dep_id}' does not exist"
            )

    task = Task(
        id=str(uuid.uuid4()),
        name=task_create.name,
        schedule=task_create.schedule,
        executor_type=task_create.executor_type,
        executor_config=task_create.executor_config,
        depends_on=task_create.depends_on,
        timeout_seconds=task_create.timeout_seconds,
        created_at=datetime.now(),
        canceled=False
    )
    store.add_task(task)
    return task


@app.get("/tasks/{task_id}", response_model=TaskDetailResponse)
async def get_task(task_id: str):
    task = store.get_task(task_id)
    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Task '{task_id}' not found"
        )
    executions = store.get_executions(task_id)
    return TaskDetailResponse(task=task, executions=executions)


@app.post("/tasks/{task_id}/cancel", response_model=TaskCancelResponse)
async def cancel_task(task_id: str):
    task = store.get_task(task_id)
    if not task:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Task '{task_id}' not found"
        )

    if task.canceled:
        return TaskCancelResponse(success=True, message="Task is already canceled")

    last_exec = store.get_last_execution(task_id)
    if last_exec and last_exec.status == TaskStatus.RUNNING:
        return TaskCancelResponse(success=False, message="Cannot cancel running task")

    store.cancel_task(task_id)
    return TaskCancelResponse(success=True, message="Task canceled successfully")


@app.get("/tasks", response_model=List[Task])
async def list_tasks():
    return store.get_all_tasks()
