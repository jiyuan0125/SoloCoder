import os
import time
import threading
from enum import Enum
from datetime import datetime, timezone
from typing import Dict, List, Optional, Any
from collections import deque

from fastapi import FastAPI, Request, HTTPException, Query, Header
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
import uvicorn


class RuleLevel(str, Enum):
    IP = "ip"
    USER = "user"
    API = "api"


class RateLimitRule(BaseModel):
    id: str
    path: str
    window_seconds: int
    threshold: int
    level: RuleLevel


class RateLimitRuleCreate(BaseModel):
    path: str
    window_seconds: int = Field(..., gt=0)
    threshold: int = Field(..., gt=0)
    level: RuleLevel


class RateLimitEvent(BaseModel):
    timestamp: float
    client_ip: str
    user_id: Optional[str]
    request_path: str
    level: RuleLevel


class RateLimitStatus(BaseModel):
    path: str
    level: RuleLevel
    current_count: int
    threshold: int
    remaining_quota: int


class SlidingWindow:
    def __init__(self, window_seconds: int):
        self.window_seconds = window_seconds
        self._lock = threading.Lock()
        self._bucket = {}

    def _current_second(self) -> int:
        return int(time.time())

    def _cleanup(self, current_second: int):
        cutoff = current_second - self.window_seconds
        keys_to_delete = [k for k in self._bucket.keys() if k <= cutoff]
        for k in keys_to_delete:
            del self._bucket[k]

    def increment(self) -> int:
        with self._lock:
            current = self._current_second()
            self._cleanup(current)
            self._bucket[current] = self._bucket.get(current, 0) + 1
            return self._bucket[current]

    def get_count(self) -> int:
        with self._lock:
            current = self._current_second()
            self._cleanup(current)
            return sum(self._bucket.values())

    def get_remaining(self, threshold: int) -> int:
        count = self.get_count()
        return max(0, threshold - count)

    def get_wait_time(self, threshold: int) -> int:
        count = self.get_count()
        if count < threshold:
            return 0
        with self._lock:
            if not self._bucket:
                return 0
            oldest_second = min(self._bucket.keys())
            return max(1, (oldest_second + self.window_seconds) - self._current_second())


class SlidingWindowManager:
    def __init__(self):
        self._lock = threading.Lock()
        self._windows: Dict[str, SlidingWindow] = {}

    def get_or_create(self, key: str, window_seconds: int) -> SlidingWindow:
        with self._lock:
            if key not in self._windows:
                self._windows[key] = SlidingWindow(window_seconds)
            return self._windows[key]

    def remove_by_path_prefix(self, path: str):
        with self._lock:
            keys_to_delete = [k for k in self._windows.keys() if k.startswith(f"{path}:")]
            for k in keys_to_delete:
                del self._windows[k]

    def get_windows_for_path(self, path: str) -> Dict[str, SlidingWindow]:
        with self._lock:
            return {
                k: v for k, v in self._windows.items()
                if k.startswith(f"{path}:")
            }


class RateLimitStore:
    def __init__(self):
        self._lock = threading.Lock()
        self._rules: Dict[str, RateLimitRule] = {}
        self._rule_counter = 0
        self._windows = SlidingWindowManager()
        self._events = deque(maxlen=5000)

    def add_rule(self, path: str, window_seconds: int, threshold: int, level: RuleLevel) -> RateLimitRule:
        with self._lock:
            self._rule_counter += 1
            rule_id = f"rule_{self._rule_counter}"
            rule = RateLimitRule(
                id=rule_id,
                path=path,
                window_seconds=window_seconds,
                threshold=threshold,
                level=level
            )
            self._rules[rule_id] = rule
            return rule

    def delete_rule(self, rule_id: str) -> bool:
        with self._lock:
            if rule_id in self._rules:
                rule = self._rules[rule_id]
                del self._rules[rule_id]
                self._windows.remove_by_path_prefix(rule.path)
                return True
            return False

    def get_rules(self) -> List[RateLimitRule]:
        with self._lock:
            return list(self._rules.values())

    def get_rules_by_path(self, path: str) -> Dict[RuleLevel, RateLimitRule]:
        with self._lock:
            result = {}
            for rule in self._rules.values():
                if rule.path == path:
                    result[rule.level] = rule
            return result

    def check_and_increment(
        self,
        path: str,
        client_ip: str,
        user_id: Optional[str]
    ) -> Optional[Dict[str, Any]]:
        rules_by_level = self.get_rules_by_path(path)

        check_order = [RuleLevel.IP, RuleLevel.USER, RuleLevel.API]
        for level in check_order:
            rule = rules_by_level.get(level)
            if not rule:
                continue

            if level == RuleLevel.IP:
                key = f"{path}:ip:{client_ip}"
            elif level == RuleLevel.USER and user_id:
                key = f"{path}:user:{user_id}"
            elif level == RuleLevel.API:
                key = f"{path}:api"
            else:
                continue

            window = self._windows.get_or_create(key, rule.window_seconds)
            current_count = window.get_count()

            if current_count >= rule.threshold:
                return {
                    "triggered": True,
                    "level": level,
                    "wait_time": window.get_wait_time(rule.threshold),
                    "rule_description": f"Level: {level.value}, Path: {path}, Window: {rule.window_seconds}s, Threshold: {rule.threshold}"
                }

            window.increment()

        return None

    def get_status(self, path: str, client_ip: Optional[str] = None, user_id: Optional[str] = None) -> List[RateLimitStatus]:
        rules_by_level = self.get_rules_by_path(path)
        status_list = []

        check_order = [RuleLevel.IP, RuleLevel.USER, RuleLevel.API]
        for level in check_order:
            rule = rules_by_level.get(level)
            if not rule:
                continue

            if level == RuleLevel.IP:
                if not client_ip:
                    continue
                key = f"{path}:ip:{client_ip}"
            elif level == RuleLevel.USER:
                if not user_id:
                    continue
                key = f"{path}:user:{user_id}"
            elif level == RuleLevel.API:
                key = f"{path}:api"
            else:
                continue

            windows = self._windows.get_windows_for_path(path)
            window = windows.get(key)

            if window:
                current_count = window.get_count()
            else:
                current_count = 0

            status_list.append(RateLimitStatus(
                path=path,
                level=level,
                current_count=current_count,
                threshold=rule.threshold,
                remaining_quota=max(0, rule.threshold - current_count)
            ))

        return status_list

    def add_event(self, client_ip: str, user_id: Optional[str], request_path: str, level: RuleLevel):
        event = RateLimitEvent(
            timestamp=time.time(),
            client_ip=client_ip,
            user_id=user_id,
            request_path=request_path,
            level=level
        )
        with self._lock:
            self._events.append(event)

    def get_events(
        self,
        start_time: Optional[float] = None,
        end_time: Optional[float] = None,
        path_filter: Optional[str] = None
    ) -> List[Dict[str, Any]]:
        with self._lock:
            events = list(self._events)

        filtered = []
        for event in events:
            if start_time and event.timestamp < start_time:
                continue
            if end_time and event.timestamp > end_time:
                continue
            if path_filter and event.request_path != path_filter:
                continue
            filtered.append({
                "timestamp": event.timestamp,
                "datetime": datetime.fromtimestamp(event.timestamp, tz=timezone.utc).isoformat(),
                "client_ip": event.client_ip,
                "user_id": event.user_id,
                "request_path": event.request_path,
                "level": event.level.value
            })

        return filtered


store = RateLimitStore()
app = FastAPI(title="Sliding Window Rate Limiter")


@app.get("/")
async def root():
    return {"message": "Sliding Window Rate Limiter API", "version": "1.0"}


@app.get("/health")
async def health():
    return {"status": "healthy", "timestamp": time.time()}


@app.post("/rules", response_model=RateLimitRule, status_code=201)
async def create_rule(rule_create: RateLimitRuleCreate):
    rule = store.add_rule(
        path=rule_create.path,
        window_seconds=rule_create.window_seconds,
        threshold=rule_create.threshold,
        level=rule_create.level
    )
    return rule


@app.delete("/rules/{rule_id}")
async def delete_rule(rule_id: str):
    if store.delete_rule(rule_id):
        return {"message": "Rule deleted successfully"}
    raise HTTPException(status_code=404, detail="Rule not found")


@app.get("/rules", response_model=List[RateLimitRule])
async def list_rules():
    return store.get_rules()


@app.get("/status/{path:path}", response_model=List[RateLimitStatus])
async def get_path_status(
    path: str,
    client_ip: Optional[str] = Query(None, description="Client IP address for IP-level check"),
    x_user_id: Optional[str] = Header(None, alias="X-User-Id")
):
    if not path.startswith("/"):
        path = "/" + path
    status_list = store.get_status(path, client_ip, x_user_id)
    if not status_list:
        raise HTTPException(status_code=404, detail="No rules found for this path")
    return status_list


@app.get("/rate-limit-events")
async def get_rate_limit_events(
    start_time: Optional[float] = Query(None, description="Unix timestamp start"),
    end_time: Optional[float] = Query(None, description="Unix timestamp end"),
    path: Optional[str] = Query(None, description="Filter by request path")
):
    events = store.get_events(start_time, end_time, path)
    return {"total": len(events), "events": events}


@app.post("/check")
async def check_rate_limit(
    request: Request,
    x_user_id: Optional[str] = Header(None, alias="X-User-Id")
):
    try:
        body = await request.json()
        path = body.get("path")
    except Exception:
        raise HTTPException(status_code=400, detail="Invalid request body")

    if not path:
        raise HTTPException(status_code=400, detail="Path is required")

    client_ip = request.client.host if request.client else "unknown"

    result = store.check_and_increment(path, client_ip, x_user_id)

    if result:
        store.add_event(client_ip, x_user_id, path, result["level"])
        return JSONResponse(
            status_code=429,
            content={
                "error": "Rate Limit Exceeded",
                "retry_after_seconds": result["wait_time"],
                "triggered_rule": result["rule_description"],
                "triggered_level": result["level"].value
            }
        )

    return {
        "allowed": True,
        "path": path,
        "timestamp": time.time()
    }


def get_client_ip(request: Request) -> str:
    forwarded = request.headers.get("X-Forwarded-For")
    if forwarded:
        return forwarded.split(",")[0].strip()
    if request.client:
        return request.client.host
    return "unknown"


def main():
    port = int(os.environ.get("PORT", 8000))
    uvicorn.run(app, host="0.0.0.0", port=port)


if __name__ == "__main__":
    main()
