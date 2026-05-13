from __future__ import annotations

import os
from typing import Dict, Optional

from circuit_breaker import CircuitBreaker, CircuitConfig, CircuitState


class CircuitBreakerManager:
    def __init__(self) -> None:
        self._breakers: Dict[str, CircuitBreaker] = {}

    def get_or_create(self, service_name: str) -> CircuitBreaker:
        if service_name not in self._breakers:
            self._breakers[service_name] = CircuitBreaker(service_name=service_name)
        return self._breakers[service_name]

    def configure(
        self,
        service_name: str,
        config: CircuitConfig,
    ) -> CircuitBreaker:
        breaker = self.get_or_create(service_name)
        breaker.config = config
        breaker.call_window.maxlen = config.window_size
        breaker.is_configured = True
        if breaker.state == CircuitState.UNCONFIGURED:
            breaker.state = CircuitState.CLOSED
        return breaker

    def list_services(self) -> Dict[str, str]:
        result = {}
        for name, breaker in self._breakers.items():
            if breaker.is_configured:
                result[name] = breaker.state.value
            else:
                result[name] = "unconfigured"
        return result


manager = CircuitBreakerManager()
