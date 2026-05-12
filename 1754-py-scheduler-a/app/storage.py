from typing import Dict, List, Optional
from threading import Lock
from datetime import datetime
from .models import Task, ExecutionRecord, TaskStatus, ExecutionStatus


def _now_utc() -> datetime:
    return datetime.utcnow()


class TaskStorage:
    def __init__(self):
        self._tasks: Dict[str, Task] = {}
        self._tasks_by_name: Dict[str, Task] = {}
        self._execution_history: Dict[str, List[ExecutionRecord]] = {}
        self._lock = Lock()

    def add_task(self, task: Task) -> bool:
        with self._lock:
            if task.name in self._tasks_by_name:
                return False
            self._tasks[task.id] = task
            self._tasks_by_name[task.name] = task
            self._execution_history[task.id] = []
            return True

    def get_task_by_id(self, task_id: str) -> Optional[Task]:
        with self._lock:
            return self._tasks.get(task_id)

    def get_task_by_name(self, name: str) -> Optional[Task]:
        with self._lock:
            return self._tasks_by_name.get(name)

    def get_all_tasks(self) -> List[Task]:
        with self._lock:
            return list(self._tasks.values())

    def update_task(self, task: Task) -> None:
        with self._lock:
            task.updated_at = _now_utc()
            self._tasks[task.id] = task
            self._tasks_by_name[task.name] = task

    def pause_task(self, task_id: str) -> Optional[Task]:
        with self._lock:
            task = self._tasks.get(task_id)
            if not task:
                return None
            task.is_paused = True
            if task.status != TaskStatus.DEAD_LETTER:
                task.status = TaskStatus.PAUSED
            task.updated_at = _now_utc()
            return task

    def resume_task(self, task_id: str) -> Optional[Task]:
        with self._lock:
            task = self._tasks.get(task_id)
            if not task:
                return None
            task.is_paused = False
            if task.status == TaskStatus.PAUSED:
                task.status = TaskStatus.PENDING
            task.updated_at = _now_utc()
            return task

    def add_execution_record(self, record: ExecutionRecord) -> None:
        with self._lock:
            if record.task_id not in self._execution_history:
                self._execution_history[record.task_id] = []
            self._execution_history[record.task_id].append(record)

    def get_execution_history(self, task_id: str) -> List[ExecutionRecord]:
        with self._lock:
            return list(self._execution_history.get(task_id, []))

    def get_due_tasks(self, now: datetime) -> List[Task]:
        with self._lock:
            due_tasks = []
            for task in self._tasks.values():
                if task.is_paused:
                    continue
                if task.status in (TaskStatus.DEAD_LETTER, TaskStatus.SUCCESS):
                    continue
                if task.next_run_at and task.next_run_at <= now:
                    due_tasks.append(task)
            return due_tasks


storage = TaskStorage()
