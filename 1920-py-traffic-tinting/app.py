import os
from contextlib import asynccontextmanager
from typing import Dict, List, Optional

from fastapi import FastAPI, HTTPException, Request
from pydantic import BaseModel, Field

from traffic_router import BackendRegistry, GroupAllocator


class RatioConfig(BaseModel):
    groups: Dict[str, int] = Field(..., description="分组比例，如 {'A': 70, 'B': 30}")


class BackendRegistration(BaseModel):
    group: str = Field(..., description="分组名称 (A 或 B)")
    address: str = Field(..., description="后端地址，如 http://127.0.0.1:8081")


class BackendStatus(BaseModel):
    address: str
    healthy: bool


class StatusResponse(BaseModel):
    ratios: Dict[str, int]
    backends: Dict[str, List[BackendStatus]]
    request_counts: Dict[str, int]


group_allocator: Optional[GroupAllocator] = None
backend_registry: Optional[BackendRegistry] = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global group_allocator, backend_registry
    group_allocator = GroupAllocator({"A": 50, "B": 50})
    backend_registry = BackendRegistry()
    backend_registry.start_health_checker()
    try:
        yield
    finally:
        if backend_registry is not None:
            backend_registry.stop_health_checker()


app = FastAPI(lifespan=lifespan)


@app.put("/config/ratio")
async def update_ratio(config: RatioConfig):
    total = sum(config.groups.values())
    if total != 100:
        raise HTTPException(status_code=400, detail=f"分组比例之和必须为 100，当前为 {total}")
    if not config.groups:
        raise HTTPException(status_code=400, detail="分组不能为空")
    for group, ratio in config.groups.items():
        if ratio < 0:
            raise HTTPException(status_code=400, detail=f"分组 '{group}' 的比例不能为负数")
    group_allocator.update_ratios(config.groups)
    return {"ratios": group_allocator.get_ratios()}


@app.get("/config/ratio")
async def get_ratio():
    return {"ratios": group_allocator.get_ratios()}


@app.post("/backends")
async def register_backend(registration: BackendRegistration):
    if not registration.address.startswith(("http://", "https://")):
        raise HTTPException(status_code=400, detail="地址必须以 http:// 或 https:// 开头")
    backend_registry.register_backend(registration.group, registration.address)
    return {
        "group": registration.group,
        "address": registration.address,
        "status": "registered"
    }


@app.get("/status")
async def get_status():
    ratios = group_allocator.get_ratios()
    all_backends = backend_registry.get_all_backends()
    backends_status = {}
    for group, backends in all_backends.items():
        backends_status[group] = [
            BackendStatus(address=b.address, healthy=b.healthy)
            for b in backends
        ]
    for group in ratios:
        if group not in backends_status:
            backends_status[group] = []
    request_counts = backend_registry.get_request_counts()
    return StatusResponse(
        ratios=ratios,
        backends=backends_status,
        request_counts=request_counts
    )


@app.get("/backends/{group}")
async def get_group_backends(group: str):
    backends = backend_registry.get_group_backends(group)
    return {
        "group": group,
        "backends": [
            {"address": b.address, "healthy": b.healthy}
            for b in backends
        ]
    }


@app.get("/route")
async def route_request(request: Request):
    user_id = request.query_params.get("user_id")
    if not user_id:
        raise HTTPException(status_code=400, detail="缺少 user_id 参数")
    group = group_allocator.get_group(user_id)
    if group is None:
        raise HTTPException(status_code=503, detail="没有可用的分组")
    backend = backend_registry.pick_backend(group)
    if backend is None:
        raise HTTPException(status_code=503, detail=f"分组 '{group}' 没有可用的健康后端")
    backend_registry.record_request(group)
    return {
        "user_id": user_id,
        "group": group,
        "backend": backend.address
    }
