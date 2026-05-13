import asyncio
import logging
import time
from abc import ABC, abstractmethod
from typing import Optional

from .config import ProtocolType

logger = logging.getLogger(__name__)


class ConnectionError(Exception):
    pass


class PooledConnection(ABC):
    def __init__(self, datasource_name: str, protocol: ProtocolType, host: str, port: int):
        self.datasource_name: str = datasource_name
        self.protocol: ProtocolType = protocol
        self.host: str = host
        self.port: int = port
        self.id: str = f"{datasource_name}-{id(self)}"
        self.is_open: bool = False
        self.created_at: float = time.monotonic()
        self.last_used_at: float = self.created_at
        self.last_health_check_at: Optional[float] = None
        self.health_fail_count: int = 0
        self._is_healthy: bool = True
        self._in_use: bool = False
        self._borrowed_at: Optional[float] = None

    @property
    def is_healthy(self) -> bool:
        return self._is_healthy and self.is_open

    @property
    def in_use(self) -> bool:
        return self._in_use

    @property
    def idle_duration(self) -> float:
        if self._in_use:
            return 0.0
        return time.monotonic() - self.last_used_at

    @property
    def borrowed_duration(self) -> float:
        if not self._in_use or self._borrowed_at is None:
            return 0.0
        return time.monotonic() - self._borrowed_at

    @abstractmethod
    async def _do_open(self) -> None:
        pass

    @abstractmethod
    async def _do_close(self) -> None:
        pass

    @abstractmethod
    async def _do_health_check(self) -> bool:
        pass

    async def open(self) -> None:
        if self.is_open:
            return
        try:
            await self._do_open()
            self.is_open = True
            self._is_healthy = True
            logger.debug("Connection %s opened", self.id)
        except Exception as e:
            self.is_open = False
            self._is_healthy = False
            raise ConnectionError(f"Failed to open connection {self.id}: {e}") from e

    async def close(self) -> None:
        if not self.is_open:
            return
        try:
            await self._do_close()
        except Exception:
            pass
        finally:
            self.is_open = False
            self._in_use = False
            self._is_healthy = False
            logger.debug("Connection %s closed", self.id)

    async def health_check(self) -> bool:
        if not self.is_open:
            self._is_healthy = False
            return False
        try:
            result = await self._do_health_check()
            self.last_health_check_at = time.monotonic()
            if result:
                self.health_fail_count = 0
                self._is_healthy = True
            else:
                self.health_fail_count += 1
        except Exception:
            self.health_fail_count += 1
            self.last_health_check_at = time.monotonic()
        self._is_healthy = self.health_fail_count == 0
        return self._is_healthy

    def mark_borrowed(self) -> None:
        self._in_use = True
        self._borrowed_at = time.monotonic()

    def mark_returned(self) -> None:
        self._in_use = False
        self._borrowed_at = None
        self.last_used_at = time.monotonic()


class TCPConnection(PooledConnection):
    def __init__(self, datasource_name: str, protocol: ProtocolType, host: str, port: int):
        super().__init__(datasource_name, protocol, host, port)
        self._reader: Optional[asyncio.StreamReader] = None
        self._writer: Optional[asyncio.StreamWriter] = None

    async def _do_open(self) -> None:
        self._reader, self._writer = await asyncio.open_connection(self.host, self.port)

    async def _do_close(self) -> None:
        if self._writer:
            try:
                self._writer.close()
                await self._writer.wait_closed()
            except Exception:
                pass
        self._reader = None
        self._writer = None

    async def _do_health_check(self) -> bool:
        if not self._writer or not self._reader:
            return False
        try:
            self._writer.write(b"")
            await self._writer.drain()
            return True
        except Exception:
            return False


class RedisConnection(PooledConnection):
    def __init__(self, datasource_name: str, protocol: ProtocolType, host: str, port: int):
        super().__init__(datasource_name, protocol, host, port)
        self._conn = None

    async def _do_open(self) -> None:
        try:
            import aioredis
        except ImportError as e:
            raise ConnectionError("aioredis is required for Redis connections") from e
        self._conn = aioredis.from_url(f"redis://{self.host}:{self.port}")
        await self._conn.ping()

    async def _do_close(self) -> None:
        if self._conn:
            try:
                await self._conn.close()
            except Exception:
                pass
        self._conn = None

    async def _do_health_check(self) -> bool:
        if not self._conn:
            return False
        try:
            await self._conn.ping()
            return True
        except Exception:
            return False


class HTTPConnection(PooledConnection):
    def __init__(self, datasource_name: str, protocol: ProtocolType, host: str, port: int):
        super().__init__(datasource_name, protocol, host, port)
        self._session = None

    async def _do_open(self) -> None:
        try:
            import aiohttp
        except ImportError as e:
            raise ConnectionError("aiohttp is required for HTTP connections") from e
        timeout = aiohttp.ClientTimeout(total=5)
        connector = aiohttp.TCPConnector(limit=1)
        self._session = aiohttp.ClientSession(timeout=timeout, connector=connector)

    async def _do_close(self) -> None:
        if self._session:
            try:
                await self._session.close()
            except Exception:
                pass
        self._session = None

    async def _do_health_check(self) -> bool:
        if not self._session:
            return False
        try:
            base_url = f"{self.protocol.value}://{self.host}:{self.port}"
            async with self._session.head(base_url, timeout=aiohttp.ClientTimeout(total=2)) as _:
                return True
        except Exception:
            return False


def create_connection(
    datasource_name: str, protocol: ProtocolType, host: str, port: int
) -> PooledConnection:
    if protocol == ProtocolType.HTTP:
        return HTTPConnection(datasource_name, protocol, host, port)
    elif protocol == ProtocolType.REDIS:
        return RedisConnection(datasource_name, protocol, host, port)
    elif protocol == ProtocolType.TCP:
        return TCPConnection(datasource_name, protocol, host, port)
    else:
        raise ValueError(f"Unsupported protocol: {protocol}")
