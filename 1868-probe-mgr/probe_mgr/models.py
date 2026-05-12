import uuid
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import Any, Optional


class HealthStatus(str, Enum):
    UNKNOWN = "unknown"
    HEALTHY = "healthy"
    UNHEALTHY = "unhealthy"


class ProbeType(str, Enum):
    HTTP = "http"
    TCP = "tcp"


@dataclass
class HTTPProbeConfig:
    url: str
    method: str = "GET"
    expected_status: int = 200


@dataclass
class TCPProbeConfig:
    host: str
    port: int


@dataclass
class Target:
    id: str = field(default_factory=lambda: uuid.uuid4().hex)
    name: str = ""
    probe_type: ProbeType = ProbeType.HTTP
    http_config: Optional[HTTPProbeConfig] = None
    tcp_config: Optional[TCPProbeConfig] = None
    interval: int = 60
    timeout: int = 10
    failure_threshold: int = 3
    success_threshold: int = 3
    group_id: Optional[str] = None
    status: HealthStatus = HealthStatus.UNKNOWN
    consecutive_failures: int = 0
    consecutive_successes: int = 0
    created_at: datetime = field(default_factory=datetime.utcnow)
    last_checked_at: Optional[datetime] = None
    last_status_change_at: Optional[datetime] = None

    def to_dict(self) -> dict[str, Any]:
        probe_config = None
        if self.probe_type == ProbeType.HTTP and self.http_config:
            probe_config = {
                "url": self.http_config.url,
                "method": self.http_config.method,
                "expected_status": self.http_config.expected_status,
            }
        elif self.probe_type == ProbeType.TCP and self.tcp_config:
            probe_config = {
                "host": self.tcp_config.host,
                "port": self.tcp_config.port,
            }
        return {
            "id": self.id,
            "name": self.name,
            "probe_type": self.probe_type.value,
            "probe_config": probe_config,
            "interval": self.interval,
            "timeout": self.timeout,
            "failure_threshold": self.failure_threshold,
            "success_threshold": self.success_threshold,
            "group_id": self.group_id,
            "status": self.status.value,
            "consecutive_failures": self.consecutive_failures,
            "consecutive_successes": self.consecutive_successes,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "last_checked_at": self.last_checked_at.isoformat() if self.last_checked_at else None,
            "last_status_change_at": self.last_status_change_at.isoformat() if self.last_status_change_at else None,
        }


@dataclass
class Group:
    id: str = field(default_factory=lambda: uuid.uuid4().hex)
    name: str = ""
    parent_id: Optional[str] = None
    status: HealthStatus = HealthStatus.UNKNOWN
    created_at: datetime = field(default_factory=datetime.utcnow)

    def to_dict(self) -> dict[str, Any]:
        return {
            "id": self.id,
            "name": self.name,
            "parent_id": self.parent_id,
            "status": self.status.value,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }


@dataclass
class Callback:
    id: str = field(default_factory=lambda: uuid.uuid4().hex)
    url: str = ""
    created_at: datetime = field(default_factory=datetime.utcnow)

    def to_dict(self) -> dict[str, Any]:
        return {
            "id": self.id,
            "url": self.url,
            "created_at": self.created_at.isoformat() if self.created_at else None,
        }


@dataclass
class CheckRecord:
    id: str = field(default_factory=lambda: uuid.uuid4().hex)
    target_id: str = ""
    success: bool = False
    duration_ms: float = 0.0
    error_message: Optional[str] = None
    checked_at: datetime = field(default_factory=datetime.utcnow)

    def to_dict(self) -> dict[str, Any]:
        return {
            "id": self.id,
            "target_id": self.target_id,
            "success": self.success,
            "duration_ms": self.duration_ms,
            "error_message": self.error_message,
            "checked_at": self.checked_at.isoformat() if self.checked_at else None,
        }


@dataclass
class StatusChangeEvent:
    target_id: str
    old_status: HealthStatus
    new_status: HealthStatus
    changed_at: datetime
