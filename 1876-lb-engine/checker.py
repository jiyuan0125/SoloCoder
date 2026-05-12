import asyncio
import httpx
from typing import Callable, Optional
from models import Node, NodeStatus
from store import store


class HealthChecker:
    def __init__(
        self,
        interval: int = 10,
        failure_threshold: int = 3,
        on_status_change: Optional[Callable[[Node, NodeStatus], None]] = None
    ):
        self.interval = interval
        self.failure_threshold = failure_threshold
        self.on_status_change = on_status_change
        self._task: Optional[asyncio.Task] = None
        self._running = False

    async def _check_node_health(self, node: Node) -> bool:
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                url = f"{node.address}/health" if not node.address.endswith("/health") else node.address
                if not url.startswith("http"):
                    url = f"http://{url}"
                response = await client.get(url)
                return 200 <= response.status_code < 400
        except Exception:
            return False

    def _transition_state(self, node: Node, is_healthy: bool):
        old_status = node.status

        if node.status == NodeStatus.NEW:
            if is_healthy:
                node.status = NodeStatus.ACTIVE
                node.consecutive_success_count = 1
                node.consecutive_fail_count = 0
            else:
                node.consecutive_fail_count += 1
                if node.consecutive_fail_count >= self.failure_threshold:
                    node.status = NodeStatus.REMOVED

        elif node.status == NodeStatus.ACTIVE:
            if is_healthy:
                node.consecutive_success_count += 1
                node.consecutive_fail_count = 0
            else:
                node.consecutive_fail_count += 1
                node.consecutive_success_count = 0
                node.status = NodeStatus.SUSPECTED

        elif node.status == NodeStatus.SUSPECTED:
            if is_healthy:
                node.consecutive_success_count += 1
                node.consecutive_fail_count = 0
                if node.consecutive_success_count >= 3:
                    node.status = NodeStatus.ACTIVE
                    node.consecutive_success_count = 0
            else:
                node.consecutive_fail_count += 1
                node.consecutive_success_count = 0
                if node.consecutive_fail_count >= self.failure_threshold:
                    node.status = NodeStatus.REMOVED

        elif node.status == NodeStatus.REMOVED:
            pass

        if old_status != node.status and self.on_status_change:
            self.on_status_change(node, old_status)

    async def _check_all_nodes(self):
        nodes = store.get_all_nodes()
        for node in nodes:
            if node.status == NodeStatus.REMOVED:
                continue
            is_healthy = await self._check_node_health(node)
            self._transition_state(node, is_healthy)
            store.update_node(
                node.id,
                status=node.status,
                consecutive_success_count=node.consecutive_success_count,
                consecutive_fail_count=node.consecutive_fail_count
            )

    async def _run(self):
        while self._running:
            await self._check_all_nodes()
            await asyncio.sleep(self.interval)

    def start(self):
        if not self._running:
            self._running = True
            self._task = asyncio.create_task(self._run())

    def stop(self):
        self._running = False
        if self._task:
            self._task.cancel()
