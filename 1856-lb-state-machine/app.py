import os
import time
import json
import uuid
import asyncio
import aiohttp
from aiohttp import web
from typing import Dict, List, Optional, Any
from datetime import datetime
from dataclasses import dataclass, field
from enum import Enum
from collections import deque


class NodeState(str, Enum):
    NEW = "新建"
    WARMING = "预热"
    ACTIVE = "活跃"
    SUSPECTED = "疑似故障"
    REMOVED = "已摘除"


@dataclass
class Node:
    id: str
    address: str
    weight: int
    health_check_path: str
    state: NodeState = NodeState.NEW
    consecutive_success: int = 0
    consecutive_failure: int = 0
    warm_start_time: Optional[float] = None
    history: List[Dict[str, Any]] = field(default_factory=list)
    created_at: float = field(default_factory=time.time)

    def add_history(self, old_state: str, new_state: str):
        self.history.append({
            "old_state": old_state,
            "new_state": new_state,
            "timestamp": datetime.now().isoformat()
        })

    def get_effective_weight(self) -> int:
        if self.state == NodeState.WARMING:
            return max(1, self.weight // 2)
        return self.weight


class StateMachine:
    WARMING_DURATION = 30
    NEW_TO_WARMING_SUCCESS = 3
    ACTIVE_TO_SUSPECTED_FAILURE = 2
    SUSPECTED_TO_ACTIVE_SUCCESS = 3

    def __init__(self, callback_notifier: "CallbackNotifier"):
        self.nodes: Dict[str, Node] = {}
        self.callback_notifier = callback_notifier
        self.lock = asyncio.Lock()

    async def register_node(self, address: str, weight: int, health_check_path: str) -> str:
        async with self.lock:
            node_id = str(uuid.uuid4())
            node = Node(
                id=node_id,
                address=address,
                weight=weight,
                health_check_path=health_check_path,
            )
            node.add_history(None, NodeState.NEW.value)
            self.nodes[node_id] = node
            await self.callback_notifier.notify(
                node_id,
                None,
                NodeState.NEW.value,
                datetime.now().isoformat()
            )
            return node_id

    async def remove_node(self, node_id: str) -> bool:
        async with self.lock:
            if node_id not in self.nodes:
                return False
            node = self.nodes[node_id]
            if node.state == NodeState.REMOVED:
                return False
            old_state = node.state.value
            node.state = NodeState.REMOVED
            node.add_history(old_state, NodeState.REMOVED.value)
            await self.callback_notifier.notify(
                node_id,
                old_state,
                NodeState.REMOVED.value,
                datetime.now().isoformat()
            )
            return True

    def get_all_nodes(self) -> List[Dict[str, Any]]:
        return [
            {
                "id": node.id,
                "address": node.address,
                "weight": node.weight,
                "effective_weight": node.get_effective_weight(),
                "health_check_path": node.health_check_path,
                "state": node.state.value,
                "created_at": datetime.fromtimestamp(node.created_at).isoformat()
            }
            for node in self.nodes.values()
        ]

    def get_node_history(self, node_id: str) -> Optional[List[Dict[str, Any]]]:
        node = self.nodes.get(node_id)
        if not node:
            return None
        return node.history

    async def process_health_check_result(self, node_id: str, success: bool):
        async with self.lock:
            node = self.nodes.get(node_id)
            if not node or node.state == NodeState.REMOVED:
                return

            old_state = node.state

            if success:
                node.consecutive_success += 1
                node.consecutive_failure = 0
            else:
                node.consecutive_failure += 1
                node.consecutive_success = 0

            new_state = await self._determine_next_state(node, success)

            if new_state and new_state != old_state:
                node.state = new_state
                if new_state == NodeState.WARMING:
                    node.warm_start_time = time.time()
                node.consecutive_success = 0
                node.consecutive_failure = 0
                node.add_history(old_state.value, new_state.value)
                await self.callback_notifier.notify(
                    node_id,
                    old_state.value,
                    new_state.value,
                    datetime.now().isoformat()
                )

    async def _determine_next_state(self, node: Node, success: bool) -> Optional[NodeState]:
        current_state = node.state

        if current_state == NodeState.NEW:
            if success and node.consecutive_success >= self.NEW_TO_WARMING_SUCCESS:
                return NodeState.WARMING

        elif current_state == NodeState.WARMING:
            if node.warm_start_time:
                elapsed = time.time() - node.warm_start_time
                if elapsed >= self.WARMING_DURATION:
                    return NodeState.ACTIVE

        elif current_state == NodeState.ACTIVE:
            if not success and node.consecutive_failure >= self.ACTIVE_TO_SUSPECTED_FAILURE:
                return NodeState.SUSPECTED

        elif current_state == NodeState.SUSPECTED:
            if success and node.consecutive_success >= self.SUSPECTED_TO_ACTIVE_SUCCESS:
                return NodeState.ACTIVE

        return None

    def get_available_nodes(self) -> List[Node]:
        return [
            node for node in self.nodes.values()
            if node.state in (NodeState.ACTIVE, NodeState.WARMING)
        ]


class HealthChecker:
    CHECK_INTERVAL = 5

    def __init__(self, state_machine: StateMachine):
        self.state_machine = state_machine
        self.session: Optional[aiohttp.ClientSession] = None
        self._running = False
        self._task: Optional[asyncio.Task] = None

    async def start(self):
        self.session = aiohttp.ClientSession()
        self._running = True
        self._task = asyncio.create_task(self._check_loop())

    async def stop(self):
        self._running = False
        if self._task:
            self._task.cancel()
            try:
                await self._task
            except asyncio.CancelledError:
                pass
        if self.session:
            await self.session.close()

    async def _check_loop(self):
        while self._running:
            await self._check_all_nodes()
            await asyncio.sleep(self.CHECK_INTERVAL)

    async def _check_all_nodes(self):
        nodes = list(self.state_machine.nodes.values())
        tasks = [self._check_node(node) for node in nodes]
        await asyncio.gather(*tasks, return_exceptions=True)

    async def _check_node(self, node: Node):
        if node.state == NodeState.REMOVED:
            return

        url = f"http://{node.address}{node.health_check_path}"
        success = False

        try:
            async with self.session.get(url, timeout=aiohttp.ClientTimeout(total=5)) as response:
                success = response.status < 500
        except Exception:
            success = False

        await self.state_machine.process_health_check_result(node.id, success)


class LoadBalancer:
    def __init__(self, state_machine: StateMachine):
        self.state_machine = state_machine
        self.weights: Dict[str, int] = {}

    def select_node(self) -> Optional[Node]:
        available = self.state_machine.get_available_nodes()
        if not available:
            return None

        total_weight = sum(node.get_effective_weight() for node in available)
        if total_weight == 0:
            return available[0]

        import random
        r = random.uniform(0, total_weight)
        current = 0
        for node in available:
            current += node.get_effective_weight()
            if r <= current:
                return node

        return available[-1]


class CallbackNotifier:
    def __init__(self):
        self.callbacks: List[str] = []
        self.session: Optional[aiohttp.ClientSession] = None
        self.lock = asyncio.Lock()

    async def start(self):
        self.session = aiohttp.ClientSession()

    async def stop(self):
        if self.session:
            await self.session.close()

    async def add_callback(self, url: str):
        async with self.lock:
            if url not in self.callbacks:
                self.callbacks.append(url)

    def get_callbacks(self) -> List[str]:
        return list(self.callbacks)

    async def notify(self, node_id: str, old_state: Optional[str], new_state: str, timestamp: str):
        if not self.callbacks:
            return

        payload = {
            "node_id": node_id,
            "old_state": old_state,
            "new_state": new_state,
            "timestamp": timestamp
        }

        async with self.lock:
            callbacks = list(self.callbacks)

        tasks = [self._send_callback(url, payload) for url in callbacks]
        await asyncio.gather(*tasks, return_exceptions=True)

    async def _send_callback(self, url: str, payload: dict):
        try:
            async with self.session.post(
                url,
                json=payload,
                timeout=aiohttp.ClientTimeout(total=10)
            ):
                pass
        except Exception:
            pass


class Server:
    def __init__(self):
        self.callback_notifier = CallbackNotifier()
        self.state_machine = StateMachine(self.callback_notifier)
        self.health_checker = HealthChecker(self.state_machine)
        self.load_balancer = LoadBalancer(self.state_machine)

    async def on_startup(self, app: web.Application):
        await self.callback_notifier.start()
        await self.health_checker.start()

    async def on_shutdown(self, app: web.Application):
        await self.health_checker.stop()
        await self.callback_notifier.stop()

    async def handle_register_node(self, request: web.Request) -> web.Response:
        try:
            data = await request.json()
        except Exception:
            return web.json_response({"error": "Invalid JSON"}, status=400)

        address = data.get("address")
        weight = data.get("weight")
        health_check_path = data.get("health_check_path")

        if not all([address, weight, health_check_path]):
            return web.json_response(
                {"error": "Missing required fields: address, weight, health_check_path"},
                status=400
            )

        if not isinstance(weight, int) or weight <= 0:
            return web.json_response(
                {"error": "weight must be a positive integer"},
                status=400
            )

        node_id = await self.state_machine.register_node(address, weight, health_check_path)
        return web.json_response({"id": node_id}, status=201)

    async def handle_remove_node(self, request: web.Request) -> web.Response:
        node_id = request.match_info["id"]
        success = await self.state_machine.remove_node(node_id)
        if not success:
            return web.json_response({"error": "Node not found or already removed"}, status=404)
        return web.json_response({"message": "Node removed"}, status=200)

    async def handle_get_nodes(self, request: web.Request) -> web.Response:
        nodes = self.state_machine.get_all_nodes()
        return web.json_response({"nodes": nodes}, status=200)

    async def handle_get_node_history(self, request: web.Request) -> web.Response:
        node_id = request.match_info["id"]
        history = self.state_machine.get_node_history(node_id)
        if history is None:
            return web.json_response({"error": "Node not found"}, status=404)
        return web.json_response({"node_id": node_id, "history": history}, status=200)

    async def handle_register_callback(self, request: web.Request) -> web.Response:
        try:
            data = await request.json()
        except Exception:
            return web.json_response({"error": "Invalid JSON"}, status=400)

        url = data.get("url")
        if not url:
            return web.json_response({"error": "Missing url field"}, status=400)

        await self.callback_notifier.add_callback(url)
        return web.json_response({"message": "Callback registered"}, status=201)

    async def handle_stats(self, request: web.Request) -> web.Response:
        nodes = self.state_machine.get_all_nodes()
        stats = []
        for node_info in nodes:
            node_id = node_info["id"]
            history = self.state_machine.get_node_history(node_id) or []
            stats.append({
                "current_state": node_info["state"],
                "history": history
            })
        return web.json_response({"stats": stats}, status=200)

    async def handle_pick_node(self, request: web.Request) -> web.Response:
        node = self.load_balancer.select_node()
        if not node:
            return web.json_response({"error": "No available nodes"}, status=503)
        return web.json_response({
            "id": node.id,
            "address": node.address,
            "state": node.state.value,
            "effective_weight": node.get_effective_weight()
        }, status=200)

    def create_app(self) -> web.Application:
        app = web.Application()
        app.router.add_post("/nodes", self.handle_register_node)
        app.router.add_delete("/nodes/{id}", self.handle_remove_node)
        app.router.add_get("/nodes", self.handle_get_nodes)
        app.router.add_get("/nodes/{id}/history", self.handle_get_node_history)
        app.router.add_post("/callbacks", self.handle_register_callback)
        app.router.add_get("/stats", self.handle_stats)
        app.router.add_get("/pick", self.handle_pick_node)

        app.on_startup.append(self.on_startup)
        app.on_shutdown.append(self.on_shutdown)

        return app


def main():
    port = int(os.environ.get("PORT", 8080))
    server = Server()
    app = server.create_app()
    web.run_app(app, port=port)


if __name__ == "__main__":
    main()
