from __future__ import annotations

import asyncio
from typing import Optional
from uuid import UUID

from models import Callback, Group, StateChangeLog, Target


class InMemoryStore:
    def __init__(self) -> None:
        self._targets: dict[UUID, Target] = {}
        self._groups: dict[UUID, Group] = {}
        self._callbacks: dict[UUID, Callback] = {}
        self._logs: list[StateChangeLog] = []
        self._lock = asyncio.Lock()

    async def add_target(self, target: Target) -> None:
        async with self._lock:
            self._targets[target.id] = target

    async def get_target(self, target_id: UUID) -> Optional[Target]:
        async with self._lock:
            return self._targets.get(target_id)

    async def get_all_targets(self) -> list[Target]:
        async with self._lock:
            return list(self._targets.values())

    async def update_target(self, target: Target) -> None:
        async with self._lock:
            self._targets[target.id] = target

    async def add_group(self, group: Group) -> None:
        async with self._lock:
            self._groups[group.id] = group

    async def get_group(self, group_id: UUID) -> Optional[Group]:
        async with self._lock:
            return self._groups.get(group_id)

    async def get_all_groups(self) -> list[Group]:
        async with self._lock:
            return list(self._groups.values())

    async def get_children_groups(self, parent_id: UUID) -> list[Group]:
        async with self._lock:
            return [g for g in self._groups.values() if g.parent_id == parent_id]

    async def get_targets_in_group(self, group_id: UUID) -> list[Target]:
        async with self._lock:
            return [t for t in self._targets.values() if t.group_id == group_id]

    async def add_callback(self, callback: Callback) -> None:
        async with self._lock:
            self._callbacks[callback.id] = callback

    async def get_all_callbacks(self) -> list[Callback]:
        async with self._lock:
            return list(self._callbacks.values())

    async def add_state_log(self, log: StateChangeLog) -> None:
        async with self._lock:
            self._logs.append(log)

    async def acquire_lock(self) -> asyncio.Lock:
        return self._lock


store = InMemoryStore()
