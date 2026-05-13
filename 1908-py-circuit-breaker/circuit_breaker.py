from __future__ import annotations

import time
from collections import deque
from dataclasses import dataclass, field
from enum import Enum
from typing import Deque, List, Optional


class CircuitState(Enum):
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"
    UNCONFIGURED = "unconfigured"


class TriggerReason(Enum):
    FAILURE_RATE = "失败率"
    SLOW_CALL_RATE = "慢调用比例"
    PROBE_SUCCESS = "探测成功"
    PROBE_FAILURE = "探测失败"
    CONFIG_CHANGE = "配置变更"


@dataclass
class CallRecord:
    timestamp: float
    is_success: bool
    is_slow: bool


@dataclass
class CircuitEvent:
    service_name: str
    from_state: str
    to_state: str
    trigger_reason: str
    timestamp: float


@dataclass
class CircuitConfig:
    failure_rate_threshold: float = 0.5
    slow_call_rate_threshold: float = 0.6
    slow_call_threshold_ms: int = 1000
    window_size: int = 20
    half_open_probes: int = 3
    open_duration_seconds: int = 10


@dataclass
class CircuitBreaker:
    service_name: str
    config: CircuitConfig = field(default_factory=CircuitConfig)
    state: CircuitState = CircuitState.CLOSED
    call_window: Deque[CallRecord] = field(default_factory=lambda: deque(maxlen=20))
    events: List[CircuitEvent] = field(default_factory=list)
    half_open_probes_allowed: int = 0
    half_open_probes_count: int = 0
    open_until: float = 0
    is_configured: bool = False

    def __post_init__(self) -> None:
        self.call_window = deque(maxlen=self.config.window_size)

    def _record_event(self, from_state: str, to_state: str, reason: TriggerReason) -> None:
        event = CircuitEvent(
            service_name=self.service_name,
            from_state=from_state,
            to_state=to_state,
            trigger_reason=reason.value,
            timestamp=time.time(),
        )
        self.events.append(event)

    def _transition_to(self, new_state: CircuitState, reason: TriggerReason) -> None:
        if self.state == new_state:
            return
        self._record_event(
            from_state=self.state.value,
            to_state=new_state.value,
            reason=reason,
        )
        self.state = new_state

    def _check_should_open(self) -> Optional[TriggerReason]:
        if len(self.call_window) < self.config.window_size:
            return None
        total = len(self.call_window)
        failures = sum(1 for r in self.call_window if not r.is_success)
        slow = sum(1 for r in self.call_window if r.is_slow)
        failure_rate = failures / total
        slow_rate = slow / total
        if failure_rate >= self.config.failure_rate_threshold:
            return TriggerReason.FAILURE_RATE
        if slow_rate >= self.config.slow_call_rate_threshold:
            return TriggerReason.SLOW_CALL_RATE
        return None

    def _transition_to_half_open(self) -> None:
        self.half_open_probes_allowed = self.config.half_open_probes
        self.half_open_probes_count = 0
        self._transition_to(CircuitState.HALF_OPEN, TriggerReason.PROBE_SUCCESS)

    def can_accept_call(self) -> bool:
        if not self.is_configured:
            return True
        if self.state == CircuitState.CLOSED:
            return True
        if self.state == CircuitState.OPEN:
            if time.time() >= self.open_until:
                self._transition_to_half_open()
                return True
            return False
        if self.state == CircuitState.HALF_OPEN:
            return self.half_open_probes_count < self.half_open_probes_allowed
        return True

    def _close_circuit(self, reason: TriggerReason) -> None:
        self.call_window.clear()
        self._transition_to(CircuitState.CLOSED, reason)

    def _open_circuit(self, reason: TriggerReason) -> None:
        self._transition_to(CircuitState.OPEN, reason)
        self.open_until = time.time() + self.config.open_duration_seconds

    def record_call(self, is_success: bool, duration_ms: float) -> None:
        if not self.is_configured:
            return
        is_slow = duration_ms >= self.config.slow_call_threshold_ms
        self.call_window.append(
            CallRecord(
                timestamp=time.time(),
                is_success=is_success,
                is_slow=is_slow,
            )
        )
        if self.state == CircuitState.CLOSED:
            reason = self._check_should_open()
            if reason:
                self._open_circuit(reason)
                return
        elif self.state == CircuitState.HALF_OPEN:
            self.half_open_probes_count += 1
            if not is_success:
                self._open_circuit(TriggerReason.PROBE_FAILURE)
                return
            if self.half_open_probes_count >= self.half_open_probes_allowed:
                all_succeeded = all(r.is_success for r in list(self.call_window)[-self.half_open_probes_allowed:])
                if all_succeeded:
                    self._close_circuit(TriggerReason.PROBE_SUCCESS)
