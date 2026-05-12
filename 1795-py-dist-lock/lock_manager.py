import asyncio
import uuid
import time
from dataclasses import dataclass, field
from typing import Dict, List, Optional
from collections import deque


@dataclass
class LockRecord:
    name: str
    lock_id: str
    acquired_at: float
    expires_at: float
    ttl: int
    original_ttl: int
    max_extend_ttl: int
    accumulated_extension: int = 0
    waiters: deque = field(default_factory=deque)


@dataclass
class Waiter:
    lock_id: str
    future: asyncio.Future
    ttl: int


class LockManager:
    def __init__(self):
        self._locks: Dict[str, LockRecord] = {}
        self._lock = asyncio.Lock()

    async def acquire(self, name: str, ttl: int = 30) -> Waiter:
        waiter = Waiter(
            lock_id=str(uuid.uuid4()),
            future=asyncio.get_event_loop().create_future(),
            ttl=ttl
        )
        
        async with self._lock:
            if name not in self._locks or self._locks[name].expires_at <= time.time():
                if name in self._locks and self._locks[name].expires_at <= time.time():
                    self._cleanup_expired(name)
                
                now = time.time()
                record = LockRecord(
                    name=name,
                    lock_id=waiter.lock_id,
                    acquired_at=now,
                    expires_at=now + ttl,
                    ttl=ttl,
                    original_ttl=ttl,
                    max_extend_ttl=ttl * 3
                )
                self._locks[name] = record
                waiter.future.set_result({
                    "lock_id": waiter.lock_id,
                    "expires_at": record.expires_at
                })
            else:
                self._locks[name].waiters.append(waiter)
        
        return waiter

    async def release(self, name: str, lock_id: str) -> int:
        async with self._lock:
            if name not in self._locks:
                return 404
            
            record = self._locks[name]
            
            if record.expires_at <= time.time():
                self._cleanup_expired(name)
                return 404
            
            if record.lock_id != lock_id:
                return 403
            
            self._release_and_notify(name)
            return 200

    async def renew(self, name: str, lock_id: str, extend_ttl: int = 30) -> Optional[float]:
        async with self._lock:
            if name not in self._locks:
                return None
            
            record = self._locks[name]
            
            if record.lock_id != lock_id:
                return None
            
            if record.expires_at <= time.time():
                self._cleanup_expired(name)
                return None
            
            new_accumulated = record.accumulated_extension + extend_ttl
            if new_accumulated > record.max_extend_ttl:
                return None
            
            now = time.time()
            record.expires_at = now + extend_ttl
            record.accumulated_extension = new_accumulated
            
            return record.expires_at

    async def force_release(self, name: str) -> bool:
        async with self._lock:
            if name not in self._locks:
                return False
            
            self._release_and_notify(name)
            return True

    def get_all_locks(self) -> List[dict]:
        result = []
        for name, record in self._locks.items():
            result.append({
                "name": name,
                "lock_id": record.lock_id,
                "acquired_at": record.acquired_at,
                "expires_at": record.expires_at,
                "waiter_count": len(record.waiters),
                "is_expired": record.expires_at <= time.time()
            })
        return result

    async def cleanup_expired(self):
        async with self._lock:
            for name in list(self._locks.keys()):
                record = self._locks[name]
                if record.expires_at <= time.time():
                    self._release_and_notify(name)

    def _release_and_notify(self, name: str):
        record = self._locks[name]
        
        waiters = record.waiters
        if waiters:
            waiter = waiters.popleft()
            now = time.time()
            record.lock_id = waiter.lock_id
            record.acquired_at = now
            record.expires_at = now + waiter.ttl
            record.ttl = waiter.ttl
            record.original_ttl = waiter.ttl
            record.max_extend_ttl = waiter.ttl * 3
            record.accumulated_extension = 0
            waiter.future.set_result({
                "lock_id": waiter.lock_id,
                "expires_at": record.expires_at
            })
        else:
            del self._locks[name]

    def _cleanup_expired(self, name: str):
        if name in self._locks:
            del self._locks[name]


lock_manager = LockManager()