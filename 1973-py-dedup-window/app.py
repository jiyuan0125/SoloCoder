import asyncio
import json
import os
import time
from dataclasses import dataclass, field
from typing import Dict, List, Optional
from collections import deque

from aiohttp import web


@dataclass
class CachedResponse:
    status: int
    body: bytes
    headers: Dict[str, str]
    created_at: float


@dataclass
class DedupEvent:
    dedup_key: str
    business_type: str
    timestamp: float
    request_summary: str


class RuleManager:
    def __init__(self, default_rules: Dict[str, int] = None):
        self._rules: Dict[str, int] = default_rules or {}
        self._lock = asyncio.Lock()
    
    async def get_window(self, business_type: str) -> Optional[int]:
        async with self._lock:
            return self._rules.get(business_type)
    
    async def set_rule(self, business_type: str, window_seconds: int) -> None:
        async with self._lock:
            self._rules[business_type] = window_seconds
    
    async def delete_rule(self, business_type: str) -> bool:
        async with self._lock:
            if business_type in self._rules:
                del self._rules[business_type]
                return True
            return False
    
    async def get_all_rules(self) -> Dict[str, int]:
        async with self._lock:
            return dict(self._rules.copy())


class StatsManager:
    def __init__(self):
        self._stats: Dict[str, Dict[str, int]] = {}
        self._lock = asyncio.Lock()
        self._events: deque = deque(maxlen=1000)
        self._events_lock = asyncio.Lock()
    
    async def record_request(self, business_type: str, is_hit: bool) -> None:
        async with self._lock:
            if business_type not in self._stats:
                self._stats[business_type] = {"total": 0, "hits": 0}
            self._stats[business_type]["total"] += 1
            if is_hit:
                self._stats[business_type]["hits"] += 1
    
    async def add_event(self, event: DedupEvent) -> None:
        async with self._events_lock:
            self._events.append(event)
    
    async def get_stats(self) -> Dict[str, Dict[str, int]]:
        async with self._lock:
            return {k: v.copy() for k, v in self._stats.items()}
    
    async def get_events(self) -> List[Dict]:
        async with self._events_lock:
            return [
                {
                    "dedup_key": e.dedup_key,
                    "business_type": e.business_type,
                    "timestamp": e.timestamp,
                    "request_summary": e.request_summary
                }
                for e in self._events
            ]


class DedupEngine:
    def __init__(self, rule_manager: RuleManager):
        self._cache: Dict[str, CachedResponse] = {}
        self._locks: Dict[str, asyncio.Lock] = {}
        self._futures: Dict[str, asyncio.Future] = {}
        self._global_lock = asyncio.Lock()
        self._rule_manager = rule_manager
        self._cleanup_task: Optional[asyncio.Task] = None
    
    async def start(self) -> None:
        self._cleanup_task = asyncio.create_task(self._cleanup_loop())
    
    async def stop(self) -> None:
        if self._cleanup_task:
            self._cleanup_task.cancel()
    
    async def _cleanup_loop(self) -> None:
        while True:
            await asyncio.sleep(60)
            await self._cleanup()
    
    async def _cleanup(self) -> None:
        now = time.time()
        expired_keys = []
        for key, cached in self._cache.items():
            business_type = key.split(':')[0] if ':' in key else "default"
            window = await self._rule_manager.get_window(business_type)
            if window is None:
                expired_keys.append(key)
            elif cached.created_at + window < now:
                expired_keys.append(key)
        async with self._global_lock:
            for key in expired_keys:
                if key in self._cache:
                    del self._cache[key]
                if key in self._locks:
                    del self._locks[key]
                if key in self._futures:
                    del self._futures[key]
    
    async def get_or_execute(self, dedup_key: str, business_type: str, execute_func):
        window = await self._rule_manager.get_window(business_type)
        if window is None:
            return await execute_func(), False
        
        now = time.time()
        
        async with self._global_lock:
            if dedup_key in self._cache:
                cached = self._cache[dedup_key]
                if cached.created_at + window > now:
                    return cached, True
            
            if dedup_key in self._futures:
                return await self._futures[dedup_key], True
            
            if dedup_key not in self._locks:
                self._locks[dedup_key] = asyncio.Lock()
            
            lock = self._locks[dedup_key]
            future = asyncio.get_event_loop().create_future()
            self._futures[dedup_key] = future
        
        async with lock:
            if dedup_key in self._cache:
                cached = self._cache[dedup_key]
                if cached.created_at + window > now:
                    async with self._global_lock:
                        if dedup_key in self._futures:
                            self._futures[dedup_key].set_result(cached)
                            del self._futures[dedup_key]
                    return cached, True
            
            result = await execute_func()
            cached = CachedResponse(
                status=result["status"],
                body=result["body"],
                headers=result.get("headers", {}),
                created_at=time.time()
            )
            
            async with self._global_lock:
                self._cache[dedup_key] = cached
                if dedup_key in self._futures:
                    self._futures[dedup_key].set_result(cached)
                    del self._futures[dedup_key]
            
            return cached, False


async def handle_dedup_request(request: web.Request) -> web.Response:
    engine = request.app["dedup_engine"]
    stats_manager = request.app["stats_manager"]
    
    try:
        raw_body = await request.read()
        try:
            data = json.loads(raw_body)
        except json.JSONDecodeError:
            return web.Response(status=400, text="Invalid JSON")
        
        dedup_key = data.get("dedup_key")
        if not dedup_key:
            return web.Response(status=400, text="Missing dedup_key")
        
        business_type = dedup_key.split(':')[0] if ':' in dedup_key else "default"
        
        request_summary = json.dumps({
            "method": request.method,
            "path": request.path,
            "body": data
        }, ensure_ascii=False)[:500]
        
        async def execute():
            backend_url = data.get("backend_url")
            backend_method = data.get("backend_method", "POST")
            backend_headers = data.get("backend_headers", {})
            backend_body = data.get("backend_body", {})
            
            if not backend_url:
                return {
                    "status": 200,
                    "body": b'{"status": "ok"}',
                    "headers": {"Content-Type": "application/json"}
                }
            
            session = request.app.get("client_session")
            async with session.request(
                method=backend_method,
                url=backend_url,
                headers=backend_headers,
                json=backend_body
            ) as resp:
                resp_body = await resp.read()
                resp_headers = dict(resp.headers)
                return {
                    "status": resp.status,
                    "body": resp_body,
                    "headers": resp_headers
                }
        
        result, is_hit = await engine.get_or_execute(dedup_key, business_type, execute)
        
        await stats_manager.record_request(business_type, is_hit)
        
        if is_hit:
            event = DedupEvent(
                dedup_key=dedup_key,
                business_type=business_type,
                timestamp=time.time(),
                request_summary=request_summary
            )
            await stats_manager.add_event(event)
        
        headers = result.headers.copy()
        headers["X-Dedup"] = "true" if is_hit else "false"
        
        return web.Response(
            status=result.status,
            body=result.body,
            headers=headers
        )
    
    except Exception as e:
        return web.Response(status=500, text=str(e))


async def handle_stats(request: web.Request) -> web.Response:
    stats_manager = request.app["stats_manager"]
    stats = await stats_manager.get_stats()
    
    result = {}
    for business_type, data in stats.items():
        result[business_type] = {
            "total_requests": data["total"],
            "dedup_hits": data["hits"],
            "saved_calls": data["hits"],
            "hit_rate": round(data["hits"] / data["total"] if data["total"] > 0 else 0, 4)
        }
    
    return web.json_response(result)


async def handle_events(request: web.Request) -> web.Response:
    stats_manager = request.app["stats_manager"]
    events = await stats_manager.get_events()
    return web.json_response(events)


async def handle_list_rules(request: web.Request) -> web.Response:
    rule_manager = request.app["rule_manager"]
    rules = await rule_manager.get_all_rules()
    return web.json_response(rules)


async def handle_create_rule(request: web.Request) -> web.Response:
    rule_manager = request.app["rule_manager"]
    try:
        data = await request.json()
    except json.JSONDecodeError:
        return web.Response(status=400, text="Invalid JSON")
    
    business_type = data.get("business_type")
    window_seconds = data.get("window_seconds")
    
    if not business_type or not isinstance(window_seconds, int) or window_seconds <= 0:
        return web.Response(status=400, text="Invalid rule: business_type and positive window_seconds required")
    
    await rule_manager.set_rule(business_type, window_seconds)
    return web.json_response({"business_type": business_type, "window_seconds": window_seconds}, status=201)


async def handle_update_rule(request: web.Request) -> web.Response:
    rule_manager = request.app["rule_manager"]
    business_type = request.match_info["type"]
    
    try:
        data = await request.json()
    except json.JSONDecodeError:
        return web.Response(status=400, text="Invalid JSON")
    
    window_seconds = data.get("window_seconds")
    
    if not isinstance(window_seconds, int) or window_seconds <= 0:
        return web.Response(status=400, text="Invalid: positive window_seconds required")
    
    existing = await rule_manager.get_window(business_type)
    if existing is None:
        return web.Response(status=404, text="Rule not found")
    
    await rule_manager.set_rule(business_type, window_seconds)
    return web.json_response({"business_type": business_type, "window_seconds": window_seconds})


async def handle_delete_rule(request: web.Request) -> web.Response:
    rule_manager = request.app["rule_manager"]
    business_type = request.match_info["type"]
    
    deleted = await rule_manager.delete_rule(business_type)
    if deleted:
        return web.Response(status=204)
    else:
        return web.Response(status=404, text="Rule not found")


async def create_app() -> web.Application:
    from aiohttp import ClientSession
    
    app = web.Application()
    
    default_rules = {
        "payment": 300,
        "sms": 60
    }
    
    rule_manager = RuleManager(default_rules)
    dedup_engine = DedupEngine(rule_manager)
    stats_manager = StatsManager()
    client_session = ClientSession()
    
    app["rule_manager"] = rule_manager
    app["dedup_engine"] = dedup_engine
    app["stats_manager"] = stats_manager
    app["client_session"] = client_session
    
    async def on_startup(app):
        await dedup_engine.start()
    
    async def on_cleanup(app):
        await dedup_engine.stop()
        await client_session.close()
    
    app.on_startup.append(on_startup)
    app.on_cleanup.append(on_cleanup)
    
    app.router.add_post("/dedup", handle_dedup_request)
    app.router.add_get("/dedup/stats", handle_stats)
    app.router.add_get("/dedup/events", handle_events)
    app.router.add_get("/rules", handle_list_rules)
    app.router.add_post("/rules", handle_create_rule)
    app.router.add_put("/rules/{type}", handle_update_rule)
    app.router.add_delete("/rules/{type}", handle_delete_rule)
    
    return app


if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8080))
    app = create_app()
    web.run_app(app, host="0.0.0.0", port=port)
