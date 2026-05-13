import os
from typing import Optional, List
from datetime import datetime, timezone
from fastapi import FastAPI, HTTPException, Query
from pydantic import BaseModel, Field

from cache_manager import CacheManager, InvalidationReason


app = FastAPI(title="缓存失效管理器")
cache_manager = CacheManager(max_logs=5000)


class RegisterRequest(BaseModel):
    key: str
    ttl: int = Field(gt=0, description="TTL 秒数")


class InvalidateRequest(BaseModel):
    key: str
    operator: Optional[str] = None


class DependencyRequest(BaseModel):
    parent_key: str
    child_key: str


class CacheStatusResponse(BaseModel):
    key: str
    exists: bool
    is_invalidated: bool
    ttl: Optional[int]
    dependencies: List[str]


class InvalidationLogResponse(BaseModel):
    id: int
    keys: List[str]
    reason: str
    operator: Optional[str]
    timestamp: str


def format_timestamp(ts: float) -> str:
    return datetime.fromtimestamp(ts, tz=timezone.utc).isoformat()


@app.post("/cache/register")
def register_cache(req: RegisterRequest):
    entry = cache_manager.register(req.key, req.ttl)
    return {
        "key": entry.key,
        "ttl": entry.ttl,
        "created_at": format_timestamp(entry.created_at),
        "is_invalidated": entry.is_invalidated
    }


@app.post("/cache/invalidate")
def invalidate_cache(req: InvalidateRequest):
    invalidated = cache_manager.invalidate(req.key, req.operator)
    return {
        "invalidated_keys": invalidated,
        "count": len(invalidated)
    }


@app.delete("/cache/prefix/{prefix}")
def invalidate_by_prefix(prefix: str, operator: Optional[str] = None):
    invalidated = cache_manager.invalidate_by_prefix(prefix, operator)
    return {
        "invalidated_keys": invalidated,
        "count": len(invalidated)
    }


@app.get("/cache/{key}", response_model=CacheStatusResponse)
def get_cache_status(key: str):
    entry = cache_manager.get(key)
    if entry is None:
        return CacheStatusResponse(
            key=key,
            exists=False,
            is_invalidated=True,
            ttl=None,
            dependencies=[]
        )
    return CacheStatusResponse(
        key=entry.key,
        exists=True,
        is_invalidated=entry.is_invalidated,
        ttl=entry.ttl,
        dependencies=cache_manager.get_dependencies(key)
    )


@app.post("/dependencies")
def register_dependency(req: DependencyRequest):
    cache_manager.register_dependency(req.parent_key, req.child_key)
    return {
        "parent_key": req.parent_key,
        "child_key": req.child_key,
        "children": cache_manager.get_children(req.parent_key)
    }


@app.get("/invalidation-logs")
def get_invalidation_logs(
    start_time: Optional[float] = Query(None, description="开始时间戳"),
    end_time: Optional[float] = Query(None, description="结束时间戳"),
    reason: Optional[InvalidationReason] = Query(None, description="触发类型")
):
    logs = cache_manager.query_logs(start_time, end_time, reason)
    return [
        InvalidationLogResponse(
            id=log.id,
            keys=log.keys,
            reason=log.reason.value,
            operator=log.operator,
            timestamp=format_timestamp(log.timestamp)
        )
        for log in logs
    ]


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run(app, host="0.0.0.0", port=port)
