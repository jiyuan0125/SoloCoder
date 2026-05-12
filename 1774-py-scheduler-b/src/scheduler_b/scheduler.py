from __future__ import annotations
import asyncio
import httpx
import logging
import uuid
from datetime import datetime
from typing import Optional
from concurrent.futures import ThreadPoolExecutor

from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger
from apscheduler.jobstores.memory import MemoryJobStore
from apscheduler.executors.pool import ThreadPoolExecutor as APThreadPoolExecutor

from .models import (
    Task, TaskStatus, TaskCreate, TaskUpdate, ExecutionLog, CronValidationResponse
)
from .storage import TaskStorage

logger = logging.getLogger(__name__)


def validate_cron_expression(expr: str) -> tuple[bool, str]:
    try:
        trigger = CronTrigger.from_crontab(expr)
        return True, f"Valid cron expression: {expr}"
    except Exception as e:
        return False, f"Invalid cron expression: {e}"


class TaskScheduler:
    def __init__(self):
        self.storage = TaskStorage()
        jobstores = {"default": MemoryJobStore()}
        executors = {
            "default": APThreadPoolExecutor(max_workers=20),
        }
        job_defaults = {"coalesce": True, "max_instances": 1}
        self.scheduler = AsyncIOScheduler(
            jobstores=jobstores,
            executors=executors,
            job_defaults=job_defaults,
            timezone="UTC",
        )
        self._http_client: Optional[httpx.AsyncClient] = None
        self._http_lock = asyncio.Lock()
        self._thread_pool = ThreadPoolExecutor(max_workers=10)

    async def _get_http_client(self) -> httpx.AsyncClient:
        async with self._http_lock:
            if self._http_client is None:
                self._http_client = httpx.AsyncClient(timeout=60.0)
            return self._http_client

    async def start(self):
        if not self.scheduler.running:
            self.scheduler.start()

    async def stop(self):
        if self.scheduler.running:
            self.scheduler.shutdown(wait=False)
        if self._http_client:
            await self._http_client.aclose()
            self._http_client = None
        self._thread_pool.shutdown(wait=False)

    def _parse_cron(self, expr: str) -> CronTrigger:
        return CronTrigger.from_crontab(expr)

    async def validate_cron(self, expr: str, count: int = 5) -> CronValidationResponse:
        valid, msg = validate_cron_expression(expr)
        if not valid:
            return CronValidationResponse(valid=False, message=msg)
        try:
            trigger = self._parse_cron(expr)
            next_runs = []
            now = datetime.utcnow()
            current = now
            for _ in range(count):
                next_run = trigger.get_next_fire_time(current, current)
                if next_run:
                    next_runs.append(next_run)
                    current = next_run
            return CronValidationResponse(valid=True, message=msg, next_runs=next_runs)
        except Exception as e:
            return CronValidationResponse(valid=False, message=str(e))

    async def _execute_callback(self, task: Task, execution_id: str):
        client = await self._get_http_client()
        started_at = datetime.utcnow()
        try_count = task.current_retry_count + 1
        log = ExecutionLog(
            id=execution_id,
            task_id=task.id,
            started_at=started_at,
            status=TaskStatus.RUNNING,
            try_count=try_count,
        )
        self.storage.add_execution_log(log)
        self.storage.update_task(
            task.id,
            status=TaskStatus.RUNNING,
            last_run_at=started_at,
            total_runs=task.total_runs + 1,
            current_retry_count=try_count,
        )

        headers = task.callback_headers or {}
        headers.setdefault("Content-Type", "application/json")
        payload = {
            "task_id": task.id,
            "task_name": task.name,
            "execution_id": execution_id,
            "try_count": try_count,
            "triggered_at": started_at.isoformat(),
        }

        try:
            response = await client.request(
                method=task.callback_method,
                url=task.callback_url,
                json=payload,
                headers=headers,
                timeout=task.timeout_seconds,
            )
            finished_at = datetime.utcnow()
            duration = (finished_at - started_at).total_seconds()
            status = TaskStatus.SUCCESS if response.status_code < 400 else TaskStatus.FAILED
            result = {
                "status_code": response.status_code,
                "body": response.text[:2000],
            }
            error = None if status == TaskStatus.SUCCESS else f"HTTP {response.status_code}"
        except Exception as e:
            finished_at = datetime.utcnow()
            duration = (finished_at - started_at).total_seconds()
            status = TaskStatus.FAILED
            result = None
            error = str(e)

        self.storage.update_task(
            task.id,
            status=status,
            last_success_at=finished_at if status == TaskStatus.SUCCESS else task.last_success_at,
            last_failure_at=finished_at if status == TaskStatus.FAILED else task.last_failure_at,
            total_successes=task.total_successes + (1 if status == TaskStatus.SUCCESS else 0),
            total_failures=task.total_failures + (1 if status == TaskStatus.FAILED else 0),
        )

        finished_log = self.storage.get_execution_logs(task.id, limit=1)[0]
        updated_log = finished_log.model_copy(update={
            "finished_at": finished_at,
            "status": status,
            "result": result,
            "error": error,
            "duration_seconds": duration,
        })
        self.storage.add_execution_log(updated_log)

        if status == TaskStatus.FAILED and try_count <= task.max_retries:
            asyncio.create_task(self._retry_task(task, try_count))

    async def _retry_task(self, task: Task, last_try: int):
        await asyncio.sleep(task.retry_delay_seconds)
        current = self.storage.get_task(task.id)
        if not current or not current.enabled:
            return
        execution_id = str(uuid.uuid4())
        await self._execute_callback(current, execution_id)

    def _build_job_func(self, task_id: str):
        async def job_wrapper():
            task = self.storage.get_task(task_id)
            if not task or not task.enabled:
                return
            execution_id = str(uuid.uuid4())
            await self._execute_callback(task, execution_id)
        return job_wrapper

    async def create_task(self, data: TaskCreate) -> Task:
        valid, msg = validate_cron_expression(data.cron_expression)
        if not valid:
            raise ValueError(msg)

        now = datetime.utcnow()
        task = Task(
            id=str(uuid.uuid4()),
            name=data.name,
            cron_expression=data.cron_expression,
            callback_url=data.callback_url,
            callback_type=data.callback_type,
            callback_headers=data.callback_headers,
            callback_method=data.callback_method,
            max_retries=data.max_retries,
            retry_delay_seconds=data.retry_delay_seconds,
            timeout_seconds=data.timeout_seconds,
            description=data.description,
            tags=data.tags or [],
            enabled=data.enabled,
            status=TaskStatus.PAUSED if not data.enabled else TaskStatus.PENDING,
            created_at=now,
            updated_at=now,
        )

        self.storage.add_task(task)

        if task.enabled:
            await self._schedule_task(task)

        return task

    async def _schedule_task(self, task: Task):
        trigger = self._parse_cron(task.cron_expression)
        self.scheduler.add_job(
            func=self._build_job_func(task.id),
            trigger=trigger,
            id=task.id,
            replace_existing=True,
        )
        next_run = trigger.get_next_fire_time(datetime.utcnow(), datetime.utcnow())
        self.storage.update_task(task.id, next_run_at=next_run)

    async def get_task(self, task_id: str) -> Optional[Task]:
        task = self.storage.get_task(task_id)
        if not task:
            return None
        updated = await self._update_next_run(task)
        return updated

    async def _update_next_run(self, task: Task) -> Task:
        try:
            job = self.scheduler.get_job(task.id)
            next_run = job.next_run_time if job else None
            if next_run != task.next_run_at:
                self.storage.update_task(task.id, next_run_at=next_run)
                task = self.storage.get_task(task.id) or task
        except Exception:
            pass
        return task

    async def list_tasks(self, tag: Optional[str] = None, status: Optional[TaskStatus] = None) -> list[Task]:
        tasks = self.storage.list_tasks(tag=tag, status=status)
        return [await self._update_next_run(t) for t in tasks]

    async def update_task(self, task_id: str, data: TaskUpdate) -> Optional[Task]:
        task = self.storage.get_task(task_id)
        if not task:
            return None

        updates = {k: v for k, v in data.model_dump().items() if v is not None}
        if not updates:
            return task

        if "cron_expression" in updates:
            valid, msg = validate_cron_expression(updates["cron_expression"])
            if not valid:
                raise ValueError(msg)

        was_enabled = task.enabled
        will_enable = updates.get("enabled", task.enabled)
        cron_changed = "cron_expression" in updates

        if was_enabled:
            try:
                self.scheduler.remove_job(task_id)
            except Exception:
                pass

        for k, v in updates.items():
            setattr(task, k, v)

        updated = self.storage.update_task(task_id, **updates)
        if not updated:
            return None

        if will_enable or cron_changed:
            if will_enable:
                await self._schedule_task(updated)
                self.storage.update_task(task_id, status=TaskStatus.PENDING)
            else:
                self.storage.update_task(task_id, status=TaskStatus.PAUSED)

        return self.storage.get_task(task_id)

    async def delete_task(self, task_id: str) -> bool:
        try:
            self.scheduler.remove_job(task_id)
        except Exception:
            pass
        return self.storage.delete_task(task_id)

    async def pause_task(self, task_id: str) -> Optional[Task]:
        task = self.storage.get_task(task_id)
        if not task:
            return None
        try:
            self.scheduler.pause_job(task_id)
        except Exception:
            pass
        self.storage.update_task(task_id, enabled=False, status=TaskStatus.PAUSED)
        return self.storage.get_task(task_id)

    async def resume_task(self, task_id: str) -> Optional[Task]:
        task = self.storage.get_task(task_id)
        if not task:
            return None
        try:
            job = self.scheduler.get_job(task_id)
            if job:
                self.scheduler.resume_job(task_id)
            else:
                await self._schedule_task(task)
        except Exception:
            await self._schedule_task(task)
        self.storage.update_task(task_id, enabled=True, status=TaskStatus.PENDING)
        return self.storage.get_task(task_id)

    async def trigger_task(self, task_id: str) -> Optional[ExecutionLog]:
        task = self.storage.get_task(task_id)
        if not task:
            return None
        execution_id = str(uuid.uuid4())
        asyncio.create_task(self._execute_callback(task, execution_id))
        return ExecutionLog(
            id=execution_id,
            task_id=task_id,
            started_at=datetime.utcnow(),
            status=TaskStatus.RUNNING,
            try_count=0,
        )

    async def get_execution_logs(self, task_id: str, limit: int = 50) -> list[ExecutionLog]:
        return self.storage.get_execution_logs(task_id, limit=limit)
