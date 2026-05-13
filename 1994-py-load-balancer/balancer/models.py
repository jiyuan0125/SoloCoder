import asyncio
import time
from dataclasses import dataclass, field
from typing import Optional, List


@dataclass
class HealthCheckPolicy:
    path: str = "/health"
    interval: float = 10.0
    timeout: float = 5.0
    consecutive_failures_to_down: int = 3
    consecutive_successes_to_up: int = 2


@dataclass
class NodeStats:
    total_requests: int = 0
    failed_requests: int = 0
    proxy_errors: int = 0
    backend_errors: int = 0
    total_response_time: float = 0.0
    last_health_status: Optional[str] = None
    last_health_time: Optional[float] = None

    @property
    def avg_response_time(self) -> float:
        if self.total_requests == 0:
            return 0.0
        return self.total_response_time / self.total_requests


@dataclass
class Node:
    url: str
    weight: int
    service_group: str
    original_weight: int = field(default=0)
    current_weight: int = 0
    effective_weight: int = 0
    is_healthy: bool = True
    is_pending: bool = True
    health_failures: int = 0
    health_successes: int = 0
    stats: NodeStats = field(default_factory=NodeStats)
    pending_requests: int = 0
    is_marked_for_deletion: bool = False
    deletion_time: Optional[float] = None
    last_failure_time: Optional[float] = None
    _lock: asyncio.Lock = field(default_factory=asyncio.Lock, repr=False, compare=False)

    def __post_init__(self):
        if self.original_weight == 0:
            self.original_weight = self.weight
        self.effective_weight = self.weight

    def can_handle_request(self) -> bool:
        return (
            self.is_healthy
            and not self.is_pending
            and not self.is_marked_for_deletion
            and self.weight > 0
        )

    async def acquire(self):
        async with self._lock:
            self.pending_requests += 1

    async def release(self):
        async with self._lock:
            if self.pending_requests > 0:
                self.pending_requests -= 1


@dataclass
class ServiceGroup:
    name: str
    health_policy: HealthCheckPolicy = field(default_factory=HealthCheckPolicy)
    nodes: dict = field(default_factory=dict)
    _nodes_lock: asyncio.Lock = field(default_factory=asyncio.Lock, repr=False, compare=False)

    async def add_node(self, url: str, weight: int) -> Node:
        async with self._nodes_lock:
            if url in self.nodes:
                node = self.nodes[url]
                node.weight = weight
                node.original_weight = weight
                return node
            node = Node(
                url=url,
                weight=weight,
                original_weight=weight,
                service_group=self.name,
                is_pending=True,
            )
            self.nodes[url] = node
            return node

    async def remove_node(self, url: str) -> bool:
        async with self._nodes_lock:
            if url not in self.nodes:
                return False
            node = self.nodes[url]
            node.is_marked_for_deletion = True
            node.deletion_time = time.time()
            return True

    async def get_nodes(self) -> List[Node]:
        async with self._nodes_lock:
            return list(self.nodes.values())

    async def get_available_nodes(self) -> List[Node]:
        async with self._nodes_lock:
            return [n for n in self.nodes.values() if n.can_handle_request()]

    async def is_all_unavailable(self) -> bool:
        nodes = await self.get_nodes()
        return not any(n.can_handle_request() for n in nodes)

    async def cleanup_marked_nodes(self, grace_period: float = 30.0) -> int:
        removed_count = 0
        async with self._nodes_lock:
            current_time = time.time()
            urls_to_remove = []
            for url, node in self.nodes.items():
                if (
                    node.is_marked_for_deletion
                    and node.deletion_time is not None
                    and node.pending_requests == 0
                    and current_time - node.deletion_time >= grace_period
                ):
                    urls_to_remove.append(url)
            for url in urls_to_remove:
                del self.nodes[url]
                removed_count += 1
        return removed_count
