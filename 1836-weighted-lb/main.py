import asyncio
import os
import random
import time
import uuid
from contextlib import asynccontextmanager
from datetime import datetime
from enum import Enum
from typing import Any, Dict, List, Optional

import httpx
from fastapi import FastAPI, HTTPException, Request, Response
from pydantic import BaseModel, Field


class HealthStatus(str, Enum):
    HEALTHY = "healthy"
    UNHEALTHY = "unhealthy"


class Node(BaseModel):
    id: str
    address: str
    weight: int = Field(..., gt=0)
    status: HealthStatus = HealthStatus.HEALTHY
    created_at: datetime = Field(default_factory=datetime.utcnow)


class NodeStatistics(BaseModel):
    node_id: str
    total_requests: int = 0
    successful_requests: int = 0
    failed_requests: int = 0
    total_response_time_ms: float = 0.0
    first_seen: datetime = Field(default_factory=datetime.utcnow)
    last_seen: Optional[datetime] = None


class NodeStatsResponse(BaseModel):
    node_id: str
    address: Optional[str]
    total_requests: int
    success_rate: float
    avg_response_time_ms: float
    first_seen: datetime
    last_seen: Optional[datetime]


class NodeStatusResponse(BaseModel):
    id: str
    address: str
    weight: int
    status: HealthStatus


class Callback(BaseModel):
    id: str
    url: str
    node_ids: List[str]
    created_at: datetime = Field(default_factory=datetime.utcnow)


class CreateNodeRequest(BaseModel):
    address: str
    weight: int = Field(..., gt=0)


class UpdateNodeWeightRequest(BaseModel):
    weight: int = Field(..., gt=0)


class RegisterCallbackRequest(BaseModel):
    url: str
    node_ids: List[str]


class Storage:
    def __init__(self) -> None:
        self.nodes: Dict[str, Node] = {}
        self.callbacks: Dict[str, Callback] = {}
        self.statistics: Dict[str, NodeStatistics] = {}
        self.node_id_by_address: Dict[str, str] = {}

    def add_node(self, address: str, weight: int) -> Node:
        if address in self.node_id_by_address:
            raise HTTPException(status_code=400, detail=f"Node with address {address} already exists")
        node_id = str(uuid.uuid4())
        node = Node(id=node_id, address=address, weight=weight)
        self.nodes[node_id] = node
        self.node_id_by_address[address] = node_id
        if node_id not in self.statistics:
            self.statistics[node_id] = NodeStatistics(node_id=node_id)
        return node

    def update_node_weight(self, node_id: str, weight: int) -> Node:
        if node_id not in self.nodes:
            raise HTTPException(status_code=404, detail=f"Node {node_id} not found")
        self.nodes[node_id].weight = weight
        return self.nodes[node_id]

    def delete_node(self, node_id: str) -> None:
        if node_id not in self.nodes:
            raise HTTPException(status_code=404, detail=f"Node {node_id} not found")
        node = self.nodes[node_id]
        del self.node_id_by_address[node.address]
        del self.nodes[node_id]

    def register_callback(self, url: str, node_ids: List[str]) -> Callback:
        callback_id = str(uuid.uuid4())
        callback = Callback(id=callback_id, url=url, node_ids=node_ids)
        self.callbacks[callback_id] = callback
        return callback

    def get_callbacks_for_node(self, node_id: str) -> List[Callback]:
        return [cb for cb in self.callbacks.values() if node_id in cb.node_ids]

    def get_healthy_nodes(self) -> List[Node]:
        return [node for node in self.nodes.values() if node.status == HealthStatus.HEALTHY]

    def update_node_status(self, node_id: str, status: HealthStatus) -> Optional[HealthStatus]:
        if node_id not in self.nodes:
            return None
        old_status = self.nodes[node_id].status
        self.nodes[node_id].status = status
        return old_status

    def record_request(self, node_id: str, success: bool, response_time_ms: float) -> None:
        if node_id not in self.statistics:
            self.statistics[node_id] = NodeStatistics(node_id=node_id)
        stats = self.statistics[node_id]
        stats.total_requests += 1
        if success:
            stats.successful_requests += 1
        else:
            stats.failed_requests += 1
        stats.total_response_time_ms += response_time_ms
        stats.last_seen = datetime.utcnow()


class WeightedLoadBalancer:
    def __init__(self, storage: Storage) -> None:
        self.storage = storage

    def select_node(self) -> Optional[Node]:
        healthy_nodes = self.storage.get_healthy_nodes()
        if not healthy_nodes:
            return None
        total_weight = sum(node.weight for node in healthy_nodes)
        if total_weight == 0:
            return None
        random_value = random.randint(1, total_weight)
        current_weight = 0
        for node in healthy_nodes:
            current_weight += node.weight
            if current_weight >= random_value:
                return node
        return healthy_nodes[-1]


class HealthChecker:
    def __init__(self, storage: Storage, notifier: "CallbackNotifier") -> None:
        self.storage = storage
        self.notifier = notifier
        self.interval_seconds = 10
        self.task: Optional[asyncio.Task] = None

    async def check_node_health(self, node: Node) -> bool:
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                url = f"{node.address}/health" if not node.address.startswith("http") else f"{node.address}/health"
                if not url.startswith("http"):
                    url = f"http://{url}"
                response = await client.get(url)
                return 200 <= response.status_code < 300
        except Exception:
            return False

    async def run_check(self) -> None:
        node_ids = list(self.storage.nodes.keys())
        for node_id in node_ids:
            if node_id not in self.storage.nodes:
                continue
            node = self.storage.nodes[node_id]
            is_healthy = await self.check_node_health(node)
            new_status = HealthStatus.HEALTHY if is_healthy else HealthStatus.UNHEALTHY
            if node.status != new_status:
                old_status = self.storage.update_node_status(node_id, new_status)
                if old_status is not None and old_status != new_status:
                    await self.notifier.notify_status_change(node_id, old_status, new_status)

    async def start(self) -> None:
        async def loop() -> None:
            while True:
                await self.run_check()
                await asyncio.sleep(self.interval_seconds)
        self.task = asyncio.create_task(loop())

    async def stop(self) -> None:
        if self.task is not None:
            self.task.cancel()
            try:
                await self.task
            except asyncio.CancelledError:
                pass


class CallbackNotifier:
    def __init__(self, storage: Storage) -> None:
        self.storage = storage

    async def notify_status_change(self, node_id: str, old_status: HealthStatus, new_status: HealthStatus) -> None:
        callbacks = self.storage.get_callbacks_for_node(node_id)
        if not callbacks:
            return
        payload = {
            "node_id": node_id,
            "old_status": old_status.value,
            "new_status": new_status.value,
            "timestamp": datetime.utcnow().isoformat(),
        }
        async with httpx.AsyncClient(timeout=10.0) as client:
            for callback in callbacks:
                try:
                    await client.post(callback.url, json=payload)
                except Exception:
                    pass


class RequestForwarder:
    def __init__(self, storage: Storage, load_balancer: WeightedLoadBalancer) -> None:
        self.storage = storage
        self.load_balancer = load_balancer

    async def forward(self, request: Request) -> Response:
        node = self.load_balancer.select_node()
        if node is None:
            raise HTTPException(status_code=503, detail="No healthy nodes available")

        start_time = time.perf_counter()
        success = False
        response_time_ms = 0.0

        try:
            target_url = node.address if node.address.startswith("http") else f"http://{node.address}"
            path = request.url.path if request.url.path else "/"
            query = request.url.query
            if query:
                path = f"{path}?{query}"

            async with httpx.AsyncClient(timeout=60.0) as client:
                method = request.method
                headers = {k: v for k, v in request.headers.items() if k.lower() not in {"host", "content-length"}}
                body = await request.body()

                response = await client.request(
                    method=method,
                    url=f"{target_url}{path}",
                    headers=headers,
                    content=body,
                )

                success = True
                response_time_ms = (time.perf_counter() - start_time) * 1000

                return Response(
                    content=response.content,
                    status_code=response.status_code,
                    headers=dict(response.headers),
                )
        except Exception as e:
            response_time_ms = (time.perf_counter() - start_time) * 1000
            raise HTTPException(status_code=502, detail=f"Failed to forward request: {str(e)}")
        finally:
            self.storage.record_request(node.id, success, response_time_ms)


storage = Storage()
notifier = CallbackNotifier(storage)
load_balancer = WeightedLoadBalancer(storage)
health_checker = HealthChecker(storage, notifier)
forwarder = RequestForwarder(storage, load_balancer)


@asynccontextmanager
async def lifespan(app: FastAPI):
    await health_checker.start()
    yield
    await health_checker.stop()


app = FastAPI(title="Weighted Load Balancer", lifespan=lifespan)


@app.post("/nodes", response_model=Node, status_code=201)
def create_node(request: CreateNodeRequest):
    return storage.add_node(address=request.address, weight=request.weight)


@app.patch("/nodes/{node_id}", response_model=Node)
def update_node_weight(node_id: str, request: UpdateNodeWeightRequest):
    return storage.update_node_weight(node_id=node_id, weight=request.weight)


@app.delete("/nodes/{node_id}", status_code=204)
def delete_node(node_id: str):
    storage.delete_node(node_id=node_id)
    return Response(status_code=204)


@app.get("/nodes", response_model=List[NodeStatusResponse])
def list_nodes():
    return [
        NodeStatusResponse(id=n.id, address=n.address, weight=n.weight, status=n.status)
        for n in storage.nodes.values()
    ]


@app.get("/nodes/{node_id}", response_model=NodeStatusResponse)
def get_node(node_id: str):
    if node_id not in storage.nodes:
        raise HTTPException(status_code=404, detail=f"Node {node_id} not found")
    node = storage.nodes[node_id]
    return NodeStatusResponse(id=node.id, address=node.address, weight=node.weight, status=node.status)


@app.post("/callbacks", response_model=Callback, status_code=201)
def register_callback(request: RegisterCallbackRequest):
    return storage.register_callback(url=request.url, node_ids=request.node_ids)


@app.get("/stats", response_model=List[NodeStatsResponse])
def get_statistics():
    results: List[NodeStatsResponse] = []
    for node_id, stats in storage.statistics.items():
        node = storage.nodes.get(node_id)
        success_rate = (stats.successful_requests / stats.total_requests * 100) if stats.total_requests > 0 else 0.0
        avg_response_time = (stats.total_response_time_ms / stats.total_requests) if stats.total_requests > 0 else 0.0
        results.append(
            NodeStatsResponse(
                node_id=node_id,
                address=node.address if node else None,
                total_requests=stats.total_requests,
                success_rate=round(success_rate, 2),
                avg_response_time_ms=round(avg_response_time, 2),
                first_seen=stats.first_seen,
                last_seen=stats.last_seen,
            )
        )
    return results


@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"])
async def proxy_request(request: Request, path: str):
    return await forwarder.forward(request)


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
