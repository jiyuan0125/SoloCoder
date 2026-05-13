import asyncio
import logging
from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional
from datetime import datetime, timedelta
import aiopg
import psycopg2

logger = logging.getLogger(__name__)


class AddressStatus(Enum):
    ACTIVE = "active"
    FAILED = "failed"
    COOLING_DOWN = "cooling_down"


@dataclass
class AddressConfig:
    host: str
    port: int
    user: str
    password: str
    database: str


@dataclass
class AddressState:
    config: AddressConfig
    status: AddressStatus = AddressStatus.ACTIVE
    consecutive_failures: int = 0
    total_failures: int = 0
    last_failure_time: Optional[datetime] = None
    cooldown_until: Optional[datetime] = None
    active_connections: int = 0
    pool: Optional[aiopg.Pool] = None


@dataclass
class PoolConfig:
    name: str
    addresses: List[AddressConfig]
    max_connections: int = 10
    connection_timeout: int = 5
    failure_threshold: int = 3
    cooldown_period: int = 60
    all_failed_wait_period: int = 30


class FailoverConnectionPool:
    def __init__(self, config: PoolConfig):
        self._config = config
        self._addresses: List[AddressState] = [
            AddressState(config=addr) for addr in config.addresses
        ]
        self._current_index: int = 0
        self._all_failed: bool = False
        self._all_failed_until: Optional[datetime] = None
        self._lock = asyncio.Lock()
        self._closed = False
        self._config_lock = asyncio.Lock()

    @property
    def config(self) -> PoolConfig:
        return self._config

    async def update_config(
        self,
        max_connections: Optional[int] = None,
        connection_timeout: Optional[int] = None,
        failure_threshold: Optional[int] = None,
        cooldown_period: Optional[int] = None,
        all_failed_wait_period: Optional[int] = None,
    ) -> None:
        async with self._config_lock:
            if max_connections is not None:
                self._config.max_connections = max_connections
            if connection_timeout is not None:
                self._config.connection_timeout = connection_timeout
            if failure_threshold is not None:
                self._config.failure_threshold = failure_threshold
            if cooldown_period is not None:
                self._config.cooldown_period = cooldown_period
            if all_failed_wait_period is not None:
                self._config.all_failed_wait_period = all_failed_wait_period

    def _check_all_failed_wait_period(self) -> bool:
        if not self._all_failed:
            return False
        
        now = datetime.now()
        if self._all_failed_until and now >= self._all_failed_until:
            self._reset_all_addresses()
            self._all_failed = False
            self._all_failed_until = None
            logger.info(f"All failed wait period elapsed for pool {self._config.name}, resetting all addresses")
            return False
        
        return True

    def _reset_all_addresses(self) -> None:
        for addr in self._addresses:
            addr.status = AddressStatus.ACTIVE
            addr.consecutive_failures = 0
        self._current_index = 0

    def _get_next_available_address(self) -> Optional[int]:
        now = datetime.now()
        
        for i, addr in enumerate(self._addresses):
            if addr.status == AddressStatus.ACTIVE:
                return i
            
            if addr.status == AddressStatus.COOLING_DOWN:
                if addr.cooldown_until and now >= addr.cooldown_until:
                    addr.status = AddressStatus.ACTIVE
                    addr.consecutive_failures = 0
                    logger.info(f"Address {addr.config.host}:{addr.config.port} cooldown period elapsed, marked as active")
                    return i
        
        return None

    def _mark_address_failed(self, index: int) -> None:
        addr = self._addresses[index]
        addr.consecutive_failures += 1
        addr.total_failures += 1
        addr.last_failure_time = datetime.now()
        
        if addr.consecutive_failures >= self._config.failure_threshold:
            addr.status = AddressStatus.COOLING_DOWN
            addr.cooldown_until = datetime.now() + timedelta(seconds=self._config.cooldown_period)
            logger.warning(
                f"Address {addr.config.host}:{addr.config.port} marked as cooling down "
                f"after {addr.consecutive_failures} consecutive failures"
            )

    def _mark_address_active(self, index: int) -> None:
        addr = self._addresses[index]
        addr.status = AddressStatus.ACTIVE
        addr.consecutive_failures = 0

    async def _create_pool_for_address(self, index: int) -> aiopg.Pool:
        addr = self._addresses[index]
        config = addr.config
        
        dsn = f"dbname={config.database} user={config.user} password={config.password} host={config.host} port={config.port}"
        
        pool = await aiopg.create_pool(
            dsn=dsn,
            minsize=1,
            maxsize=self._config.max_connections,
            timeout=self._config.connection_timeout,
        )
        
        addr.pool = pool
        return pool

    async def _try_acquire_from_address(self, index: int):
        addr = self._addresses[index]
        
        if addr.pool is None:
            await self._create_pool_for_address(index)
        
        try:
            conn = await asyncio.wait_for(
                addr.pool.acquire(),
                timeout=self._config.connection_timeout
            )
            self._mark_address_active(index)
            addr.active_connections += 1
            return conn, index
        except (asyncio.TimeoutError, psycopg2.OperationalError, Exception) as e:
            self._mark_address_failed(index)
            logger.error(f"Failed to acquire connection from {addr.config.host}:{addr.config.port}: {e}")
            raise

    async def acquire(self):
        if self._closed:
            raise RuntimeError("Pool is closed")

        async with self._lock:
            if self._check_all_failed_wait_period():
                remaining_seconds = int((self._all_failed_until - datetime.now()).total_seconds())
                raise RuntimeError(
                    f"All addresses are failed. Waiting period active, {remaining_seconds} seconds remaining"
                )

            attempts = 0
            
            while attempts < len(self._addresses):
                next_idx = self._get_next_available_address()
                
                if next_idx is None:
                    self._all_failed = True
                    self._all_failed_until = datetime.now() + timedelta(
                        seconds=self._config.all_failed_wait_period
                    )
                    logger.error(
                        f"All addresses failed for pool {self._config.name}. "
                        f"Entering wait period for {self._config.all_failed_wait_period} seconds"
                    )
                    raise RuntimeError(
                        f"All addresses failed. Will retry from primary after "
                        f"{self._config.all_failed_wait_period} seconds"
                    )

                try:
                    conn, idx = await self._try_acquire_from_address(next_idx)
                    self._current_index = idx
                    return conn, idx
                except Exception:
                    attempts += 1
                    continue
            
            self._all_failed = True
            self._all_failed_until = datetime.now() + timedelta(
                seconds=self._config.all_failed_wait_period
            )
            raise RuntimeError(
                f"All addresses failed. Will retry from primary after "
                f"{self._config.all_failed_wait_period} seconds"
            )

    async def release(self, conn, address_index: int) -> None:
        if self._closed:
            return
        
        addr = self._addresses[address_index]
        if addr.pool:
            await addr.pool.release(conn)
            if addr.active_connections > 0:
                addr.active_connections -= 1

    async def close(self) -> None:
        self._closed = True
        for addr in self._addresses:
            if addr.pool:
                addr.pool.close()
                await addr.pool.wait_closed()
                addr.pool = None

    def get_status(self) -> Dict[str, Any]:
        now = datetime.now()
        return {
            "name": self._config.name,
            "config": {
                "max_connections": self._config.max_connections,
                "connection_timeout": self._config.connection_timeout,
                "failure_threshold": self._config.failure_threshold,
                "cooldown_period": self._config.cooldown_period,
                "all_failed_wait_period": self._config.all_failed_wait_period,
            },
            "current_active_index": self._current_index,
            "all_failed": self._all_failed,
            "all_failed_remaining_seconds": (
                int((self._all_failed_until - now).total_seconds())
                if self._all_failed and self._all_failed_until
                else 0
            ),
            "addresses": [
                {
                    "host": addr.config.host,
                    "port": addr.config.port,
                    "database": addr.config.database,
                    "status": addr.status.value,
                    "consecutive_failures": addr.consecutive_failures,
                    "total_failures": addr.total_failures,
                    "active_connections": addr.active_connections,
                    "cooldown_remaining_seconds": (
                        int((addr.cooldown_until - now).total_seconds())
                        if addr.cooldown_until and addr.status == AddressStatus.COOLING_DOWN
                        else 0
                    ),
                }
                for addr in self._addresses
            ],
        }
