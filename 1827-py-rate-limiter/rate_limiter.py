import asyncio
import time
from typing import Dict, Optional, List, Tuple
from dataclasses import dataclass, field
from collections import defaultdict
from fastapi import Request, Response
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware
from starlette.types import ASGIApp
from pydantic import BaseModel, Field


class RateLimitConfig(BaseModel):
    window_size: int = Field(gt=0, description="时间窗口大小（秒）")
    max_requests: int = Field(gt=0, description="窗口内最大请求数")


class RateLimitStatus(BaseModel):
    dimension: str = Field(description="限流维度：IP/路由")
    window_size: int = Field(description="时间窗口大小（秒）")
    used: int = Field(description="窗口内已用额度")
    remaining: int = Field(description="窗口内剩余额度")
    retry_after: int = Field(description="剩余冷却时间（秒）")


@dataclass
class RateLimitEntry:
    count: int = 0
    window_start: float = 0.0
    client_ip: str = ""
    prefix: str = ""


class RateLimiter:
    def __init__(self):
        self.configs: Dict[str, RateLimitConfig] = {}
        self.counters: Dict[Tuple[str, str], RateLimitEntry] = {}
        self._cleanup_task: Optional[asyncio.Task] = None
        self._lock: asyncio.Lock = asyncio.Lock()

    def start_cleanup_task(self):
        if self._cleanup_task is None or self._cleanup_task.done():
            self._cleanup_task = asyncio.create_task(self._periodic_cleanup())

    def stop_cleanup_task(self):
        if self._cleanup_task and not self._cleanup_task.done():
            self._cleanup_task.cancel()

    async def _periodic_cleanup(self):
        while True:
            await asyncio.sleep(1)
            await self._cleanup_expired_entries()

    async def _cleanup_expired_entries(self):
        async with self._lock:
            current_time = time.time()
            keys_to_remove = []
            
            for key, entry in self.counters.items():
                config = self.configs.get(entry.prefix)
                if not config:
                    keys_to_remove.append(key)
                    continue
                
                if current_time - entry.window_start >= config.window_size:
                    keys_to_remove.append(key)
            
            for key in keys_to_remove:
                del self.counters[key]

    async def check_rate_limit(self, client_ip: str, prefix: str) -> Optional[RateLimitStatus]:
        config = self.configs.get(prefix)
        if not config:
            return None
        
        async with self._lock:
            current_time = time.time()
            key = (client_ip, prefix)
            entry = self.counters.get(key)
            
            if entry is None or current_time - entry.window_start >= config.window_size:
                entry = RateLimitEntry(
                    count=0,
                    window_start=current_time,
                    client_ip=client_ip,
                    prefix=prefix
                )
                self.counters[key] = entry
            
            if entry.count >= config.max_requests:
                elapsed = current_time - entry.window_start
                retry_after = max(0, config.window_size - int(elapsed))
                return RateLimitStatus(
                    dimension="IP/路由",
                    window_size=config.window_size,
                    used=entry.count,
                    remaining=0,
                    retry_after=retry_after
                )
            
            entry.count += 1
            return RateLimitStatus(
                dimension="IP/路由",
                window_size=config.window_size,
                used=entry.count,
                remaining=config.max_requests - entry.count,
                retry_after=0
            )

    async def check_all_prefixes(self, client_ip: str, path: str) -> Optional[RateLimitStatus]:
        statuses: List[RateLimitStatus] = []
        
        matching_prefixes = []
        for prefix in self.configs.keys():
            if path.startswith(prefix):
                matching_prefixes.append(prefix)
        
        if not matching_prefixes:
            return None
        
        for prefix in matching_prefixes:
            status = await self.check_rate_limit(client_ip, prefix)
            if status and status.retry_after > 0:
                statuses.append(status)
        
        if not statuses:
            return None
        
        max_retry_status = max(statuses, key=lambda s: s.retry_after)
        return max_retry_status

    async def get_all_configs(self) -> Dict[str, RateLimitConfig]:
        async with self._lock:
            return dict(self.configs)

    async def set_config(self, prefix: str, config: RateLimitConfig) -> None:
        async with self._lock:
            self.configs[prefix] = config

    async def delete_config(self, prefix: str) -> bool:
        async with self._lock:
            if prefix in self.configs:
                del self.configs[prefix]
                return True
            return False

    async def get_usage(
        self,
        client_ip: Optional[str] = None,
        prefix: Optional[str] = None,
        page: int = 1,
        page_size: int = 20
    ) -> Tuple[List[dict], int]:
        async with self._lock:
            current_time = time.time()
            results = []
            
            for key, entry in self.counters.items():
                entry_ip, entry_prefix = key
                
                if client_ip and entry_ip != client_ip:
                    continue
                if prefix and entry_prefix != prefix:
                    continue
                
                config = self.configs.get(entry_prefix)
                if not config:
                    continue
                
                elapsed = current_time - entry.window_start
                remaining = config.max_requests - entry.count
                retry_after = max(0, config.window_size - int(elapsed)) if remaining <= 0 else 0
                
                results.append({
                    "client_ip": entry_ip,
                    "prefix": entry_prefix,
                    "used": entry.count,
                    "remaining": remaining,
                    "retry_after": retry_after,
                    "window_size": config.window_size
                })
            
            total = len(results)
            start = (page - 1) * page_size
            end = start + page_size
            paginated_results = results[start:end]
            
            return paginated_results, total


class RateLimitMiddleware(BaseHTTPMiddleware):
    def __init__(self, app: ASGIApp, rate_limiter: RateLimiter):
        super().__init__(app)
        self.rate_limiter = rate_limiter

    async def dispatch(self, request: Request, call_next):
        client_ip = self._get_client_ip(request)
        path = request.url.path
        
        status = await self.rate_limiter.check_all_prefixes(client_ip, path)
        
        if status and status.retry_after > 0:
            return JSONResponse(
                status_code=429,
                content={
                    "detail": "Too Many Requests",
                    "rate_limit": {
                        "dimension": status.dimension,
                        "window_size": status.window_size,
                        "used": status.used,
                        "remaining": status.remaining,
                        "retry_after": status.retry_after
                    }
                },
                headers={"Retry-After": str(status.retry_after)}
            )
        
        response = await call_next(request)
        return response

    def _get_client_ip(self, request: Request) -> str:
        x_forwarded_for = request.headers.get("x-forwarded-for")
        if x_forwarded_for:
            return x_forwarded_for.split(",")[0].strip()
        return request.client.host if request.client else "unknown"
