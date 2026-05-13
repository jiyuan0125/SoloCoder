import asyncio
import logging
from typing import Dict, List, Optional, Any

from .config import DatasourceConfig, PoolConfig
from .pool import ConnectionPool, PoolClosedError, PoolDrainingError, PoolAcquireTimeoutError
from .connection import PooledConnection

logger = logging.getLogger(__name__)


class DatasourceNotFoundError(Exception):
    pass


class DatasourceAlreadyExistsError(Exception):
    pass


class DatasourceManager:
    def __init__(self):
        self._pools: Dict[str, ConnectionPool] = {}
        self._lock: asyncio.Lock = asyncio.Lock()
        self._drain_tasks: Dict[str, asyncio.Task] = {}

    async def register(self, config: DatasourceConfig) -> None:
        async with self._lock:
            if config.name in self._pools:
                raise DatasourceAlreadyExistsError(f"Datasource '{config.name}' already exists")

            pool = ConnectionPool(config)
            await pool.start()
            self._pools[config.name] = pool
            logger.info("Registered datasource '%s' with protocol %s", config.name, config.protocol.value)

    async def unregister(self, name: str, wait: bool = True) -> None:
        async with self._lock:
            if name not in self._pools:
                raise DatasourceNotFoundError(f"Datasource '{name}' not found")

            pool = self._pools[name]
            del self._pools[name]

        drain_task = asyncio.create_task(pool.drain())
        self._drain_tasks[name] = drain_task

        async def _drain_and_close(pool: ConnectionPool, task_ref: asyncio.Task):
            try:
                await task_ref
            finally:
                await pool.close()
                if name in self._drain_tasks and self._drain_tasks[name] is task_ref:
                    del self._drain_tasks[name]
                logger.info("Unregistered datasource '%s'", name)

        cleanup_task = asyncio.create_task(_drain_and_close(pool, drain_task))

        if wait:
            await cleanup_task

    async def get_pool(self, name: str) -> ConnectionPool:
        async with self._lock:
            if name not in self._pools:
                raise DatasourceNotFoundError(f"Datasource '{name}' not found")
            return self._pools[name]

    async def acquire_connection(self, name: str) -> PooledConnection:
        pool = await self.get_pool(name)
        try:
            return await pool.acquire()
        except PoolDrainingError:
            raise DatasourceNotFoundError(f"Datasource '{name}' is being drained") from None

    async def release_connection(self, name: str, conn: PooledConnection) -> None:
        try:
            pool = await self.get_pool(name)
            await pool.release(conn)
        except DatasourceNotFoundError:
            await conn.close()

    async def update_pool_config(self, name: str, new_config: PoolConfig) -> None:
        pool = await self.get_pool(name)
        await pool.update_pool_config(new_config)

    async def get_stats(self, name: Optional[str] = None) -> Any:
        async with self._lock:
            if name:
                if name not in self._pools:
                    raise DatasourceNotFoundError(f"Datasource '{name}' not found")
                return self._pools[name].stats()

            result = {}
            for pool_name, pool in self._pools.items():
                result[pool_name] = pool.stats()
            return result

    async def list_pools(self) -> List[str]:
        async with self._lock:
            return list(self._pools.keys())

    async def close_all(self) -> None:
        async with self._lock:
            pools = list(self._pools.values())
            self._pools.clear()

        drain_tasks = []
        for pool in pools:
            drain_tasks.append(asyncio.create_task(pool.drain()))

        for task in drain_tasks:
            try:
                await task
            except Exception:
                pass

        for pool in pools:
            await pool.close()

        for name, task in list(self._drain_tasks.items()):
            try:
                task.cancel()
            except Exception:
                pass
        self._drain_tasks.clear()

    async def connection(self, name: str):
        class _ConnectionContext:
            def __init__(self, manager: DatasourceManager, pool_name: str):
                self.manager = manager
                self.pool_name = pool_name
                self.conn: Optional[PooledConnection] = None

            async def __aenter__(self) -> PooledConnection:
                self.conn = await self.manager.acquire_connection(self.pool_name)
                return self.conn

            async def __aexit__(self, exc_type, exc, tb):
                if self.conn:
                    await self.manager.release_connection(self.pool_name, self.conn)

        return _ConnectionContext(self, name)
