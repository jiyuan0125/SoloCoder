from __future__ import annotations

import asyncio
from datetime import datetime, timedelta
from typing import Optional
from uuid import uuid4

import httpx
from apscheduler.schedulers.asyncio import AsyncIOScheduler

from models import (
    ExecutionHistory,
    ScheduleType,
    Task,
    TaskStatus,
)
from storage import Storage


class TaskExecutor:
    CALLBACK_TIMEOUT = 30

    async def execute(self, task: Task) -> tuple[bool, Optional[str]]:
        start_time = datetime.now()
        try:
            async with httpx.AsyncClient(timeout=self.CALLBACK_TIMEOUT) as client:
                response = await client.post(
                    task.callback_url,
                    json={
                        "task_id": task.id,
                        "name": task.name,
                        "executed_at": start_time.isoformat(),
                    },
                )
                if response.status_code < 400:
                    return True, None
                return False, f"HTTP {response.status_code}: {response.text}"
        except httpx.TimeoutException:
            return False, "Callback timeout"
        except Exception as e:
            return False, str(e)


class EventScheduler:
    def __init__(self, storage: Storage):
        self.storage = storage
        self.scheduler = AsyncIOScheduler()
        self.executor = TaskExecutor()
        self.tasks: dict[str, Task] = {}
        self._apscheduler_jobs: dict[str, str] = {}
        self._async_tasks: dict[str, asyncio.Task] = {}

    def load_tasks(self) -> None:
        raw_tasks = self.storage.load()
        for raw in raw_tasks:
            try:
                task = Task.model_validate(raw)
                if task.status == TaskStatus.RUNNING:
                    task.status = TaskStatus.PENDING
                if task.status in (TaskStatus.PENDING, TaskStatus.RUNNING):
                    self._register_task_with_scheduler(task)
                self.tasks[task.id] = task
            except Exception:
                continue

    def _save(self) -> None:
        self.storage.save(list(self.tasks.values()))

    def _register_task_with_scheduler(self, task: Task) -> None:
        if task.id in self._apscheduler_jobs:
            return

        if task.is_cron() and task.schedule.cron:
            job = self.scheduler.add_job(
                self._run_task_cron,
                "cron",
                args=[task.id],
                id=f"cron_{task.id}",
                **self._parse_cron(task.schedule.cron),
            )
            self._apscheduler_jobs[task.id] = job.id
        elif task.is_delay() and task.schedule.delay_seconds:
            run_date = datetime.now() + timedelta(seconds=task.schedule.delay_seconds)
            job = self.scheduler.add_job(
                self._run_task_delay,
                "date",
                run_date=run_date,
                args=[task.id],
                id=f"delay_{task.id}",
            )
            self._apscheduler_jobs[task.id] = job.id

    def _parse_cron(self, cron_expr: str) -> dict:
        parts = cron_expr.split()
        if len(parts) < 5:
            parts = parts + ["*"] * (5 - len(parts))
        return {
            "minute": parts[0],
            "hour": parts[1],
            "day": parts[2],
            "month": parts[3],
            "day_of_week": parts[4],
        }

    def _unregister_task_from_scheduler(self, task_id: str) -> None:
        job_id = self._apscheduler_jobs.pop(task_id, None)
        if job_id and self.scheduler.get_job(job_id):
            self.scheduler.remove_job(job_id)

        if task_id in self._async_tasks:
            self._async_tasks[task_id].cancel()
            self._async_tasks.pop(task_id, None)

    def create_task(self, task: Task) -> Task:
        self._register_task_with_scheduler(task)
        self.tasks[task.id] = task
        self._save()
        return task

    def get_task(self, task_id: str) -> Optional[Task]:
        return self.tasks.get(task_id)

    def list_tasks(
        self,
        status: Optional[TaskStatus] = None,
        schedule_type: Optional[ScheduleType] = None,
    ) -> list[Task]:
        result = list(self.tasks.values())
        if status:
            result = [t for t in result if t.status == status]
        if schedule_type:
            result = [t for t in result if t.schedule.type == schedule_type]
        return result

    def retry_task(self, task_id: str) -> Optional[Task]:
        task = self.get_task(task_id)
        if not task or not task.can_retry():
            return None
        task.status = TaskStatus.PENDING
        self._register_task_with_scheduler(task)
        self._save()
        return task

    def cancel_task(self, task_id: str) -> Optional[Task]:
        task = self.get_task(task_id)
        if not task or not task.can_cancel():
            return None
        task.status = TaskStatus.CANCELLED
        self._unregister_task_from_scheduler(task_id)
        self._save()
        return task

    async def _run_task_cron(self, task_id: str) -> None:
        task = self.get_task(task_id)
        if not task or task.status == TaskStatus.CANCELLED:
            return

        new_instance = Task(
            id=uuid4().hex,
            name=task.name,
            schedule=task.schedule,
            callback_url=task.callback_url,
            status=TaskStatus.PENDING,
        )
        self.tasks[new_instance.id] = new_instance
        self._save()

        await self._execute_task(new_instance.id)

    async def _run_task_delay(self, task_id: str) -> None:
        await self._execute_task(task_id)
        self._apscheduler_jobs.pop(task_id, None)

    async def _execute_task(self, task_id: str) -> None:
        task = self.get_task(task_id)
        if not task or task.status != TaskStatus.PENDING:
            return

        task.status = TaskStatus.RUNNING
        task.last_run_at = datetime.now()
        self._save()

        start_time = datetime.now()
        success, error = await self.executor.execute(task)
        end_time = datetime.now()
        duration_ms = int((end_time - start_time).total_seconds() * 1000)

        task.status = TaskStatus.SUCCESS if success else TaskStatus.FAILED

        history = ExecutionHistory(
            executed_at=start_time,
            status=TaskStatus.SUCCESS if success else TaskStatus.FAILED,
            duration_ms=duration_ms,
            error=error,
        )
        task.execution_history.append(history)

        if len(task.execution_history) > 5:
            task.execution_history = task.execution_history[-5:]

        self._save()

    def start(self) -> None:
        self.load_tasks()
        self.scheduler.start()

    def stop(self) -> None:
        for task_id in list(self._async_tasks.keys()):
            self._async_tasks[task_id].cancel()
        self.scheduler.shutdown(wait=False)
