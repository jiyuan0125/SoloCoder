import asyncio
import logging
import time
import traceback
from typing import Optional, Dict, Any, List
from dataclasses import dataclass, field
from enum import Enum
import uuid

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class PoolState(Enum):
    HEALTHY = "healthy"
    DEGRADED = "degraded"


@dataclass
class PoolConfig:
    name: str
    target_url: str
    max_connections: int = 10
    wait_timeout: float = 30.0
    create_retry_count: int = 3
    create_retry_delay: float = 1.0
    leak_timeout: float = 60.0
    recovery_interval: float = 30.0

    def validate(self):
        if self.max_connections <= 0:
            raise ValueError(f"max_connections must be positive, got {self.max_connections}")
        if self.wait_timeout < 0:
            raise ValueError(f"wait_timeout cannot be negative, got {self.wait_timeout}")
        if self.create_retry_count < 0:
            raise ValueError(f"create_retry_count cannot be negative, got {self.create_retry_count}")
        if self.leak_timeout <= 0:
            raise ValueError(f"leak_timeout must be positive, got {self.leak_timeout}")
        if self.recovery_interval <= 0:
            raise ValueError(f"recovery_interval must be positive, got {self.recovery_interval}")


@dataclass
class LeasedConnection:
    conn_id: str
    borrow_time: float
    caller_info: str
    acquired: bool = True

    def elapsed(self) -> float:
        return time.time() - self.borrow_time


class ConnectionPool:
    def __init__(self, config: PoolConfig):
        config.validate()
        self.config = config
        self._idle: List[Any] = []
        self._leased: Dict[str, LeasedConnection] = {}
        self._state: PoolState = PoolState.HEALTHY
        self._lock = asyncio.Lock()
        self._not_empty = asyncio.Condition(self._lock)
        self._recovery_task: Optional[asyncio.Task] = None
        self._leak_detector_task: Optional[asyncio.Task] = None
        self._initialized = False
        self._closing = False
        self._total_connections = 0

    async def initialize(self):
        if self._initialized:
            return
        async with self._lock:
            if self._initialized:
                return
            self._recovery_task = asyncio.create_task(self._recovery_loop())
            self._leak_detector_task = asyncio.create_task(self._leak_detection_loop())
            self._initialized = True
            logger.info(f"Connection pool '{self.config.name}' initialized with {self.config.max_connections} max connections")

    async def close(self):
        self._closing = True
        if self._recovery_task:
            self._recovery_task.cancel()
        if self._leak_detector_task:
            self._leak_detector_task.cancel()
        
        async with self._lock:
            for conn in self._idle:
                try:
                    await self._close_connection(conn)
                except Exception as e:
                    logger.warning(f"Error closing idle connection: {e}")
            self._idle.clear()
            
            for conn_id, leased in list(self._leased.items()):
                logger.warning(f"Force-closing leased connection {conn_id} borrowed by {leased.caller_info}")
            self._leased.clear()
        
        logger.info(f"Connection pool '{self.config.name}' closed")

    def _get_caller_info(self) -> str:
        frames = traceback.extract_stack()
        relevant_frames = []
        for frame in frames:
            if "pool.py" not in frame.filename:
                relevant_frames.append(frame)
        
        if relevant_frames:
            frame = relevant_frames[-1]
            return f"{frame.filename}:{frame.lineno}:{frame.name}"
        return "unknown"

    async def _create_connection(self) -> Any:
        last_error = None
        for attempt in range(self.config.create_retry_count + 1):
            try:
                logger.info(f"Attempting to create connection to {self.config.target_url} (attempt {attempt + 1})")
                reader, writer = await asyncio.wait_for(
                    asyncio.open_connection(
                        self.config.target_url.split('://')[-1].split(':')[0],
                        int(self.config.target_url.split(':')[-1]) if ':' in self.config.target_url.split('://')[-1] else 80
                    ),
                    timeout=5.0
                )
                writer.close()
                await writer.wait_closed()
                return {"url": self.config.target_url, "created_at": time.time()}
            except Exception as e:
                last_error = e
                logger.warning(f"Connection creation attempt {attempt + 1} failed: {e}")
                if attempt < self.config.create_retry_count:
                    await asyncio.sleep(self.config.create_retry_delay)
        
        raise last_error

    async def _close_connection(self, conn: Any):
        pass

    async def acquire(self, wait_timeout: Optional[float] = None, caller_info: Optional[str] = None) -> str:
        if self._closing:
            raise RuntimeError("Pool is closing")

        effective_timeout = wait_timeout if wait_timeout is not None else self.config.wait_timeout
        caller = caller_info or self._get_caller_info()
        
        async with self._lock:
            if self._state == PoolState.DEGRADED:
                raise RuntimeError("Pool is in degraded state: backend unreachable")

            while len(self._idle) == 0 and self._total_connections >= self.config.max_connections:
                if effective_timeout == 0:
                    raise RuntimeError("Pool is full and wait_timeout is 0")
                
                try:
                    await asyncio.wait_for(
                        self._not_empty.wait(),
                        timeout=effective_timeout
                    )
                except asyncio.TimeoutError:
                    raise RuntimeError(f"Timeout waiting for available connection after {effective_timeout}s")

            if len(self._idle) > 0:
                conn = self._idle.pop()
            else:
                try:
                    conn = await self._create_connection()
                    self._total_connections += 1
                except Exception as e:
                    logger.error(f"Failed to create connection after retries, entering degraded state: {e}")
                    self._state = PoolState.DEGRADED
                    raise RuntimeError(f"Pool is in degraded state: backend unreachable ({e})")

            conn_id = str(uuid.uuid4())
            self._leased[conn_id] = LeasedConnection(
                conn_id=conn_id,
                borrow_time=time.time(),
                caller_info=caller
            )
            
            logger.info(f"Acquired connection {conn_id} for {caller}")
            return conn_id

    async def release(self, conn_id: str):
        if self._closing:
            return

        async with self._lock:
            if conn_id not in self._leased:
                logger.warning(f"Attempted to release unknown connection {conn_id}")
                return

            leased = self._leased.pop(conn_id)
            elapsed = leased.elapsed()
            
            conn = {"url": self.config.target_url, "created_at": time.time()}
            self._idle.append(conn)
            self._not_empty.notify()
            
            logger.info(f"Released connection {conn_id} after {elapsed:.2f}s")

    async def _leak_detection_loop(self):
        while not self._closing:
            try:
                await asyncio.sleep(self.config.leak_timeout / 2)
                
                async with self._lock:
                    current_time = time.time()
                    leaked_connections = []
                    
                    for conn_id, leased in list(self._leased.items()):
                        if leased.elapsed() > self.config.leak_timeout:
                            leaked_connections.append((conn_id, leased))
                    
                    for conn_id, leased in leaked_connections:
                        logger.error(
                            f"LEAK DETECTED: Connection {conn_id} borrowed at {time.strftime('%Y-%m-%d %H:%M:%S', time.localtime(leased.borrow_time))} "
                            f"by {leased.caller_info} has exceeded timeout of {self.config.leak_timeout}s. "
                            f"Elapsed: {leased.elapsed():.2f}s"
                        )
                        
                        self._leased.pop(conn_id)
                        self._total_connections -= 1
                        
                        try:
                            new_conn = await self._create_connection()
                            self._total_connections += 1
                            self._idle.append(new_conn)
                            self._not_empty.notify()
                            logger.info(f"Created replacement connection for leaked {conn_id}")
                        except Exception as e:
                            logger.warning(f"Failed to create replacement connection for leaked {conn_id}: {e}")
                            
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in leak detection loop: {e}")
                await asyncio.sleep(1.0)

    async def _recovery_loop(self):
        while not self._closing:
            try:
                await asyncio.sleep(self.config.recovery_interval)
                
                if self._state != PoolState.DEGRADED:
                    continue

                logger.info(f"Attempting to recover pool '{self.config.name}' from degraded state")
                
                try:
                    test_conn = await self._create_connection()
                    await self._close_connection(test_conn)
                    
                    async with self._lock:
                        self._state = PoolState.HEALTHY
                        self._total_connections = 0
                        self._idle.clear()
                        self._leased.clear()
                        
                        for _ in range(min(2, self.config.max_connections)):
                            try:
                                conn = await self._create_connection()
                                self._idle.append(conn)
                                self._total_connections += 1
                            except Exception as e:
                                logger.warning(f"Failed to create initial connection during recovery: {e}")
                    
                    logger.info(f"Pool '{self.config.name}' recovered from degraded state")
                    
                except Exception as e:
                    logger.warning(f"Recovery attempt failed, pool remains degraded: {e}")
                    
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error(f"Error in recovery loop: {e}")
                await asyncio.sleep(1.0)

    def get_stats(self) -> Dict[str, Any]:
        return {
            "name": self.config.name,
            "target_url": self.config.target_url,
            "state": self._state.value,
            "max_connections": self.config.max_connections,
            "idle_connections": len(self._idle),
            "leased_connections": len(self._leased),
            "total_connections": self._total_connections,
            "config": {
                "wait_timeout": self.config.wait_timeout,
                "create_retry_count": self.config.create_retry_count,
                "leak_timeout": self.config.leak_timeout,
                "recovery_interval": self.config.recovery_interval
            }
        }

    @property
    def state(self) -> PoolState:
        return self._state

    @property
    def idle_count(self) -> int:
        return len(self._idle)

    @property
    def leased_count(self) -> int:
        return len(self._leased)


class PoolManager:
    def __init__(self):
        self._pools: Dict[str, ConnectionPool] = {}
        self._lock = asyncio.Lock()

    async def create_pool(self, config: PoolConfig) -> ConnectionPool:
        try:
            config.validate()
        except ValueError as e:
            raise ValueError(str(e))

        async with self._lock:
            if config.name in self._pools:
                raise ValueError(f"Pool '{config.name}' already exists")

            pool = ConnectionPool(config)
            await pool.initialize()
            self._pools[config.name] = pool
            return pool

    async def remove_pool(self, name: str):
        async with self._lock:
            if name not in self._pools:
                return
            pool = self._pools.pop(name)
            await pool.close()

    def get_pool(self, name: str) -> Optional[ConnectionPool]:
        return self._pools.get(name)

    def get_all_pools(self) -> Dict[str, ConnectionPool]:
        return dict(self._pools)

    async def close_all(self):
        async with self._lock:
            for name, pool in self._pools.items():
                await pool.close()
            self._pools.clear()
