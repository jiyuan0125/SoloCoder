from __future__ import annotations

import asyncio
from datetime import datetime, timedelta
from typing import Dict, Optional, Set
from uuid import UUID

from app.models import Task, TaskStatus


class TaskStore:
    def __init__(self) -> None:
        self._tasks: Dict[UUID, Task] = {}
        self._pending_queue: asyncio.Queue[UUID] = asyncio.Queue()
        self._idempotency_map: Dict[str, UUID] = {}
        self._idempotency_expiry: Dict[str, datetime] = {}
        self._lock: asyncio.Lock = asyncio.Lock()
        self._retrying_task_ids: Set[UUID] = set()

    async def add_task(self, task: Task) -> None:
        async with self._lock:
            self._tasks[task.id] = task
            await self._pending_queue.put(task.id)

    async def get_task(self, task_id: UUID) -> Optional[Task]:
        async with self._lock:
            return self._tasks.get(task_id)

    async def get_next_pending(self) -> Optional[UUID]:
        try:
            return await asyncio.wait_for(self._pending_queue.get(), timeout=0.1)
        except asyncio.TimeoutError:
            return None

    async def mark_pending(self, task_id: UUID) -> None:
        await self._pending_queue.put(task_id)

    async def update_task(self, task_id: UUID, **kwargs) -> None:
        async with self._lock:
            task = self._tasks.get(task_id)
            if task is None:
                return
            for key, value in kwargs.items():
                if hasattr(task, key):
                    setattr(task, key, value)

    async def register_idempotency_key(self, key: str, task_id: UUID, ttl_seconds: int = 300) -> bool:
        async with self._lock:
            now = datetime.utcnow()
            if key in self._idempotency_map:
                expiry = self._idempotency_expiry.get(key)
                if expiry is None or expiry > now:
                    return False
            self._idempotency_map[key] = task_id
            self._idempotency_expiry[key] = now + timedelta(seconds=ttl_seconds)
            return True

    async def get_task_by_idempotency_key(self, key: str) -> Optional[Task]:
        async with self._lock:
            now = datetime.utcnow()
            task_id = self._idempotency_map.get(key)
            if task_id is None:
                return None
            expiry = self._idempotency_expiry.get(key)
            if expiry is None or expiry <= now:
                return None
            return self._tasks.get(task_id)

    async def add_retrying(self, task_id: UUID) -> None:
        async with self._lock:
            self._retrying_task_ids.add(task_id)

    async def remove_retrying(self, task_id: UUID) -> None:
        async with self._lock:
            self._retrying_task_ids.discard(task_id)

    async def get_retrying_tasks(self) -> list[Task]:
        async with self._lock:
            return [
                self._tasks[tid].model_copy(deep=True)
                for tid in self._retrying_task_ids
                if tid in self._tasks and self._tasks[tid].status == TaskStatus.RETRYING
            ]


task_store = TaskStore()
