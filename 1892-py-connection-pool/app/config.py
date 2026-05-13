import os
from dataclasses import dataclass, field
from enum import Enum
from typing import Optional


class ProtocolType(Enum):
    HTTP = "http"
    REDIS = "redis"
    TCP = "tcp"


@dataclass
class PoolConfig:
    max_size: int = 10
    min_idle: int = 1
    acquire_timeout: float = 5.0
    idle_timeout: float = 60.0
    health_check_interval: float = 30.0
    health_check_fail_threshold: int = 2
    leak_detection_threshold: float = 1800.0
    refill_rate_limit: int = 5

    def validate(self) -> None:
        if self.max_size <= 0:
            raise ValueError("max_size must be positive")
        if self.min_idle < 0:
            raise ValueError("min_idle must be non-negative")
        if self.min_idle > self.max_size:
            raise ValueError("min_idle must not exceed max_size")
        if self.acquire_timeout <= 0:
            raise ValueError("acquire_timeout must be positive")
        if self.idle_timeout <= 0:
            raise ValueError("idle_timeout must be positive")
        if self.health_check_interval <= 0:
            raise ValueError("health_check_interval must be positive")
        if self.health_check_fail_threshold <= 0:
            raise ValueError("health_check_fail_threshold must be positive")
        if self.leak_detection_threshold <= 0:
            raise ValueError("leak_detection_threshold must be positive")
        if self.refill_rate_limit <= 0:
            raise ValueError("refill_rate_limit must be positive")


@dataclass
class DatasourceConfig:
    name: str
    protocol: ProtocolType
    host: str
    port: int
    pool_config: PoolConfig = field(default_factory=PoolConfig)

    def validate(self) -> None:
        if not self.name:
            raise ValueError("datasource name cannot be empty")
        if self.port <= 0 or self.port > 65535:
            raise ValueError("port must be between 1 and 65535")
        if not self.host:
            raise ValueError("host cannot be empty")
        self.pool_config.validate()


class AppConfig:
    @property
    def port(self) -> int:
        return int(os.environ.get("PORT", "8907"))
