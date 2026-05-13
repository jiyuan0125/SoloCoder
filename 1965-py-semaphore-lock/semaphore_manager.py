import asyncio
import uuid
from dataclasses import dataclass, field
from datetime import datetime
from typing import Dict, List, Optional
from collections import deque


@dataclass
class AcquireInfo:
    acquire_id: str
    acquired_at: datetime
    event: asyncio.Event = field(default_factory=asyncio.Event)
    timeout: bool = False


@dataclass
class Holder:
    acquire_id: str
    acquired_at: datetime


class Semaphore:
    def __init__(self, name: str, capacity: int, default_wait_timeout: float = 30.0):
        self.name = name
        self.capacity = capacity
        self.default_wait_timeout = default_wait_timeout
        self.holders: Dict[str, Holder] = {}
        self.wait_queue: deque = deque()
        self._lock = asyncio.Lock()

    @property
    def current_holders_count(self) -> int:
        return len(self.holders)

    @property
    def wait_queue_length(self) -> int:
        return len(self.wait_queue)

    @property
    def available_slots(self) -> int:
        return max(0, self.capacity - self.current_holders_count)

    async def acquire(self, timeout: Optional[float] = None) -> Optional[str]:
        if timeout is None:
            timeout = self.default_wait_timeout
        
        acquire_id = str(uuid.uuid4())
        acquire_info = AcquireInfo(
            acquire_id=acquire_id,
            acquired_at=datetime.now()
        )

        async with self._lock:
            if self._can_acquire_now():
                self._grant_acquire(acquire_info)
                return acquire_id
            
            self.wait_queue.append(acquire_info)

        try:
            await asyncio.wait_for(acquire_info.event.wait(), timeout=timeout)
            return acquire_id
        except asyncio.TimeoutError:
            async with self._lock:
                acquire_info.timeout = True
                if acquire_info in self.wait_queue:
                    self.wait_queue.remove(acquire_info)
                return None

    async def release(self, acquire_id: str) -> bool:
        async with self._lock:
            if acquire_id not in self.holders:
                return False
            
            del self.holders[acquire_id]
            self._wake_up_waiters()
            return True

    async def update_capacity(self, new_capacity: int) -> None:
        async with self._lock:
            if new_capacity < 1:
                raise ValueError("Capacity must be at least 1")
            
            old_capacity = self.capacity
            self.capacity = new_capacity
            
            if new_capacity > old_capacity:
                self._wake_up_waiters()

    def get_status(self) -> Dict:
        return {
            "name": self.name,
            "capacity": self.capacity,
            "current_holders_count": self.current_holders_count,
            "wait_queue_length": self.wait_queue_length,
            "holders": [
                {
                    "acquire_id": h.acquire_id,
                    "acquired_at": h.acquired_at.isoformat()
                }
                for h in self.holders.values()
            ]
        }

    def _can_acquire_now(self) -> bool:
        return self.current_holders_count < self.capacity

    def _grant_acquire(self, acquire_info: AcquireInfo) -> None:
        self.holders[acquire_info.acquire_id] = Holder(
            acquire_id=acquire_info.acquire_id,
            acquired_at=datetime.now()
        )
        acquire_info.event.set()

    def _wake_up_waiters(self) -> None:
        while self._can_acquire_now() and self.wait_queue:
            acquire_info = self.wait_queue.popleft()
            if not acquire_info.timeout:
                self._grant_acquire(acquire_info)


class SemaphoreManager:
    def __init__(self, default_wait_timeout: float = 30.0):
        self.semaphores: Dict[str, Semaphore] = {}
        self.default_wait_timeout = default_wait_timeout
        self._lock = asyncio.Lock()

    async def create_semaphore(self, name: str, capacity: int) -> bool:
        if capacity < 1:
            raise ValueError("Capacity must be at least 1")
        
        async with self._lock:
            if name in self.semaphores:
                return False
            
            self.semaphores[name] = Semaphore(
                name=name,
                capacity=capacity,
                default_wait_timeout=self.default_wait_timeout
            )
            return True

    def get_semaphore(self, name: str) -> Optional[Semaphore]:
        return self.semaphores.get(name)

    async def delete_semaphore(self, name: str) -> bool:
        async with self._lock:
            if name not in self.semaphores:
                return False
            del self.semaphores[name]
            return True

    def list_all(self) -> List[str]:
        return list(self.semaphores.keys())
