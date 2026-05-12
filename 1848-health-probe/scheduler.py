from __future__ import annotations

import asyncio
from typing import Dict
from uuid import UUID

from checker import HealthManager, ProbeChecker
from models import HealthStatus
from notifier import Notifier
from storage import store


class HealthCheckScheduler:
    def __init__(self) -> None:
        self._tasks: Dict[UUID, asyncio.Task] = {}
        self._running = False
        self._group_status_cache: Dict[UUID, HealthStatus] = {}

    async def start(self) -> None:
        self._running = True
        existing_targets = await store.get_all_targets()
        for target in existing_targets:
            await self._schedule_target(target.id)

    async def stop(self) -> None:
        self._running = False
        for task in self._tasks.values():
            task.cancel()
        self._tasks.clear()

    async def schedule_new_target(self, target_id: UUID) -> None:
        if self._running:
            await self._schedule_target(target_id)

    async def _schedule_target(self, target_id: UUID) -> None:
        if target_id in self._tasks and not self._tasks[target_id].done():
            return
        task = asyncio.create_task(self._check_loop(target_id))
        self._tasks[target_id] = task

    async def _check_loop(self, target_id: UUID) -> None:
        while self._running:
            try:
                target = await store.get_target(target_id)
                if not target:
                    break

                success = await ProbeChecker.check_target(target)
                lock = await store.acquire_lock()
                async with lock:
                    target = await store.get_target(target_id)
                    if not target:
                        break
                    old_status = target.status
                    new_status = HealthManager.evaluate_status(target, success)
                    if new_status is not None:
                        target.status = new_status
                        HealthManager.log_state_change(
                            target.id, target.address, old_status, new_status
                        )
                    await store.update_target(target)

                if target.group_id:
                    await self._propagate_group_status(target.group_id)

                await asyncio.sleep(target.interval)

            except asyncio.CancelledError:
                break
            except Exception:
                await asyncio.sleep(5)

    async def _propagate_group_status(self, group_id: UUID) -> None:
        current_group_id = group_id
        while current_group_id:
            group = await store.get_group(current_group_id)
            if not group:
                break
            old_status = self._group_status_cache.get(current_group_id, HealthStatus.unknown)
            new_status = await HealthManager.calculate_group_status(HealthManager, current_group_id)
            if old_status != new_status:
                self._group_status_cache[current_group_id] = new_status
                await Notifier.notify_group_state_change(
                    current_group_id, group.name, old_status, new_status
                )
            current_group_id = group.parent_id


scheduler = HealthCheckScheduler()
