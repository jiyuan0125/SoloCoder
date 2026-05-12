from dataclasses import dataclass
from typing import Callable, Optional, Awaitable

from connection import Connection


@dataclass
class PoolConfig:
    max_connections: int = 10
    idle_timeout: int = 300
    wait_timeout: int = 30
    min_usage_count: int = 0
    min_connections: int = 1
    idle_recycle_interval: int = 60
    create_connection: Optional[Callable[[], Awaitable[Connection]]] = None


@dataclass
class ConnectionInfo:
    connection: Connection
    last_used_time: float
    usage_count: int = 0


@dataclass
class PoolStats:
    active_connections: int = 0
    idle_connections: int = 0
    waiting_requests: int = 0
    total_created: int = 0
    total_destroyed: int = 0
