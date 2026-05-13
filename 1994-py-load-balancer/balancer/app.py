import asyncio
import os
import json
import logging
from typing import Dict
from aiohttp import web, ClientSession
from .models import ServiceGroup, HealthCheckPolicy
from .weighted_round_robin import WeightedRoundRobin
from .health_checker import HealthChecker
from .proxy import ProxyHandler

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


class LoadBalancerApp:
    CLEANUP_INTERVAL = 10.0
    DELETION_GRACE_PERIOD = 30.0

    def __init__(self):
        self._service_groups: Dict[str, ServiceGroup] = {}
        self._service_group_lock = asyncio.Lock()
        self._wrr = WeightedRoundRobin()
        self._health_checker = HealthChecker()
        self._proxy = ProxyHandler(self._wrr)
        self._client_session: ClientSession = None
        self._cleanup_task: asyncio.Task = None

    async def _ensure_service_group(self, name: str) -> ServiceGroup:
        async with self._service_group_lock:
            if name not in self._service_groups:
                policy = HealthCheckPolicy()
                group = ServiceGroup(name=name, health_policy=policy)
                self._service_groups[name] = group
                self._health_checker.register_service(group)
                logger.info(f"Created new service group: {name}")
            return self._service_groups[name]

    async def get_service_group(self, name: str) -> ServiceGroup:
        async with self._service_group_lock:
            return self._service_groups.get(name)

    async def start(self):
        self._client_session = ClientSession()
        await self._health_checker.start(self._client_session)
        await self._proxy.start(self._client_session)
        self._cleanup_task = asyncio.create_task(self._cleanup_loop())

    async def stop(self):
        if self._cleanup_task:
            self._cleanup_task.cancel()
            try:
                await self._cleanup_task
            except asyncio.CancelledError:
                pass
        
        await self._health_checker.stop()
        
        if self._client_session:
            await self._client_session.close()

    async def _cleanup_loop(self):
        try:
            while True:
                await asyncio.sleep(self.CLEANUP_INTERVAL)
                async with self._service_group_lock:
                    for group in self._service_groups.values():
                        removed = await group.cleanup_marked_nodes(self.DELETION_GRACE_PERIOD)
                        if removed > 0:
                            logger.info(f"Cleaned up {removed} nodes from {group.name}")
        except asyncio.CancelledError:
            pass

    def create_app(self) -> web.Application:
        app = web.Application()
        app.router.add_get("/stats", self._handle_stats)
        app.router.add_post("/services/{service_name}/nodes", self._handle_add_node)
        app.router.add_delete("/services/{service_name}/nodes", self._handle_remove_node)
        app.router.add_route("*", "/services/{service_name}/{path:.*}", self._handle_proxy)
        app.router.add_route("*", "/{service_name}/{path:.*}", self._handle_proxy)
        
        app.on_startup.append(self._on_startup)
        app.on_cleanup.append(self._on_cleanup)
        
        return app

    async def _on_startup(self, app):
        await self.start()

    async def _on_cleanup(self, app):
        await self.stop()

    async def _handle_stats(self, request: web.Request) -> web.Response:
        stats = {}
        async with self._service_group_lock:
            for name, group in self._service_groups.items():
                nodes = await group.get_nodes()
                all_unavailable = await group.is_all_unavailable()
                
                stats[name] = {
                    "status": "全量不可用" if all_unavailable else "正常",
                    "health_policy": {
                        "path": group.health_policy.path,
                        "interval": group.health_policy.interval,
                        "timeout": group.health_policy.timeout,
                    },
                    "nodes": [],
                }
                
                for node in nodes:
                    stats[name]["nodes"].append({
                        "url": node.url,
                        "weight": node.weight,
                        "original_weight": node.original_weight,
                        "is_healthy": node.is_healthy,
                        "is_pending": node.is_pending,
                        "is_marked_for_deletion": node.is_marked_for_deletion,
                        "pending_requests": node.pending_requests,
                        "health_failures": node.health_failures,
                        "health_successes": node.health_successes,
                        "stats": {
                            "total_requests": node.stats.total_requests,
                            "failed_requests": node.stats.failed_requests,
                            "proxy_errors": node.stats.proxy_errors,
                            "backend_errors": node.stats.backend_errors,
                            "avg_response_time_ms": round(node.stats.avg_response_time, 2),
                            "last_health_status": node.stats.last_health_status,
                            "last_health_time": node.stats.last_health_time,
                        },
                    })
        
        return web.Response(
            body=json.dumps(stats, indent=2, ensure_ascii=False),
            content_type="application/json",
        )

    async def _handle_add_node(self, request: web.Request) -> web.Response:
        service_name = request.match_info["service_name"]
        
        try:
            data = await request.json()
        except json.JSONDecodeError:
            return web.Response(status=400, text="Invalid JSON body")
        
        url = data.get("url")
        weight = data.get("weight", 1)
        
        if not url:
            return web.Response(status=400, text="Missing 'url' field")
        
        if not isinstance(weight, int) or weight < 0:
            return web.Response(status=400, text="'weight' must be a non-negative integer")
        
        group = await self._ensure_service_group(service_name)
        node = await group.add_node(url, weight)
        
        logger.info(f"Added node {url} (weight={weight}) to {service_name}")
        
        return web.Response(
            status=201,
            body=json.dumps({
                "service": service_name,
                "url": node.url,
                "weight": node.weight,
                "is_pending": node.is_pending,
            }),
            content_type="application/json",
        )

    async def _handle_remove_node(self, request: web.Request) -> web.Response:
        service_name = request.match_info["service_name"]
        
        try:
            data = await request.json()
        except json.JSONDecodeError:
            return web.Response(status=400, text="Invalid JSON body")
        
        url = data.get("url")
        if not url:
            return web.Response(status=400, text="Missing 'url' field")
        
        group = await self.get_service_group(service_name)
        if not group:
            return web.Response(status=404, text=f"Service group '{service_name}' not found")
        
        success = await group.remove_node(url)
        if not success:
            return web.Response(status=404, text=f"Node '{url}' not found in {service_name}")
        
        logger.info(f"Marked node {url} for deletion in {service_name}")
        
        return web.Response(
            status=202,
            body=json.dumps({
                "message": "Node marked for deletion",
                "grace_period_seconds": self.DELETION_GRACE_PERIOD,
            }),
            content_type="application/json",
        )

    async def _handle_proxy(self, request: web.Request) -> web.Response:
        service_name = request.match_info["service_name"]
        path = request.match_info.get("path", "")
        
        if not path:
            path = "/"
        
        group = await self.get_service_group(service_name)
        if not group:
            return web.Response(
                status=404,
                text=f"Service group '{service_name}' not found",
            )
        
        return await self._proxy.handle_request(request, group, path)


def main():
    port = int(os.environ.get("PORT", "8080"))
    
    lb_app = LoadBalancerApp()
    app = lb_app.create_app()
    
    logger.info(f"Starting load balancer on port {port}")
    web.run_app(app, host="0.0.0.0", port=port)


if __name__ == "__main__":
    main()
