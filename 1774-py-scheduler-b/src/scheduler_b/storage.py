from __future__ import annotations
from datetime import datetime
from threading import RLock
from typing import Optional, Iterable
from collections import deque

from .models import Task, ExecutionLog, TaskStatus


class TaskStorage:
    def __init__(self, max_logs_per_task: int = 100):
        self._tasks: dict[str, Task] = {}
        self._task_logs: dict[str, deque[ExecutionLog]] = {}
        self._lock = RLock()
        self._max_logs_per_task = max_logs_per_task

    def add_task(self, task: Task) -> Task:
        with self._lock:
            self._tasks[task.id] = task
            self._task_logs[task.id] = deque(maxlen=self._max_logs_per_task)
            return task

    def get_task(self, task_id: str) -> Optional[Task]:
        with self._lock:
            return self._tasks.get(task_id)

    def list_tasks(self, tag: Optional[str] = None, status: Optional[TaskStatus] = None) -> list[Task]:
        with self._lock:
            tasks = list(self._tasks.values())
            if tag:
                tasks = [t for t in tasks if tag in t.tags]
            if status:
                tasks = [t for t in tasks if t.status == status]
            return tasks

    def update_task(self, task_id: str, **updates) -> Optional[Task]:
        with self._lock:
            task = self._tasks.get(task_id)
            if not task:
                return None
            updates["updated_at"] = datetime.utcnow()
            updated = task.model_copy(update=updates)
            self._tasks[task_id] = updated
            return updated

    def delete_task(self, task_id: str) -> bool:
        with self._lock:
            if task_id in self._tasks:
                del self._tasks[task_id]
                self._task_logs.pop(task_id, None)
                return True
            return False

    def add_execution_log(self, log: ExecutionLog) -> None:
        with self._lock:
            logs = self._task_logs.get(log.task_id)
            if logs is not None:
                logs.append(log)

    def get_execution_logs(self, task_id: str, limit: int = 50) -> list[ExecutionLog]:
        with self._lock:
            logs = self._task_logs.get(task_id)
            if not logs:
                return []
            return list(logs)[-limit:]

    def get_all_tasks(self) -> Iterable[Task]:
        with self._lock:
            return list(self._tasks.values())
