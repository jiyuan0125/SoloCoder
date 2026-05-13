from __future__ import annotations

import os
from contextlib import asynccontextmanager
from typing import Optional

from dotenv import load_dotenv
from fastapi import FastAPI, HTTPException, Query, status
from fastapi.responses import JSONResponse

from models import Schedule, ScheduleType, Task, TaskCreate, TaskStatus
from scheduler import EventScheduler
from storage import Storage

load_dotenv()

STORAGE_FILE = os.getenv("STORAGE_FILE", "./tasks.json")
PORT = int(os.getenv("PORT", "9102"))

scheduler: Optional[EventScheduler] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global scheduler
    storage = Storage(STORAGE_FILE)
    scheduler = EventScheduler(storage)
    scheduler.start()
    yield
    if scheduler:
        scheduler.stop()


app = FastAPI(lifespan=lifespan, title="Event Scheduler API")


@app.post("/tasks", response_model=Task, status_code=status.HTTP_201_CREATED)
def create_task(task_create: TaskCreate):
    if not scheduler:
        raise HTTPException(status_code=500, detail="Scheduler not initialized")

    schedule = task_create.get_schedule()

    if schedule.type == ScheduleType.CRON and not schedule.cron:
        raise HTTPException(status_code=400, detail="Cron expression required for cron schedule")
    if schedule.type == ScheduleType.DELAY and not schedule.delay_seconds:
        raise HTTPException(status_code=400, detail="delay_seconds required for delay schedule")

    task = Task(
        name=task_create.name,
        schedule=schedule,
        callback_url=task_create.callback_url,
    )
    return scheduler.create_task(task)


@app.get("/tasks", response_model=list[Task])
def list_tasks(
    status: Optional[TaskStatus] = Query(default=None),
    type: Optional[ScheduleType] = Query(default=None, alias="type"),
):
    if not scheduler:
        raise HTTPException(status_code=500, detail="Scheduler not initialized")
    return scheduler.list_tasks(status=status, schedule_type=type)


@app.get("/tasks/{task_id}")
def get_task(task_id: str):
    if not scheduler:
        raise HTTPException(status_code=500, detail="Scheduler not initialized")

    task = scheduler.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")

    task_dict = task.model_dump(mode="json")
    task_dict["recent_execution_history"] = task_dict.pop("execution_history", [])
    return JSONResponse(content=task_dict)


@app.post("/tasks/{task_id}/retry", response_model=Task)
def retry_task(task_id: str):
    if not scheduler:
        raise HTTPException(status_code=500, detail="Scheduler not initialized")

    task = scheduler.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")

    if not task.can_retry():
        raise HTTPException(status_code=409, detail="Task cannot be retried")

    return scheduler.retry_task(task_id)


@app.post("/tasks/{task_id}/cancel", response_model=Task)
def cancel_task(task_id: str):
    if not scheduler:
        raise HTTPException(status_code=500, detail="Scheduler not initialized")

    task = scheduler.get_task(task_id)
    if not task:
        raise HTTPException(status_code=404, detail="Task not found")

    if not task.can_cancel():
        raise HTTPException(status_code=409, detail="Task cannot be cancelled")

    return scheduler.cancel_task(task_id)


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("main:app", host="0.0.0.0", port=PORT, reload=False)
