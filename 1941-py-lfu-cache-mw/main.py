import os
from typing import Optional, List
from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field

from lfu_cache import LFUCache

app = FastAPI(title="LFU Cache Middleware", description="A lightweight LFU cache with hot-key detection and automatic TTL extension", version="1.0.0")

MAX_SIZE = int(os.environ.get("CACHE_MAX_SIZE", "1000"))
DEFAULT_TTL = int(os.environ.get("CACHE_DEFAULT_TTL", "300"))
HOT_TTL = int(os.environ.get("CACHE_HOT_TTL", "1800"))
HOT_KEY_COUNT = int(os.environ.get("CACHE_HOT_KEY_COUNT", "20"))
HOT_KEY_REFRESH_INTERVAL = int(os.environ.get("CACHE_HOT_KEY_REFRESH_INTERVAL", "60"))

cache = LFUCache(
    max_size=MAX_SIZE,
    default_ttl=DEFAULT_TTL,
    hot_ttl=HOT_TTL,
    hot_key_count=HOT_KEY_COUNT,
    hot_key_refresh_interval=HOT_KEY_REFRESH_INTERVAL,
)


class CacheEntry(BaseModel):
    value: str = Field(..., description="The value to cache")
    ttl: Optional[int] = Field(None, description="Optional TTL in seconds. Defaults to the cache's default TTL if not provided")


class WarmupItem(BaseModel):
    key: str = Field(..., description="Cache key")
    value: str = Field(..., description="Cache value")
    ttl: Optional[int] = Field(None, description="Optional TTL in seconds")


class WarmupRequest(BaseModel):
    items: List[WarmupItem] = Field(..., min_length=1, description="List of key-value pairs to warm up")


@app.put("/cache/{key}")
async def put_cache(key: str, entry: CacheEntry):
    cache.put(key, entry.value, entry.ttl)
    return JSONResponse(content={"success": True, "key": key}, status_code=200)


@app.get("/cache/{key}")
async def get_cache(key: str):
    value = cache.get(key)
    if value is None:
        raise HTTPException(status_code=404, detail="Key not found")
    return {"key": key, "value": value}


@app.delete("/cache/{key}")
async def delete_cache(key: str):
    success = cache.delete(key)
    if not success:
        raise HTTPException(status_code=404, detail="Key not found")
    return {"success": True, "key": key}


@app.post("/cache/warmup")
async def warmup_cache(request: WarmupRequest):
    items = [(item.key, item.value, item.ttl) for item in request.items]
    success = cache.batch_put(items)
    if success:
        return {"success": True, "count": len(items)}
    else:
        raise HTTPException(status_code=500, detail="Warmup failed. No changes were applied.")


@app.get("/cache/hotkeys")
async def get_hotkeys():
    hot_keys = cache.get_hot_keys()
    return {"hot_keys": hot_keys}


@app.get("/cache/stats")
async def get_stats():
    stats = cache.get_stats()
    return {
        "total_hits": stats.total_hits,
        "total_misses": stats.total_misses,
        "total_items": stats.total_items,
        "min_frequency": stats.min_frequency,
        "hit_rate": round(stats.hit_rate, 4),
    }


@app.get("/health")
async def health_check():
    return {"status": "healthy"}
