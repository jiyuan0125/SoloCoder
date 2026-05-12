import asyncio
import time
from collections import OrderedDict
from typing import Any, Dict, List, Optional, Set
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
import uvicorn
import os


class CacheEntry:
    __slots__ = ('value', 'expire_at', 'tags')

    def __init__(self, value: Any, expire_at: Optional[float], tags: List[str]):
        self.value = value
        self.expire_at = expire_at
        self.tags = tags


class Stats:
    def __init__(self):
        self.hits = 0
        self.misses = 0
        self.evictions = 0

    @property
    def hit_rate(self) -> float:
        total = self.hits + self.misses
        return self.hits / total if total > 0 else 0.0


class CacheEngine:
    def __init__(self, max_capacity: int = 10000):
        self._data: Dict[str, CacheEntry] = {}
        self._lru_order: OrderedDict = OrderedDict()
        self._tag_index: Dict[str, Set[str]] = {}
        self._max_capacity = max_capacity
        self._stats = Stats()
        self._lock = asyncio.Lock()

    def _is_expired(self, entry: CacheEntry) -> bool:
        if entry.expire_at is None:
            return False
        return time.time() > entry.expire_at

    async def _evict_lru(self, count: int) -> int:
        evicted = 0
        keys_to_evict = list(self._lru_order.keys())[:count]
        for key in keys_to_evict:
            await self._delete_key(key)
            self._stats.evictions += 1
            evicted += 1
        return evicted

    async def _check_capacity(self):
        if len(self._data) > self._max_capacity:
            evict_count = max(1, int(self._max_capacity * 0.1))
            await self._evict_lru(evict_count)

    async def _add_tag_index(self, key: str, tags: List[str]):
        for tag in tags:
            if tag not in self._tag_index:
                self._tag_index[tag] = set()
            self._tag_index[tag].add(key)

    async def _remove_tag_index(self, key: str, tags: List[str]):
        for tag in tags:
            if tag in self._tag_index:
                self._tag_index[tag].discard(key)
                if not self._tag_index[tag]:
                    del self._tag_index[tag]

    async def _delete_key(self, key: str) -> bool:
        if key in self._data:
            entry = self._data[key]
            await self._remove_tag_index(key, entry.tags)
            del self._data[key]
            if key in self._lru_order:
                del self._lru_order[key]
            return True
        return False

    async def get(self, key: str) -> Optional[Any]:
        async with self._lock:
            if key not in self._data:
                self._stats.misses += 1
                return None

            entry = self._data[key]
            if self._is_expired(entry):
                await self._delete_key(key)
                self._stats.misses += 1
                return None

            if key in self._lru_order:
                self._lru_order.move_to_end(key)
            self._stats.hits += 1
            return entry.value

    async def set(self, key: str, value: Any, ttl: int = 0, tags: Optional[List[str]] = None) -> bool:
        async with self._lock:
            if tags is None:
                tags = []

            if key in self._data:
                old_entry = self._data[key]
                await self._remove_tag_index(key, old_entry.tags)

            expire_at = None if ttl == 0 else time.time() + ttl
            self._data[key] = CacheEntry(value, expire_at, tags)

            self._lru_order[key] = True
            self._lru_order.move_to_end(key)

            await self._add_tag_index(key, tags)
            await self._check_capacity()

            return True

    async def delete(self, key: str) -> bool:
        async with self._lock:
            return await self._delete_key(key)

    async def delete_by_tag(self, tag: str) -> int:
        async with self._lock:
            if tag not in self._tag_index:
                return 0

            keys = list(self._tag_index[tag])
            count = 0
            for key in keys:
                if await self._delete_key(key):
                    count += 1
            return count

    async def batch_get(self, keys: List[str]) -> Dict[str, Any]:
        results: Dict[str, Any] = {}
        for key in keys:
            results[key] = await self.get(key)
        return results

    async def batch_set(self, items: Dict[str, Any], ttl: int = 0, tags: Optional[List[str]] = None) -> int:
        count = 0
        for key, value in items.items():
            if await self.set(key, value, ttl, tags):
                count += 1
        return count

    async def cleanup_expired(self) -> int:
        async with self._lock:
            keys_to_delete: List[str] = []
            for key, entry in self._data.items():
                if self._is_expired(entry):
                    keys_to_delete.append(key)

            for key in keys_to_delete:
                await self._delete_key(key)

            return len(keys_to_delete)

    def get_stats(self) -> Dict[str, Any]:
        return {
            'hit_rate': self._stats.hit_rate,
            'current_keys': len(self._data),
            'evictions': self._stats.evictions
        }


cache_engine = CacheEngine()


class SetRequest(BaseModel):
    value: Any
    ttl: int = Field(default=0, ge=0)
    tags: List[str] = Field(default_factory=list, max_length=10)


class BatchSetItem(BaseModel):
    key: str
    value: Any


class BatchSetRequest(BaseModel):
    items: List[BatchSetItem] = Field(..., min_length=1, max_length=100)
    ttl: int = Field(default=0, ge=0)
    tags: List[str] = Field(default_factory=list, max_length=10)


class BatchGetRequest(BaseModel):
    keys: List[str] = Field(..., min_length=1)


class BatchGetResponse(BaseModel):
    hits: Dict[str, Any]
    misses: List[str]


class StatsResponse(BaseModel):
    hit_rate: float
    current_keys: int
    evictions: int


@asynccontextmanager
async def lifespan(app: FastAPI):
    cleanup_task = asyncio.create_task(periodic_cleanup())
    yield
    cleanup_task.cancel()
    try:
        await cleanup_task
    except asyncio.CancelledError:
        pass


async def periodic_cleanup():
    while True:
        await asyncio.sleep(10)
        try:
            await cache_engine.cleanup_expired()
        except Exception:
            pass


app = FastAPI(lifespan=lifespan)


@app.post("/cache/{key}")
async def set_cache(key: str, request: SetRequest):
    if len(request.tags) > 10:
        raise HTTPException(status_code=400, detail="Tags cannot exceed 10")
    await cache_engine.set(key, request.value, request.ttl, request.tags)
    return {"status": "ok"}


@app.get("/cache/{key}")
async def get_cache(key: str):
    value = await cache_engine.get(key)
    if value is None:
        raise HTTPException(status_code=404, detail="Key not found")
    return {"value": value}


@app.delete("/cache/{key}")
async def delete_cache(key: str):
    deleted = await cache_engine.delete(key)
    return {"deleted": deleted}


@app.post("/cache/batch/get", response_model=BatchGetResponse)
async def batch_get(request: BatchGetRequest):
    results = await cache_engine.batch_get(request.keys)
    hits: Dict[str, Any] = {}
    misses: List[str] = []
    for key, value in results.items():
        if value is not None:
            hits[key] = value
        else:
            misses.append(key)
    return BatchGetResponse(hits=hits, misses=misses)


@app.post("/cache/batch/set")
async def batch_set(request: BatchSetRequest):
    if len(request.items) > 100:
        raise HTTPException(status_code=400, detail="Batch set cannot exceed 100 items")
    if len(request.tags) > 10:
        raise HTTPException(status_code=400, detail="Tags cannot exceed 10")

    items_dict = {item.key: item.value for item in request.items}
    count = await cache_engine.batch_set(items_dict, request.ttl, request.tags)
    return {"set_count": count}


@app.delete("/cache/tag/{tag}")
async def delete_by_tag(tag: str):
    deleted = await cache_engine.delete_by_tag(tag)
    return {"deleted": deleted}


@app.get("/stats", response_model=StatsResponse)
async def get_stats():
    stats = cache_engine.get_stats()
    return StatsResponse(**stats)


if __name__ == "__main__":
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run(app, host="0.0.0.0", port=port)
