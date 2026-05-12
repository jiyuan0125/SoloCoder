from __future__ import annotations

from uuid import UUID

from fastapi import APIRouter, HTTPException, status

from app.executor import start_workers
from app.models import CreateTaskRequest, Task, TaskResponse, TaskStatus
from app.store import task_store

router = APIRouter(prefix="/tasks", tags=["tasks"])


@router.post("", status_code=status.HTTP_202_ACCEPTED, response_model=TaskResponse)
async def create_task(request: CreateTaskRequest) -> TaskResponse:
    if not request.idempotency_key:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="idempotency_key must not be empty",
        )
    if len(request.idempotency_key) > 128:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="idempotency_key exceeds 128 characters",
        )

    existing = await task_store.get_task_by_idempotency_key(request.idempotency_key)
    if existing is not None:
        return TaskResponse.model_validate(existing)

    task = Task(
        url=request.url,
        method=request.method,
        body=request.body,
        idempotency_key=request.idempotency_key,
        retry_config=request.retry_config,
        status=TaskStatus.PENDING,
    )

    registered = await task_store.register_idempotency_key(request.idempotency_key, task.id)
    if not registered:
        existing = await task_store.get_task_by_idempotency_key(request.idempotency_key)
        if existing is not None:
            return TaskResponse.model_validate(existing)

    await task_store.add_task(task)
    return TaskResponse.model_validate(task)


@router.get("/queue", response_model=list[TaskResponse])
async def get_retry_queue() -> list[TaskResponse]:
    retrying = await task_store.get_retrying_tasks()
    return [TaskResponse.model_validate(t) for t in retrying]


@router.get("/{task_id}", response_model=TaskResponse)
async def get_task(task_id: UUID) -> TaskResponse:
    task = await task_store.get_task(task_id)
    if task is None:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Task not found",
        )
    return TaskResponse.model_validate(task)
