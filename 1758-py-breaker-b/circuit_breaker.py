import asyncio
import time
from dataclasses import dataclass, field
from enum import Enum
from typing import Callable, Dict, List, Optional
from datetime import datetime, timezone


class CircuitState(Enum):
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"


@dataclass
class StateChangeRecord:
    service: str
    from_state: CircuitState
    to_state: CircuitState
    reason: str
    timestamp: str = field(default_factory=lambda: datetime.now(timezone.utc).isoformat())


@dataclass
class ServiceStatus:
    service: str
    state: CircuitState
    failure_count: int
    success_count: int
    last_failure_time: Optional[float]
    last_state_change: str


class CircuitBreaker:
    def __init__(
        self,
        service: str,
        failure_threshold: int = 5,
        recovery_timeout: float = 30.0,
        success_threshold: int = 3,
        on_state_change: Optional[Callable[[StateChangeRecord], None]] = None,
    ):
        self.service = service
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.success_threshold = success_threshold
        self.on_state_change = on_state_change

        self._state = CircuitState.CLOSED
        self._failure_count = 0
        self._success_count = 0
        self._last_failure_time: Optional[float] = None
        self._last_state_change = datetime.now(timezone.utc).isoformat()
        self._lock = asyncio.Lock()

    def _record_state_change(self, from_state: CircuitState, to_state: CircuitState, reason: str):
        self._last_state_change = datetime.now(timezone.utc).isoformat()
        if self.on_state_change:
            record = StateChangeRecord(
                service=self.service,
                from_state=from_state,
                to_state=to_state,
                reason=reason,
            )
            self.on_state_change(record)

    def _transition_to_open(self, reason: str):
        from_state = self._state
        self._state = CircuitState.OPEN
        self._last_failure_time = time.time()
        self._record_state_change(from_state, CircuitState.OPEN, reason)

    def _transition_to_half_open(self, reason: str):
        from_state = self._state
        self._state = CircuitState.HALF_OPEN
        self._success_count = 0
        self._record_state_change(from_state, CircuitState.HALF_OPEN, reason)

    def _transition_to_closed(self, reason: str):
        from_state = self._state
        self._state = CircuitState.CLOSED
        self._failure_count = 0
        self._record_state_change(from_state, CircuitState.CLOSED, reason)

    def _can_try_half_open(self) -> bool:
        if self._state != CircuitState.OPEN:
            return False
        if self._last_failure_time is None:
            return False
        return (time.time() - self._last_failure_time) >= self.recovery_timeout

    def get_status(self) -> ServiceStatus:
        return ServiceStatus(
            service=self.service,
            state=self._state,
            failure_count=self._failure_count,
            success_count=self._success_count,
            last_failure_time=self._last_failure_time,
            last_state_change=self._last_state_change,
        )

    async def call(self, func, *args, fallback=None, **kwargs):
        async with self._lock:
            if self._state == CircuitState.CLOSED:
                try:
                    result = await func(*args, **kwargs)
                    self._failure_count = 0
                    return result
                except Exception as e:
                    self._failure_count += 1
                    if self._failure_count >= self.failure_threshold:
                        self._transition_to_open(
                            f"连续失败 {self._failure_count} 次，达到阈值 {self.failure_threshold}"
                        )
                    raise
            elif self._state == CircuitState.OPEN:
                if self._can_try_half_open():
                    self._transition_to_half_open("恢复超时已过，尝试半开状态")
                else:
                    if fallback is not None:
                        return fallback
                    raise CircuitBreakerOpenError(
                        f"服务 '{self.service}' 处于熔断状态，"
                        f"剩余恢复时间: {self._get_remaining_recovery_time():.1f}秒"
                    )

            if self._state == CircuitState.HALF_OPEN:
                try:
                    result = await func(*args, **kwargs)
                    self._success_count += 1
                    if self._success_count >= self.success_threshold:
                        self._transition_to_closed(
                            f"连续成功 {self._success_count} 次，达到阈值 {self.success_threshold}"
                        )
                    return result
                except Exception as e:
                    self._transition_to_open(
                        f"半开状态下调用失败，立即重新熔断"
                    )
                    raise

    def _get_remaining_recovery_time(self) -> float:
        if self._last_failure_time is None:
            return 0.0
        elapsed = time.time() - self._last_failure_time
        remaining = self.recovery_timeout - elapsed
        return max(0.0, remaining)


class CircuitBreakerOpenError(Exception):
    pass


class CircuitBreakerManager:
    def __init__(
        self,
        failure_threshold: int = 5,
        recovery_timeout: float = 30.0,
        success_threshold: int = 3,
        max_history: int = 100,
    ):
        self.failure_threshold = failure_threshold
        self.recovery_timeout = recovery_timeout
        self.success_threshold = success_threshold
        self.max_history = max_history

        self._breakers: Dict[str, CircuitBreaker] = {}
        self._history: List[StateChangeRecord] = []
        self._history_lock = asyncio.Lock()

    def _on_state_change(self, record: StateChangeRecord):
        asyncio.create_task(self._add_history(record))

    async def _add_history(self, record: StateChangeRecord):
        async with self._history_lock:
            self._history.append(record)
            if len(self._history) > self.max_history:
                self._history = self._history[-self.max_history :]

    def get_breaker(self, service: str) -> CircuitBreaker:
        if service not in self._breakers:
            self._breakers[service] = CircuitBreaker(
                service=service,
                failure_threshold=self.failure_threshold,
                recovery_timeout=self.recovery_timeout,
                success_threshold=self.success_threshold,
                on_state_change=self._on_state_change,
            )
        return self._breakers[service]

    async def call(self, service: str, func, *args, fallback=None, **kwargs):
        breaker = self.get_breaker(service)
        return await breaker.call(func, *args, fallback=fallback, **kwargs)

    def get_all_statuses(self) -> List[ServiceStatus]:
        return [breaker.get_status() for breaker in self._breakers.values()]

    def get_history(self, service: Optional[str] = None) -> List[StateChangeRecord]:
        if service:
            return [record for record in self._history if record.service == service]
        return list(self._history)
