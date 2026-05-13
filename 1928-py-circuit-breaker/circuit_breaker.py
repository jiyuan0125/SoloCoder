import asyncio
import time
import uuid
from datetime import datetime
from typing import Dict, List, Optional, Set, Callable
from collections import defaultdict

from models import (
    CircuitBreakerState,
    DegradationState,
    ServiceStatus,
    DependencyConfig,
    DependencyRecord,
)


class CircuitBreakerManager:
    def __init__(self):
        self.services: Dict[str, ServiceStatus] = {}
        self.dependencies: Dict[str, DependencyRecord] = {}
        self.upstream_to_downstream: Dict[str, List[str]] = defaultdict(list)
        self.downstream_to_upstream: Dict[str, List[str]] = defaultdict(list)
        self._lock = asyncio.Lock()
        self._degraded_services: Set[str] = set()
        self._degraded_by_downstream: Dict[str, Set[str]] = defaultdict(set)
        self._circuit_open_handlers: List[Callable[[str], None]] = []
        self._circuit_close_handlers: List[Callable[[str], None]] = []

        self.failure_threshold = 5
        self.open_duration = 30.0
        self.half_open_timeout = 5.0
        self.default_timeout = 5.0

    def _get_or_create_service(self, service_name: str) -> ServiceStatus:
        if service_name not in self.services:
            self.services[service_name] = ServiceStatus(name=service_name)
        return self.services[service_name]

    def add_dependency(self, config: DependencyConfig) -> DependencyRecord:
        dep_id = str(uuid.uuid4())
        record = DependencyRecord(
            id=dep_id,
            upstream=config.upstream,
            downstream=config.downstream.copy(),
            degraded_timeout=config.degraded_timeout,
        )

        self.dependencies[dep_id] = record
        self.upstream_to_downstream[config.upstream].extend(config.downstream)
        for ds in config.downstream:
            self.downstream_to_upstream[ds].append(config.upstream)
            self._get_or_create_service(ds)

        upstream = self._get_or_create_service(config.upstream)
        upstream.degraded_timeout = config.degraded_timeout

        return record

    def remove_dependency(self, dep_id: str) -> bool:
        if dep_id not in self.dependencies:
            return False

        record = self.dependencies.pop(dep_id)
        upstream = record.upstream

        for ds in record.downstream:
            if ds in self.downstream_to_upstream:
                if upstream in self.downstream_to_upstream[ds]:
                    self.downstream_to_upstream[ds].remove(upstream)

        if upstream in self.upstream_to_downstream:
            for ds in record.downstream:
                if ds in self.upstream_to_downstream[upstream]:
                    self.upstream_to_downstream[upstream].remove(ds)

        return True

    def get_all_dependencies(self) -> List[DependencyRecord]:
        return list(self.dependencies.values())

    async def record_failure(self, service_name: str):
        async with self._lock:
            service = self._get_or_create_service(service_name)
            service.failure_count += 1
            service.last_failure_time = datetime.utcnow()

            if (
                service.circuit_breaker_state == CircuitBreakerState.CLOSED
                and service.failure_count >= self.failure_threshold
            ):
                await self._open_circuit(service_name)
            elif service.circuit_breaker_state == CircuitBreakerState.HALF_OPEN:
                await self._handle_half_open_failure(service_name)

    async def record_success(self, service_name: str):
        async with self._lock:
            service = self._get_or_create_service(service_name)

            if service.circuit_breaker_state == CircuitBreakerState.HALF_OPEN:
                await self._close_circuit(service_name)
            elif service.circuit_breaker_state == CircuitBreakerState.CLOSED:
                service.failure_count = max(0, service.failure_count - 1)
                if service.failure_count == 0:
                    service.last_failure_time = None

    async def _open_circuit(self, service_name: str):
        service = self._get_or_create_service(service_name)
        service.circuit_breaker_state = CircuitBreakerState.OPEN
        service.last_state_change = datetime.utcnow()
        service.failure_count = 0

        await self._trigger_degradation_for_upstreams(service_name)

        asyncio.create_task(self._schedule_half_open(service_name))

    async def _schedule_half_open(self, service_name: str):
        await asyncio.sleep(self.open_duration)
        async with self._lock:
            service = self._get_or_create_service(service_name)
            if service.circuit_breaker_state == CircuitBreakerState.OPEN:
                service.circuit_breaker_state = CircuitBreakerState.HALF_OPEN
                service.last_state_change = datetime.utcnow()
                service.half_open_probe_in_progress = True

    async def _handle_half_open_failure(self, service_name: str):
        service = self._get_or_create_service(service_name)
        service.circuit_breaker_state = CircuitBreakerState.OPEN
        service.last_state_change = datetime.utcnow()
        service.half_open_probe_in_progress = False
        asyncio.create_task(self._schedule_half_open(service_name))

    async def _close_circuit(self, service_name: str):
        service = self._get_or_create_service(service_name)
        service.circuit_breaker_state = CircuitBreakerState.CLOSED
        service.last_state_change = datetime.utcnow()
        service.half_open_probe_in_progress = False
        service.failure_count = 0
        service.last_failure_time = None

        await self._trigger_recovery_for_upstreams(service_name)

    async def _trigger_degradation_for_upstreams(self, downstream_name: str):
        upstreams = self.downstream_to_upstream.get(downstream_name, [])
        degraded_configs = self._find_dependency_configs(
            downstream_name, upstreams
        )

        for upstream, dep_record in degraded_configs:
            if upstream in self._degraded_by_downstream:
                if downstream_name in self._degraded_by_downstream[upstream]:
                    continue

            self._degraded_by_downstream[upstream].add(downstream_name)
            await self._degrade_service(upstream, dep_record.degraded_timeout)

    async def _trigger_recovery_for_upstreams(self, downstream_name: str):
        upstreams = self.downstream_to_upstream.get(downstream_name, [])

        for upstream in upstreams:
            if downstream_name in self._degraded_by_downstream[upstream]:
                self._degraded_by_downstream[upstream].remove(downstream_name)

            if not self._degraded_by_downstream[upstream]:
                await self._recover_service(upstream)

    def _find_dependency_configs(
        self, downstream_name: str, upstreams: List[str]
    ) -> List[tuple]:
        results = []
        for dep_record in self.dependencies.values():
            if (
                dep_record.upstream in upstreams
                and downstream_name in dep_record.downstream
            ):
                results.append((dep_record.upstream, dep_record))
        return results

    async def _degrade_service(self, service_name: str, degraded_timeout: float):
        service = self._get_or_create_service(service_name)

        if service.degradation_state == DegradationState.DEGRADED:
            return

        service.degradation_state = DegradationState.DEGRADED
        service.current_timeout = degraded_timeout
        service.last_state_change = datetime.utcnow()

        if service.degraded_by_downstream is None:
            service.degraded_by_downstream = []

        self._degraded_services.add(service_name)

    async def _recover_service(self, service_name: str):
        service = self._get_or_create_service(service_name)

        if service.degradation_state == DegradationState.NORMAL:
            return

        service.degradation_state = DegradationState.NORMAL
        service.current_timeout = service.original_timeout
        service.last_state_change = datetime.utcnow()

        if service_name in self._degraded_services:
            self._degraded_services.remove(service_name)

    def can_call(self, service_name: str) -> bool:
        service = self._get_or_create_service(service_name)
        return service.circuit_breaker_state != CircuitBreakerState.OPEN

    def get_current_timeout(self, service_name: str) -> float:
        service = self._get_or_create_service(service_name)
        return service.current_timeout

    def get_service_status(self, service_name: str) -> Optional[ServiceStatus]:
        if service_name not in self.services:
            return None
        service = self.services[service_name]

        status = service.model_copy()
        status.degraded_by_downstream = list(
            self._degraded_by_downstream.get(service_name, set())
        )
        return status

    def get_all_services_status(self) -> Dict[str, ServiceStatus]:
        result = {}
        for name in list(self.services.keys()) + list(self._degraded_services):
            status = self.get_service_status(name)
            if status:
                result[name] = status
        return result

    def get_dependency_graph(self) -> Dict:
        return {
            "dependencies": [
                record.model_dump() for record in self.dependencies.values()
            ]
        }
