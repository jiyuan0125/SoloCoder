from typing import Dict, Optional
from fastapi import APIRouter, HTTPException

from src.core.utils.scheduler import get_scheduler


router = APIRouter(prefix="/scheduler", tags=["scheduler"])
scheduler = get_scheduler()


@router.get("/status")
def get_status():
    return {
        "running": scheduler.is_running(),
        "tasks": scheduler.list_tasks()
    }


@router.post("/start")
def start_scheduler():
    if scheduler.is_running():
        return {"status": "already running", "tasks": scheduler.list_tasks()}
    scheduler.start()
    return {"status": "started", "tasks": scheduler.list_tasks()}


@router.post("/stop")
def stop_scheduler():
    if not scheduler.is_running():
        return {"status": "already stopped"}
    scheduler.stop()
    return {"status": "stopped"}


@router.post("/tasks/{task_name}")
def add_task(task_name: str, interval_seconds: int):
    if scheduler.is_running():
        raise HTTPException(
            status_code=400,
            detail="Cannot add tasks while scheduler is running"
        )
    success = scheduler.add_interval_task(
        task_name,
        _dummy_task,
        interval_seconds,
        task_name
    )
    if not success:
        raise HTTPException(status_code=400, detail="Task already exists")
    return {"status": "added", "task_name": task_name, "interval_seconds": interval_seconds}


@router.delete("/tasks/{task_name}")
def remove_task(task_name: str):
    if scheduler.is_running():
        raise HTTPException(
            status_code=400,
            detail="Cannot remove tasks while scheduler is running"
        )
    success = scheduler.remove_task(task_name)
    if not success:
        raise HTTPException(status_code=404, detail="Task not found")
    return {"status": "removed", "task_name": task_name}


def _dummy_task(name: str):
    pass
