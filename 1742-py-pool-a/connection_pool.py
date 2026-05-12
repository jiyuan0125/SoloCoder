import asyncio
import time
from typing import AsyncGenerator, Optional, Set
from collections import deque
from contextlib import asynccontextmanager

from connection import Connection, MockConnection
from pool_config import PoolConfig, ConnectionInfo, PoolStats


class ConnectionPool:
    def __init__(self, config: PoolConfig, name: str = "default"):
        self.name = name
        self.config = config
        self._idle_connections: deque[ConnectionInfo] = deque()
        self._active_connections: Set[Connection] = set()
        self._lock = asyncio.Lock()
        self._semaphore = asyncio.Semaphore(config.max_connections)
        self._stats = PoolStats()
        self._closed = False
        self._recycle_task: Optional[asyncio.Task] = None
        self._not_empty = asyncio.Condition(self._lock)

    async def initialize(self) -> None:
        async with self._lock:
            for _ in range(self.config.min_connections):
                conn_info = await self._create_connection()
                self._idle_connections.append(conn_info)
        self._recycle_task = asyncio.create_task(self._idle_recycle_loop())

    async def close(self) -> None:
        self._closed = True
        if self._recycle_task:
            self._recycle_task.cancel()
            try:
                await self._recycle_task
            except asyncio.CancelledError:
                pass
        
        async with self._lock:
            while self._idle_connections:
                conn_info = self._idle_connections.popleft()
                await self._destroy_connection(conn_info)
            
            for conn in list(self._active_connections):
                await conn.close()
                self._active_connections.clear()

    async def _create_connection(self) -> ConnectionInfo:
        if self.config.create_connection:
            conn = await self.config.create_connection()
        else:
            conn = MockConnection()
            if hasattr(conn, 'connect'):
                await conn.connect()
        
        self._stats.total_created += 1
        return ConnectionInfo(
            connection=conn,
            last_used_time=time.time(),
            usage_count=0
        )

    async def _destroy_connection(self, conn_info: ConnectionInfo) -> None:
        await conn_info.connection.close()
        self._stats.total_destroyed += 1

    async def acquire(self) -> Connection:
        if self._closed:
            raise RuntimeError("Connection pool is closed")
        
        start_time = time.time()
        
        acquired = False
        try:
            acquired = await asyncio.wait_for(
                self._semaphore.acquire(),
                timeout=self.config.wait_timeout
            )
        except asyncio.TimeoutError:
            pass
        
        if not acquired:
            async with self._lock:
                self._stats.waiting_requests -= 1 if self._stats.waiting_requests > 0 else 0
            raise asyncio.TimeoutError(
                f"Timeout waiting for connection in pool '{self.name}' after {self.config.wait_timeout}s"
            )
        
        try:
            async with self._lock:
                self._stats.waiting_requests += 1
                
                while not self._idle_connections and not self._closed:
                    try:
                        await asyncio.wait_for(
                            self._not_empty.wait(),
                            timeout=max(0.1, self.config.wait_timeout - (time.time() - start_time))
                        )
                    except asyncio.TimeoutError:
                        pass
                
                if self._closed:
                    raise RuntimeError("Connection pool is closed")
                
                if self._idle_connections:
                    conn_info = self._idle_connections.popleft()
                else:
                    conn_info = await self._create_connection()
                
                self._stats.waiting_requests -= 1
                
                conn_info.usage_count += 1
                conn_info.last_used_time = time.time()
                self._active_connections.add(conn_info.connection)
                self._stats.active_connections += 1
                self._stats.idle_connections = len(self._idle_connections)
                
                return conn_info.connection
                
        except Exception:
            self._semaphore.release()
            raise

    async def release(self, conn: Connection) -> None:
        if self._closed:
            await conn.close()
            return
        
        async with self._lock:
            if conn not in self._active_connections:
                return
            
            self._active_connections.remove(conn)
            self._stats.active_connections -= 1
            
            is_healthy = await conn.is_healthy()
            
            if is_healthy:
                conn_info = ConnectionInfo(
                    connection=conn,
                    last_used_time=time.time(),
                    usage_count=1
                )
                self._idle_connections.append(conn_info)
                self._stats.idle_connections = len(self._idle_connections)
                self._not_empty.notify()
            else:
                await self._destroy_connection(ConnectionInfo(connection=conn, last_used_time=time.time()))
                new_conn_info = await self._create_connection()
                self._idle_connections.append(new_conn_info)
                self._stats.idle_connections = len(self._idle_connections)
                self._not_empty.notify()
            
            self._semaphore.release()

    async def _idle_recycle_loop(self) -> None:
        while not self._closed:
            try:
                await asyncio.sleep(self.config.idle_recycle_interval)
                await self._recycle_idle_connections()
            except asyncio.CancelledError:
                break
            except Exception:
                continue

    async def _recycle_idle_connections(self) -> None:
        async with self._lock:
            current_time = time.time()
            to_recycle: list[ConnectionInfo] = []
            remaining: list[ConnectionInfo] = []
            
            all_connections = list(self._idle_connections)
            all_connections.sort(key=lambda c: c.usage_count)
            
            for conn_info in all_connections:
                idle_time = current_time - conn_info.last_used_time
                if idle_time >= self.config.idle_timeout:
                    if conn_info.usage_count <= self.config.min_usage_count:
                        to_recycle.append(conn_info)
                        continue
                remaining.append(conn_info)
            
            for conn_info in to_recycle:
                await self._destroy_connection(conn_info)
            
            self._idle_connections = deque(remaining)
            
            current_total = len(self._idle_connections) + len(self._active_connections)
            while current_total < self.config.min_connections:
                new_conn_info = await self._create_connection()
                self._idle_connections.append(new_conn_info)
                current_total += 1
            
            self._stats.idle_connections = len(self._idle_connections)
            self._not_empty.notify_all()

    @asynccontextmanager
    async def connection(self) -> AsyncGenerator[Connection, None]:
        conn = await self.acquire()
        try:
            yield conn
        finally:
            await self.release(conn)

    def get_stats(self) -> PoolStats:
        return PoolStats(
            active_connections=self._stats.active_connections,
            idle_connections=len(self._idle_connections),
            waiting_requests=self._stats.waiting_requests,
            total_created=self._stats.total_created,
            total_destroyed=self._stats.total_destroyed
        )
