import asyncio
import time
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Callable, Dict, List, Optional, Set
from uuid import uuid4

import aiosqlite


class PoolFullStrategy(Enum):
    WAIT = "wait"
    REJECT = "reject"


class ConnectionState(Enum):
    IDLE = "idle"
    IN_USE = "in_use"
    INVALID = "invalid"


@dataclass
class PoolConfig:
    name: str
    db_url: str
    min_size: int = 1
    max_size: int = 10
    max_idle_time: float = 300.0
    pool_full_strategy: PoolFullStrategy = PoolFullStrategy.WAIT
    wait_timeout: float = 10.0
    health_check_interval: float = 60.0
    max_retries: int = 3
    retry_delay: float = 0.5


@dataclass
class PooledConnection:
    conn_id: str
    db_url: str
    state: ConnectionState = ConnectionState.IDLE
    created_at: float = field(default_factory=time.time)
    last_used_at: float = field(default_factory=time.time)
    last_checked_at: float = field(default_factory=time.time)
    use_count: int = 0
    _connection: Any = None
    _lock: asyncio.Lock = field(default_factory=asyncio.Lock)


@dataclass
class PoolStats:
    name: str
    total_connections: int
    idle_connections: int
    in_use_connections: int
    min_size: int
    max_size: int
    connections_created: int
    connections_destroyed: int
    borrows: int
    borrow_timeouts: int
    health_check_failures: int
    max_idle_time: float
    last_health_check_at: Optional[float]


class DatabasePool:
    def __init__(self, config: PoolConfig, create_conn: Callable, destroy_conn: Callable):
        self.config = config
        self._create_conn = create_conn
        self._destroy_conn = destroy_conn
        self._pool: List[PooledConnection] = []
        self._pool_lock = asyncio.Lock()
        self._condition = asyncio.Condition()
        self._is_running = False
        self._health_check_task: Optional[asyncio.Task] = None
        self._stats = {
            'connections_created': 0,
            'connections_destroyed': 0,
            'borrows': 0,
            'borrow_timeouts': 0,
            'health_check_failures': 0,
        }
        self._last_health_check_at: Optional[float] = None

    async def init(self):
        if self._is_running:
            return
        self._is_running = True
        for _ in range(self.config.min_size):
            conn = await self._create_new_connection()
            if conn:
                self._pool.append(conn)
        self._health_check_task = asyncio.create_task(self._health_check_loop())

    async def _create_new_connection(self) -> Optional[PooledConnection]:
        try:
            raw_conn = None
            for attempt in range(self.config.max_retries):
                try:
                    raw_conn = await self._create_conn(self.config.db_url)
                    break
                except Exception:
                    if attempt < self.config.max_retries - 1:
                        await asyncio.sleep(self.config.retry_delay)
                    else:
                        raise
            if raw_conn is None:
                return None
            pooled = PooledConnection(
                conn_id=str(uuid4()),
                db_url=self.config.db_url,
                _connection=raw_conn
            )
            self._stats['connections_created'] += 1
            return pooled
        except Exception:
            return None

    async def _destroy_connection(self, pooled: PooledConnection):
        try:
            if pooled._connection:
                await self._destroy_conn(pooled._connection)
        except Exception:
            pass
        self._stats['connections_destroyed'] += 1

    async def _check_connection_health(self, pooled: PooledConnection) -> bool:
        if pooled._connection is None:
            return False
        try:
            if isinstance(pooled._connection, aiosqlite.Connection):
                async with pooled._connection.execute("SELECT 1") as cursor:
                    await cursor.fetchone()
                return True
            return True
        except Exception:
            return False

    async def borrow(self) -> Optional[PooledConnection]:
        self._stats['borrows'] += 1
        if self.config.pool_full_strategy == PoolFullStrategy.WAIT:
            try:
                async with asyncio.timeout(self.config.wait_timeout):
                    return await self._borrow_connection()
            except asyncio.TimeoutError:
                self._stats['borrow_timeouts'] += 1
                raise PoolTimeoutError(f"Timeout waiting for connection from pool '{self.config.name}'")
        else:
            return await self._borrow_connection_non_blocking()

    async def _borrow_connection(self) -> PooledConnection:
        async with self._condition:
            while True:
                conn = await self._try_get_idle_connection()
                if conn:
                    return conn
                if len(self._pool) < self.config.max_size:
                    new_conn = await self._create_new_connection()
                    if new_conn:
                        self._pool.append(new_conn)
                        new_conn.state = ConnectionState.IN_USE
                        new_conn.last_used_at = time.time()
                        new_conn.use_count += 1
                        return new_conn
                await self._condition.wait()

    async def _try_get_idle_connection(self) -> Optional[PooledConnection]:
        for conn in self._pool:
            if conn.state == ConnectionState.IDLE:
                healthy = await self._check_connection_health(conn)
                if not healthy:
                    await self._destroy_connection(conn)
                    self._pool.remove(conn)
                    self._stats['health_check_failures'] += 1
                    continue
                conn.state = ConnectionState.IN_USE
                conn.last_used_at = time.time()
                conn.use_count += 1
                return conn
        return None

    async def _borrow_connection_non_blocking(self) -> Optional[PooledConnection]:
        async with self._pool_lock:
            conn = await self._try_get_idle_connection()
            if conn:
                return conn
            if len(self._pool) < self.config.max_size:
                new_conn = await self._create_new_connection()
                if new_conn:
                    self._pool.append(new_conn)
                    new_conn.state = ConnectionState.IN_USE
                    new_conn.last_used_at = time.time()
                    new_conn.use_count += 1
                    return new_conn
            raise PoolFullError(f"Pool '{self.config.name}' is full")

    async def release(self, pooled: PooledConnection):
        async with self._condition:
            if pooled in self._pool and pooled.state == ConnectionState.IN_USE:
                pooled.state = ConnectionState.IDLE
                pooled.last_used_at = time.time()
                self._condition.notify_all()

    async def _health_check_loop(self):
        while self._is_running:
            await asyncio.sleep(self.config.health_check_interval)
            if not self._is_running:
                break
            await self._perform_health_check()

    async def _perform_health_check(self):
        self._last_health_check_at = time.time()
        to_remove: Set[PooledConnection] = set()
        now = time.time()
        async with self._pool_lock:
            for conn in list(self._pool):
                if conn.state == ConnectionState.IDLE:
                    if now - conn.last_used_at > self.config.max_idle_time:
                        if len(self._pool) - len(to_remove) > self.config.min_size:
                            to_remove.add(conn)
                            continue
                    healthy = await self._check_connection_health(conn)
                    if not healthy:
                        self._stats['health_check_failures'] += 1
                        to_remove.add(conn)
            for conn in to_remove:
                await self._destroy_connection(conn)
                if conn in self._pool:
                    self._pool.remove(conn)
        async with self._condition:
            self._condition.notify_all()

    def get_stats(self) -> PoolStats:
        total = len(self._pool)
        idle = sum(1 for c in self._pool if c.state == ConnectionState.IDLE)
        in_use = sum(1 for c in self._pool if c.state == ConnectionState.IN_USE)
        return PoolStats(
            name=self.config.name,
            total_connections=total,
            idle_connections=idle,
            in_use_connections=in_use,
            min_size=self.config.min_size,
            max_size=self.config.max_size,
            connections_created=self._stats['connections_created'],
            connections_destroyed=self._stats['connections_destroyed'],
            borrows=self._stats['borrows'],
            borrow_timeouts=self._stats['borrow_timeouts'],
            health_check_failures=self._stats['health_check_failures'],
            max_idle_time=self.config.max_idle_time,
            last_health_check_at=self._last_health_check_at
        )

    async def close(self):
        self._is_running = False
        if self._health_check_task:
            self._health_check_task.cancel()
            try:
                await self._health_check_task
            except (asyncio.CancelledError, Exception):
                pass
        async with self._pool_lock:
            for conn in list(self._pool):
                await self._destroy_connection(conn)
            self._pool.clear()


class PoolTimeoutError(Exception):
    pass


class PoolFullError(Exception):
    pass


class PoolManager:
    def __init__(self):
        self._pools: Dict[str, DatabasePool] = {}
        self._manager_lock = asyncio.Lock()

    async def _default_create_conn(self, db_url: str):
        conn = await aiosqlite.connect(db_url)
        await conn.execute("PRAGMA journal_mode=WAL")
        await conn.commit()
        return conn

    async def _default_destroy_conn(self, conn):
        await conn.close()

    async def create_pool(self, config: PoolConfig, 
                         create_conn: Callable = None,
                         destroy_conn: Callable = None) -> DatabasePool:
        async with self._manager_lock:
            if config.name in self._pools:
                raise ValueError(f"Pool '{config.name}' already exists")
            pool = DatabasePool(
                config=config,
                create_conn=create_conn or self._default_create_conn,
                destroy_conn=destroy_conn or self._default_destroy_conn
            )
            await pool.init()
            self._pools[config.name] = pool
            return pool

    def get_pool(self, name: str) -> Optional[DatabasePool]:
        return self._pools.get(name)

    def list_pools(self) -> List[str]:
        return list(self._pools.keys())

    def get_all_stats(self) -> Dict[str, PoolStats]:
        return {name: pool.get_stats() for name, pool in self._pools.items()}

    async def remove_pool(self, name: str):
        async with self._manager_lock:
            pool = self._pools.pop(name, None)
            if pool:
                await pool.close()

    async def close_all(self):
        async with self._manager_lock:
            for pool in self._pools.values():
                await pool.close()
            self._pools.clear()
