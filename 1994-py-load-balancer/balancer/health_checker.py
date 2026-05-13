import asyncio
import time
import logging
from typing import Dict, Optional
from aiohttp import ClientSession, ClientTimeout, ClientError
from .models import ServiceGroup, Node, HealthCheckPolicy

logger = logging.getLogger(__name__)


class HealthChecker:
    def __init__(self):
        self._tasks: Dict[str, asyncio.Task] = {}
        self._session: Optional[ClientSession] = None
        self._stop_event = asyncio.Event()

    async def start(self, session: ClientSession):
        self._session = session
        self._stop_event.clear()

    async def stop(self):
        self._stop_event.set()
        for task in self._tasks.values():
            task.cancel()
        await asyncio.gather(*self._tasks.values(), return_exceptions=True)
        self._tasks.clear()

    def register_service(self, service_group: ServiceGroup):
        if service_group.name in self._tasks:
            return
        task = asyncio.create_task(self._health_check_loop(service_group))
        self._tasks[service_group.name] = task

    def unregister_service(self, service_name: str):
        if service_name in self._tasks:
            self._tasks[service_name].cancel()
            del self._tasks[service_name]

    async def _health_check_loop(self, service_group: ServiceGroup):
        policy = service_group.health_policy
        try:
            while not self._stop_event.is_set():
                nodes = await service_group.get_nodes()
                for node in nodes:
                    if node.is_marked_for_deletion:
                        continue
                    await self._check_node(node, policy, service_group)
                
                await asyncio.sleep(policy.interval)
        except asyncio.CancelledError:
            pass
        except Exception as e:
            logger.error(f"Health check loop error for {service_group.name}: {e}")

    async def _check_node(
        self, node: Node, policy: HealthCheckPolicy, service_group: ServiceGroup
    ):
        if self._session is None:
            return

        health_url = f"{node.url.rstrip('/')}{policy.path}"
        timeout = ClientTimeout(total=policy.timeout)

        try:
            start_time = time.time()
            async with self._session.get(health_url, timeout=timeout) as response:
                elapsed = time.time() - start_time
                is_success = 200 <= response.status < 500

                node.stats.last_health_time = time.time()

                if is_success:
                    node.health_successes += 1
                    node.health_failures = 0
                    node.stats.last_health_status = "success"

                    if node.is_pending:
                        node.is_pending = False
                        node.is_healthy = True
                        node.weight = node.original_weight
                        node.effective_weight = node.original_weight
                        logger.info(f"Node {node.url} passed initial health check, ready for traffic")

                    if not node.is_healthy and node.health_successes >= policy.consecutive_successes_to_up:
                        node.is_healthy = True
                        node.weight = node.original_weight
                        node.effective_weight = node.original_weight
                        logger.info(f"Node {node.url} recovered to healthy")
                else:
                    await self._handle_health_failure(node, policy, f"HTTP {response.status}")

        except asyncio.TimeoutError:
            await self._handle_health_failure(node, policy, "timeout")
        except ClientError as e:
            await self._handle_health_failure(node, policy, f"client_error: {str(e)}")
        except Exception as e:
            logger.error(f"Unexpected health check error for {node.url}: {e}")
            await self._handle_health_failure(node, policy, f"error: {str(e)}")

    async def _handle_health_failure(self, node: Node, policy, reason: str):
        node.health_failures += 1
        node.health_successes = 0
        node.stats.last_health_time = time.time()
        node.stats.last_health_status = f"failed: {reason}"

        if node.is_pending:
            return

        if node.is_healthy and node.health_failures >= policy.consecutive_failures_to_down:
            node.is_healthy = False
            node.weight = 0
            node.effective_weight = 0
            logger.warning(f"Node {node.url} marked as unhealthy: {reason}")
