import os
import time
import httpx
from fastapi import FastAPI, HTTPException, Request
from contextlib import asynccontextmanager
from typing import List, Dict

from models import (
    Node, NodeStatus, NodeStats, NodeRegistrationRequest,
    NodeUpdateRequest, CallbackRegistrationRequest
)
from store import store
from checker import HealthChecker
from balancer import balancer
from callback import notifier

PORT = int(os.environ.get("PORT", 8000))
HEALTH_CHECK_INTERVAL = int(os.environ.get("HEALTH_CHECK_INTERVAL", 10))
HEALTH_CHECK_FAILURE_THRESHOLD = int(os.environ.get("HEALTH_CHECK_FAILURE_THRESHOLD", 3))

checker: HealthChecker | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global checker
    checker = HealthChecker(
        interval=HEALTH_CHECK_INTERVAL,
        failure_threshold=HEALTH_CHECK_FAILURE_THRESHOLD,
        on_status_change=notifier.notify_status_change
    )
    checker.start()
    yield
    if checker:
        checker.stop()


app = FastAPI(title="Load Balancer Engine", lifespan=lifespan)


@app.post("/nodes", response_model=Node)
async def register_node(request: NodeRegistrationRequest):
    node = store.add_node(address=request.address, weight=request.weight)
    return node


@app.get("/nodes", response_model=List[Node])
async def get_all_nodes():
    return store.get_all_nodes()


@app.get("/nodes/{node_id}", response_model=Node)
async def get_node(node_id: str):
    node = store.get_node(node_id)
    if not node:
        raise HTTPException(status_code=404, detail="Node not found")
    return node


@app.patch("/nodes/{node_id}", response_model=Node)
async def update_node(node_id: str, request: NodeUpdateRequest):
    node = store.get_node(node_id)
    if not node:
        raise HTTPException(status_code=404, detail="Node not found")

    updates = {}
    if request.weight is not None:
        updates["weight"] = request.weight
    if request.address is not None:
        updates["address"] = request.address

    if updates:
        updated_node = store.update_node(node_id, **updates)
        return updated_node
    return node


@app.delete("/nodes/{node_id}")
async def delete_node(node_id: str):
    if not store.delete_node(node_id):
        raise HTTPException(status_code=404, detail="Node not found")
    return {"message": "Node deleted successfully"}


@app.get("/nodes/{node_id}/stats", response_model=NodeStats)
async def get_node_stats(node_id: str):
    stats = store.get_stats(node_id)
    if not stats:
        node = store.get_node(node_id)
        if not node:
            raise HTTPException(status_code=404, detail="Node not found")
        stats = NodeStats(node_id=node_id)
    return stats


@app.get("/stats", response_model=Dict[str, List[NodeStats]])
async def get_overall_stats():
    return {"stats": store.get_all_stats()}


@app.post("/callbacks")
async def register_callback(request: CallbackRegistrationRequest):
    callback = store.add_callback(url=request.url)
    return {"message": "Callback registered successfully", "url": callback.url}


@app.api_route("/proxy/{path:path}", methods=["GET", "POST", "PUT", "DELETE", "PATCH"])
async def proxy_request(path: str, request: Request):
    node = balancer.select_node()
    if not node:
        raise HTTPException(status_code=503, detail="No available nodes")

    start_time = time.time()
    address = node.address
    if not address.startswith("http"):
        address = f"http://{address}"
    if not path.startswith("/"):
        path = f"/{path}"
    target_url = f"{address.rstrip('/')}/{path.lstrip('/')}"

    try:
        async with httpx.AsyncClient(timeout=30.0) as client:
            body = await request.body()
            response = await client.request(
                method=request.method,
                url=target_url,
                params=dict(request.query_params),
                headers=dict(request.headers),
                content=body
            )
            elapsed_ms = int((time.time() - start_time) * 1000)
            store.record_request(node.id, elapsed_ms)

            return response.content
    except Exception as e:
        elapsed_ms = int((time.time() - start_time) * 1000)
        store.record_request(node.id, elapsed_ms)
        raise HTTPException(status_code=502, detail=f"Backend error: {str(e)}")


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=PORT, reload=True)
