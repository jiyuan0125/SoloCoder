import asyncio
import random
import time
from dataclasses import dataclass, field
from threading import RLock
from typing import Dict, List, Optional

import httpx


@dataclass
class BackendInstance:
    address: str
    group: str
    registered_at: float
    healthy: bool = True
    grace_period_remaining: float = 30.0


@dataclass
class GroupStats:
    total_requests: int = 0


class BackendRegistry:
    def __init__(self):
        self._backends: Dict[str, List[BackendInstance]] = {}
        self._stats: Dict[str, GroupStats] = {}
        self._lock = RLock()
        self._health_check_interval = 10.0
        self._grace_period = 30.0
        self._running = False
        self._check_task: Optional[asyncio.Task] = None

    def register_backend(self, group: str, address: str):
        with self._lock:
            if group not in self._backends:
                self._backends[group] = []
                self._stats[group] = GroupStats()
            existing = next((b for b in self._backends[group] if b.address == address), None)
            if existing is not None:
                existing.registered_at = time.time()
                existing.grace_period_remaining = self._grace_period
                existing.healthy = True
                return
            instance = BackendInstance(
                address=address,
                group=group,
                registered_at=time.time(),
                grace_period_remaining=self._grace_period,
                healthy=True
            )
            self._backends[group].append(instance)

    def unregister_backend(self, group: str, address: str):
        with self._lock:
            if group not in self._backends:
                return
            self._backends[group] = [b for b in self._backends[group] if b.address != address]

    def get_healthy_backends(self, group: str) -> List[BackendInstance]:
        with self._lock:
            if group not in self._backends:
                return []
            return [b for b in self._backends[group] if b.healthy]

    def pick_backend(self, group: str) -> Optional[BackendInstance]:
        healthy = self.get_healthy_backends(group)
        if not healthy:
            return None
        return random.choice(healthy)

    def record_request(self, group: str):
        with self._lock:
            if group in self._stats:
                self._stats[group].total_requests += 1

    def get_all_backends(self) -> Dict[str, List[BackendInstance]]:
        with self._lock:
            result = {}
            for group, backends in self._backends.items():
                result[group] = [
                    BackendInstance(
                        address=b.address,
                        group=b.group,
                        registered_at=b.registered_at,
                        healthy=b.healthy,
                        grace_period_remaining=b.grace_period_remaining
                    )
                    for b in backends
                ]
            return result

    def get_group_backends(self, group: str) -> List[BackendInstance]:
        with self._lock:
            if group not in self._backends:
                return []
            return [
                BackendInstance(
                    address=b.address,
                    group=b.group,
                    registered_at=b.registered_at,
                    healthy=b.healthy,
                    grace_period_remaining=b.grace_period_remaining
                )
                for b in self._backends[group]
            ]

    def get_request_counts(self) -> Dict[str, int]:
        with self._lock:
            return {group: stats.total_requests for group, stats in self._stats.items()}

    def start_health_checker(self):
        if self._running:
            return
        self._running = True
        loop = asyncio.get_event_loop()
        self._check_task = loop.create_task(self._health_check_loop())

    def stop_health_checker(self):
        self._running = False
        if self._check_task is not None:
            self._check_task.cancel()
            self._check_task = None

    async def _health_check_loop(self):
        while self._running:
            await asyncio.sleep(self._health_check_interval)
            await self._run_health_checks()

    async def _run_health_checks(self):
        with self._lock:
            backends_to_check = []
            for group, backends in self._backends.items():
                for backend in backends:
                    if backend.grace_period_remaining > 0:
                        backend.grace_period_remaining = max(0, backend.grace_period_remaining - self._health_check_interval)
                    backends_to_check.append(backend)

        async with httpx.AsyncClient(timeout=5.0) as client:
            tasks = []
            for backend in backends_to_check:
                if backend.grace_period_remaining <= 0:
                    tasks.append(self._check_backend(client, backend))
            if tasks:
                await asyncio.gather(*tasks, return_exceptions=True)

    async def _check_backend(self, client: httpx.AsyncClient, backend: BackendInstance):
        try:
            resp = await client.get(f"{backend.address}/health")
            if resp.status_code < 500:
                with self._lock:
                    backend.healthy = True
            else:
                with self._lock:
                    backend.healthy = False
        except Exception:
            with self._lock:
                backend.healthy = False
