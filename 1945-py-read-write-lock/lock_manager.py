import asyncio
from enum import Enum
from dataclasses import dataclass, field
from typing import Dict, List, Optional


class LockType(str, Enum):
    READ = "read"
    WRITE = "write"


@dataclass
class Waiter:
    client_id: str
    lock_type: LockType
    event: asyncio.Event
    granted: bool = False


@dataclass
class ResourceLock:
    holders: Dict[str, LockType] = field(default_factory=dict)
    waiters: List[Waiter] = field(default_factory=list)
    downgrading_client: Optional[str] = None


class LockManager:
    def __init__(self):
        self._locks: Dict[str, ResourceLock] = {}
        self._global_lock = asyncio.Lock()

    def _get_resource_lock(self, resource_name: str) -> ResourceLock:
        if resource_name not in self._locks:
            self._locks[resource_name] = ResourceLock()
        return self._locks[resource_name]

    def _has_write_holder(self, rlock: ResourceLock) -> bool:
        return LockType.WRITE in rlock.holders.values()

    def _has_any_holder(self, rlock: ResourceLock) -> bool:
        return len(rlock.holders) > 0

    def _has_pending_writer(self, rlock: ResourceLock) -> bool:
        return any(w.lock_type == LockType.WRITE for w in rlock.waiters)

    def _can_grant_read(self, rlock: ResourceLock, client_id: str) -> bool:
        if client_id in rlock.holders:
            return rlock.holders[client_id] == LockType.READ
        if self._has_write_holder(rlock):
            return False
        if self._has_pending_writer(rlock):
            return False
        return True

    def _can_grant_write(self, rlock: ResourceLock, client_id: str) -> bool:
        if client_id in rlock.holders:
            return rlock.holders[client_id] == LockType.WRITE
        return not self._has_any_holder(rlock) and len(rlock.waiters) == 0

    def _wake_next(self, rlock: ResourceLock) -> None:
        while rlock.waiters:
            next_waiter = rlock.waiters[0]
            if next_waiter.lock_type == LockType.WRITE:
                if not self._has_any_holder(rlock):
                    waiter = rlock.waiters.pop(0)
                    rlock.holders[waiter.client_id] = LockType.WRITE
                    waiter.granted = True
                    waiter.event.set()
                    return
                else:
                    return
            else:
                if self._can_grant_read(rlock, next_waiter.client_id):
                    idx = 0
                    while idx < len(rlock.waiters):
                        waiter = rlock.waiters[idx]
                        if waiter.lock_type == LockType.READ:
                            rlock.holders[waiter.client_id] = LockType.READ
                            waiter.granted = True
                            waiter.event.set()
                            rlock.waiters.pop(idx)
                        else:
                            idx += 1
                    return
                else:
                    return

    async def acquire(
        self,
        resource_name: str,
        lock_type: LockType,
        client_id: str,
        timeout: float = 10.0,
    ) -> bool:
        if timeout <= 0:
            timeout = 10.0

        async with self._global_lock:
            rlock = self._get_resource_lock(resource_name)

            if lock_type == LockType.READ:
                if self._can_grant_read(rlock, client_id):
                    rlock.holders[client_id] = LockType.READ
                    return True
            else:
                if self._can_grant_write(rlock, client_id):
                    rlock.holders[client_id] = LockType.WRITE
                    return True

            waiter = Waiter(
                client_id=client_id, lock_type=lock_type, event=asyncio.Event()
            )
            rlock.waiters.append(waiter)

        try:
            await asyncio.wait_for(waiter.event.wait(), timeout=timeout)
            return waiter.granted
        except asyncio.TimeoutError:
            async with self._global_lock:
                if waiter in rlock.waiters:
                    rlock.waiters.remove(waiter)
                if waiter.granted:
                    await self.release(resource_name, client_id)
            return False

    async def release(self, resource_name: str, client_id: str) -> bool:
        async with self._global_lock:
            rlock = self._get_resource_lock(resource_name)

            if client_id not in rlock.holders:
                return False

            del rlock.holders[client_id]
            self._wake_next(rlock)
            return True

    async def downgrade(self, resource_name: str, client_id: str) -> bool:
        async with self._global_lock:
            rlock = self._get_resource_lock(resource_name)

            if client_id not in rlock.holders:
                return False

            if rlock.holders[client_id] != LockType.WRITE:
                return False

            rlock.holders[client_id] = LockType.READ

            self._wake_next(rlock)
            return True

    def get_status(self, resource_name: str) -> dict:
        if resource_name not in self._locks:
            return {
                "resource": resource_name,
                "holders": [],
                "waiters": [],
            }

        rlock = self._locks[resource_name]
        return {
            "resource": resource_name,
            "holders": [
                {"client_id": cid, "lock_type": lt.value}
                for cid, lt in rlock.holders.items()
            ],
            "waiters": [
                {"client_id": w.client_id, "lock_type": w.lock_type.value}
                for w in rlock.waiters
            ],
        }


lock_manager = LockManager()
