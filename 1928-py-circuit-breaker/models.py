from datetime import datetime
from enum import Enum
from typing import List, Optional
from pydantic import BaseModel, Field


class CircuitBreakerState(str, Enum):
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"


class DegradationState(str, Enum):
    NORMAL = "normal"
    DEGRADED = "degraded"


class ServiceStatus(BaseModel):
    name: str
    circuit_breaker_state: CircuitBreakerState = CircuitBreakerState.CLOSED
    degradation_state: DegradationState = DegradationState.NORMAL
    original_timeout: float = 5.0
    current_timeout: float = 5.0
    degraded_timeout: float = 1.0
    failure_count: int = 0
    last_failure_time: Optional[datetime] = None
    last_state_change: datetime = Field(default_factory=datetime.utcnow)
    half_open_probe_in_progress: bool = False
    degraded_by_downstream: Optional[List[str]] = None


class DependencyConfig(BaseModel):
    upstream: str
    downstream: List[str]
    degraded_timeout: float = 1.0


class DependencyRecord(BaseModel):
    id: str
    upstream: str
    downstream: List[str]
    degraded_timeout: float = 1.0


class DependencyGraph(BaseModel):
    dependencies: List[DependencyRecord] = []


class StatusResponse(BaseModel):
    services: dict
    dependency_graph: DependencyGraph
