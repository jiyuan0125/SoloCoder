from typing import Dict, Optional
from contextlib import asynccontextmanager
from typing import AsyncGenerator

from connection_pool import ConnectionPool
from pool_config import PoolConfig, PoolStats


class PoolManager:
    def __init__(self):
        self._pools: Dict[str, ConnectionPool] = {}
    
    async def create_pool(self, name: str, config: PoolConfig) -> ConnectionPool:
        if name in self._pools:
            raise ValueError(f"Pool '{name}' already exists")
        
        pool = ConnectionPool(config, name)
        await pool.initialize()
        self._pools[name] = pool
        return pool
    
    def get_pool(self, name: str) -> Optional[ConnectionPool]:
        return self._pools.get(name)
    
    def has_pool(self, name: str) -> bool:
        return name in self._pools
    
    async def remove_pool(self, name: str) -> None:
        if name in self._pools:
            pool = self._pools.pop(name)
            await pool.close()
    
    async def close_all(self) -> None:
        for name in list(self._pools.keys()):
            await self.remove_pool(name)
    
    def get_all_stats(self) -> Dict[str, PoolStats]:
        return {name: pool.get_stats() for name, pool in self._pools.items()}
    
    @asynccontextmanager
    async def connection(self, pool_name: str) -> AsyncGenerator:
        pool = self.get_pool(pool_name)
        if not pool:
            raise ValueError(f"Pool '{pool_name}' does not exist")
        
        async with pool.connection() as conn:
            yield conn
