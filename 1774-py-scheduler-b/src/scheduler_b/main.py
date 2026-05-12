from __future__ import annotations
import os
import logging
from contextlib import asynccontextmanager
from typing import Optional

import uvicorn
from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import JSONResponse

from .scheduler import TaskScheduler
from .models import (
    Task, TaskStatus, TaskCreate, TaskUpdate, ExecutionLog, CronValidationResponse
)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger("scheduler_b")

scheduler = TaskScheduler()


@asynccontextmanager
async def lifespan(app: FastAPI):
    await scheduler.start()
    logger.info("Scheduler started successfully")
    yield
    await scheduler.stop()
    logger.info("Scheduler stopped")


app = FastAPI(
    title="统一任务调度器",
    description="基于 FastAPI 和 APScheduler 的统一 Cron 任务调度服务",
    version="0.1.0",
    lifespan=lifespan,
)


@app.get("/health")
async def health_check():
    return {"status": "ok", "service": "scheduler-b"}


@app.get("/api/cron/validate", response_model=CronValidationResponse)
async def validate_cron(
    expression: str = Query(..., description="Cron 表达式"),
    count: int = Query(5, ge=1, le=20, description="预览下次执行的次数"),
):
    return await scheduler.validate_cron(expression, count=count)


@app.post("/api/tasks", response_model=Task, status_code=201)
async def create_task(data: TaskCreate):
    try:
        return await scheduler.create_task(data)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@app.get("/api/tasks", response_model=list[Task])
async def list_tasks(
    tag: Optional[str] = None,
    status: Optional[TaskStatus] = None,
):
    return await scheduler.list_tasks(tag=tag, status=status)


@app.get("/api/tasks/{task_id}", response_model=Task)
async def get_task(task_id: str):
    task = await scheduler.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    return task


@app.patch("/api/tasks/{task_id}", response_model=Task)
async def update_task(task_id: str, data: TaskUpdate):
    try:
        task = await scheduler.update_task(task_id, data)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    return task


@app.delete("/api/tasks/{task_id}", status_code=204)
async def delete_task(task_id: str):
    deleted = await scheduler.delete_task(task_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="Task not found")
    return JSONResponse(content=None, status_code=204)


@app.post("/api/tasks/{task_id}/pause", response_model=Task)
async def pause_task(task_id: str):
    task = await scheduler.pause_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    return task


@app.post("/api/tasks/{task_id}/resume", response_model=Task)
async def resume_task(task_id: str):
    task = await scheduler.resume_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")
    return task


@app.post("/api/tasks/{task_id}/trigger", response_model=ExecutionLog)
async def trigger_task(task_id: str):
    log = await scheduler.trigger_task(task_id)
    if not log:
        raise HTTPException(status_code=404, detail="Task not found")
    return log


@app.get("/api/tasks/{task_id}/logs", response_model=list[ExecutionLog])
async def get_task_logs(
    task_id: str,
    limit: int = Query(50, ge=1, le=100, description="返回日志数量"),
):
    logs = await scheduler.get_execution_logs(task_id, limit=limit)
    return logs


def main():
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run(
        "scheduler_b.main:app",
        host="0.0.0.0",
        port=port,
        reload=False,
        workers=1,
    )


if __name__ == "__main__":
    main()
