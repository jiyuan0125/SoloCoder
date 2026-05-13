import asyncio
import logging
import time
from typing import Optional, Set, Deque, Dict, Any
from collections import deque

from .config import PoolConfig, DatasourceConfig, ProtocolType
from .connection import PooledConnection, create_connection, ConnectionError

logger = logging.getLogger(__name__)


class PoolAcquireTimeoutError(Exception):
    pass


class PoolClosedError(Exception):
    pass


class PoolDrainingError(Exception):
    pass


class PoolStats:
    def __init__(self) -> None:
        self.total_created: int = 0
        self.total_destroyed: int = 0
        self.leak_count: int = 0

    def to_dict(self) -> Dict[str, Any]:
        return {
            "total_created": self.total_created,
            "total_destroyed": self.total_destroyed,
            "leak_count": self.leak_count,
        }


class ConnectionPool:
    def __init__(self, datasource_config: DatasourceConfig):
        datasource_config.validate()
        self._datasource_config: DatasourceConfig = datasource_config
        self._pool_config: PoolConfig = datasource_config.pool_config

        self._idle: Deque[PooledConnection] = deque()
        self._in_use: Set[PooledConnection] = set()
        self._waiters: Deque[asyncio.Future] = deque()
        self._closed: bool = False
        self._draining: bool = False
        self._destroy_task: Optional[asyncio.Task] = None

        self._stats: PoolStats = PoolStats()
        self._health_task: Optional[asyncio.Task] = None
        self._leak_task: Optional[asyncio.Task] = None
        self._lock: asyncio.Lock = asyncio.Lock()
        self._refill_semaphore: asyncio.Semaphore = asyncio.Semaphore(self._pool_config.refill_rate_limit)
        self._last_refill_window: float = 0.0
        self._refill_window_count: int = 0

        self._all_connections_closed_event: Optional[asyncio.Event] = None

    @property
    def name(self) -> str:
        return self._datasource_config.name

    @property
    def config(self) -> DatasourceConfig:
        return self._datasource_config

    @property
    def pool_config(self) -> PoolConfig:
        return self._pool_config

    @property
    def closed(self) -> bool:
        return self._closed

    @property
    def draining(self) -> bool:
        return self._draining

    def stats(self) -> Dict[str, Any]:
        total = len(self._idle) + len(self._in_use)
        return {
            "name": self._datasource_config.name,
            "protocol": self._datasource_config.protocol.value,
            "host": self._datasource_config.host,
            "port": self._datasource_config.port,
            "active": len(self._in_use),
            "idle": len(self._idle),
            "total": total,
            "max_size": self._pool_config.max_size,
            "min_idle": self._pool_config.min_idle,
            "waiting": len(self._waiters),
            "closed": self._closed,
            "draining": self._draining,
            **self._stats.to_dict(),
        }

    async def update_pool_config(self, new_config: PoolConfig) -> None:
        new_config.validate()
        if new_config.min_idle > new_config.max_size:
            raise ValueError("min_idle must not exceed max_size")

        async with self._lock:
            self._pool_config = new_config
            self._refill_semaphore = asyncio.Semaphore(new_config.refill_rate_limit)
            logger.info("Pool %s config updated: max_size=%d, min_idle=%d, acquire_timeout=%.2f",
                        self.name, new_config.max_size, new_config.min_idle, new_config.acquire_timeout)

    async def start(self) -> None:
        async with self._lock:
            if self._closed:
                raise PoolClosedError(f"Pool {self.name} is already closed")
            if self._health_task is not None:
                return

            await self._ensure_min_idle()
            self._health_task = asyncio.create_task(self._health_check_loop())
            self._leak_task = asyncio.create_task(self._leak_detection_loop())
            logger.info("Pool %s started", self.name)

    async def _ensure_min_idle(self) -> None:
        need = max(0, self._pool_config.min_idle - len(self._idle))
        for _ in range(need):
            conn = await self._create_new_connection()
            if conn:
                self._idle.append(conn)

    async def _create_new_connection(self) -> Optional[PooledConnection]:
        await self._throttle_refill()
        conn = create_connection(
            self._datasource_config.name,
            self._datasource_config.protocol,
            self._datasource_config.host,
            self._datasource_config.port,
        )
        try:
            await conn.open()
            self._stats.total_created += 1
            return conn
        except ConnectionError as e:
            logger.warning("Failed to create connection for pool %s: %s", self.name, e)
            return None
        except Exception as e:
            logger.exception("Unexpected error creating connection for pool %s: %s", self.name, e)
            return None

    async def _throttle_refill(self) -> None:
        now = time.monotonic()
        if now - self._last_refill_window >= 1.0:
            self._last_refill_window = now
            self._refill_window_count = 0
        while self._refill_window_count >= self._pool_config.refill_rate_limit:
            await asyncio.sleep(0.1)
            now = time.monotonic()
            if now - self._last_refill_window >= 1.0:
                self._last_refill_window = now
                self._refill_window_count = 0
        self._refill_window_count += 1

    async def acquire(self) -> PooledConnection:
        if self._closed:
            raise PoolClosedError(f"Pool {self.name} is closed")
        if self._draining:
            raise PoolDrainingError(f"Pool {self.name} is draining")

        try:
            return await asyncio.wait_for(self._do_acquire(), timeout=self._pool_config.acquire_timeout)
        except asyncio.TimeoutError:
            raise PoolAcquireTimeoutError(f"Timeout acquiring connection from pool {self.name}")

    async def _do_acquire(self) -> PooledConnection:
        while True:
            async with self._lock:
                if self._closed:
                    raise PoolClosedError(f"Pool {self.name} is closed")
                if self._draining:
                    raise PoolDrainingError(f"Pool {self.name} is draining")

                while self._idle:
                    conn = self._idle.popleft()
                    if conn.is_healthy:
                        conn.mark_borrowed()
                        self._in_use.add(conn)
                        return conn
                    else:
                        await self._destroy_connection(conn)

                total = len(self._idle) + len(self._in_use)
                if total < self._pool_config.max_size:
                    conn = await self._create_new_connection()
                    if conn and conn.is_healthy:
                        conn.mark_borrowed()
                        self._in_use.add(conn)
                        return conn
                    if conn:
                        await self._destroy_connection(conn)

                waiter = asyncio.get_running_loop().create_future()
                self._waiters.append(waiter)

            try:
                conn = await waiter
                if conn is not None and conn.is_healthy:
                    return conn
            except asyncio.CancelledError:
                async with self._lock:
                    try:
                        self._waiters.remove(waiter)
                    except ValueError:
                        pass
                raise

    async def release(self, conn: PooledConnection) -> None:
        if conn.datasource_name != self.name:
            logger.warning("Connection %s does not belong to pool %s", conn.id, self.name)
            return

        async with self._lock:
            try:
                self._in_use.remove(conn)
            except KeyError:
                pass
            conn.mark_returned()

            if not conn.is_healthy:
                await self._destroy_connection(conn)
                return

            if not conn.is_open:
                await self._destroy_connection(conn)
                return

            if self._closed or self._draining:
                await self._destroy_connection(conn)
                return

            total = len(self._idle) + len(self._in_use)
            if total >= self._pool_config.max_size:
                await self._destroy_connection(conn)
                return

            self._idle.append(conn)

            while self._waiters:
                waiter = self._waiters.popleft()
                if not waiter.done():
                    while self._idle:
                        idle_conn = self._idle.popleft()
                        if idle_conn.is_healthy:
                            idle_conn.mark_borrowed()
                            self._in_use.add(idle_conn)
                            waiter.set_result(idle_conn)
                            return
                        else:
                            await self._destroy_connection(idle_conn)
                    waiter.set_result(None)
                    break

    async def _destroy_connection(self, conn: PooledConnection) -> None:
        await conn.close()
        self._stats.total_destroyed += 1

    async def _health_check_loop(self) -> None:
        while not self._closed:
            try:
                await asyncio.sleep(self._pool_config.health_check_interval)
                await self._run_health_check()
            except asyncio.CancelledError:
                break
            except Exception:
                logger.exception("Error in health check loop for pool %s", self.name)

    async def _run_health_check(self) -> None:
        async with self._lock:
            to_check = list(self._idle)

        for conn in to_check:
            try:
                is_healthy = await conn.health_check()
                if not is_healthy:
                    async with self._lock:
                        if conn in self._idle:
                            self._idle.remove(conn)
                            await self._destroy_connection(conn)
            except Exception:
                logger.exception("Error checking health of connection %s", conn.id)

        async with self._lock:
            now = time.monotonic()
            to_remove = []
            for conn in list(self._idle):
                if conn.idle_duration > self._pool_config.idle_timeout:
                    to_remove.append(conn)
            for conn in to_remove:
                if conn in self._idle:
                    self._idle.remove(conn)
                    await self._destroy_connection(conn)

            total = len(self._idle) + len(self._in_use)
            if len(self._idle) < self._pool_config.min_idle and total < self._pool_config.max_size:
                need = max(0, self._pool_config.min_idle - len(self._idle))
                need = min(need, self._pool_config.max_size - total)
                for _ in range(need):
                    conn = await self._create_new_connection()
                    if conn:
                        self._idle.append(conn)

    async def _leak_detection_loop(self) -> None:
        while not self._closed:
            try:
                await asyncio.sleep(60)
                await self._run_leak_detection()
            except asyncio.CancelledError:
                break
            except Exception:
                logger.exception("Error in leak detection loop for pool %s", self.name)

    async def _run_leak_detection(self) -> None:
        async with self._lock:
            to_close = []
            for conn in list(self._in_use):
                if conn.borrowed_duration > self._pool_config.leak_detection_threshold:
                    logger.warning(
                        "Leak detected: connection %s borrowed for %.2f seconds, threshold %.2f",
                        conn.id, conn.borrowed_duration, self._pool_config.leak_detection_threshold,
                    )
                    self._stats.leak_count += 1
                    to_close.append(conn)

            for conn in to_close:
                if conn in self._in_use:
                    self._in_use.remove(conn)
                    await self._destroy_connection(conn)

    async def drain(self) -> None:
        async with self._lock:
            if self._draining or self._closed:
                return
            self._draining = True
            self._all_connections_closed_event = asyncio.Event()

            for waiter in list(self._waiters):
                if not waiter.done():
                    waiter.set_exception(PoolDrainingError(f"Pool {self.name} is draining"))
            self._waiters.clear()

            logger.info("Pool %s started draining", self.name)

        while True:
            async with self._lock:
                if not self._in_use:
                    for conn in list(self._idle):
                        await self._destroy_connection(conn)
                    self._idle.clear()
                    if self._all_connections_closed_event:
                        self._all_connections_closed_event.set()
                    logger.info("Pool %s drained successfully", self.name)
                    return
            await asyncio.sleep(0.5)

    async def close(self) -> None:
        async with self._lock:
            if self._closed:
                return
            self._closed = True

            if self._health_task:
                self._health_task.cancel()
                self._health_task = None
            if self._leak_task:
                self._leak_task.cancel()
                self._leak_task = None

            for waiter in list(self._waiters):
                if not waiter.done():
                    waiter.set_exception(PoolClosedError(f"Pool {self.name} is closed"))
            self._waiters.clear()

            for conn in list(self._idle):
                await self._destroy_connection(conn)
            self._idle.clear()

            for conn in list(self._in_use):
                await self._destroy_connection(conn)
            self._in_use.clear()

            logger.info("Pool %s closed", self.name)
