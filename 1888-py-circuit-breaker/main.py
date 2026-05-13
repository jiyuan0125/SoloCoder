import os
import time
import asyncio
from enum import Enum
from typing import Dict, List, Optional
from datetime import datetime
from collections import deque

from fastapi import FastAPI, HTTPException
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field


class BreakerStatus(str, Enum):
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"
    UNKNOWN = "unknown"


class CallResult(str, Enum):
    SUCCESS = "success"
    FAILURE = "failure"


class CallRecord(BaseModel):
    timestamp: datetime
    result: CallResult
    duration_ms: float


class StateChangeRecord(BaseModel):
    timestamp: datetime
    reason: str


class BreakerConfig(BaseModel):
    service_name: str
    failure_threshold: int = Field(default=5, ge=1)
    open_duration_seconds: int = Field(default=30, ge=1)


class BreakerConfigUpdate(BaseModel):
    failure_threshold: Optional[int] = Field(default=None, ge=1)
    open_duration_seconds: Optional[int] = Field(default=None, ge=1)


class BreakerState(BaseModel):
    status: BreakerStatus
    service_name: str
    failure_threshold: int
    open_duration_seconds: int
    current_open_duration: Optional[int]
    consecutive_failures: int
    last_state_change: Optional[StateChangeRecord]
    recent_calls: List[CallRecord]


class BreakerListResponse(BaseModel):
    services: List[str]


MAX_OPEN_DURATION_SECONDS = 5 * 60
MAX_CALL_HISTORY = 20


class CircuitBreaker:
    def __init__(self, service_name: str, failure_threshold: int = 5, open_duration_seconds: int = 30):
        self.service_name = service_name
        self.failure_threshold = failure_threshold
        self.open_duration_seconds = open_duration_seconds
        self.current_open_duration = open_duration_seconds
        
        self._status = BreakerStatus.CLOSED
        self.consecutive_failures = 0
        self.open_until = 0.0
        self._half_open_in_progress = False
        self._lock = asyncio.Lock()
        
        self.last_state_change: Optional[StateChangeRecord] = None
        self.recent_calls: deque = deque(maxlen=MAX_CALL_HISTORY)
        self._change_state(BreakerStatus.CLOSED, "初始化熔断器")

    def update_config(self, failure_threshold: Optional[int] = None, open_duration_seconds: Optional[int] = None):
        if failure_threshold is not None:
            self.failure_threshold = failure_threshold
        if open_duration_seconds is not None:
            self.open_duration_seconds = open_duration_seconds

    def _change_state(self, new_status: BreakerStatus, reason: str):
        self._status = new_status
        self.last_state_change = StateChangeRecord(
            timestamp=datetime.now(),
            reason=reason
        )

    @property
    def status(self) -> BreakerStatus:
        if self._status == BreakerStatus.OPEN and time.time() >= self.open_until:
            self._change_state(BreakerStatus.HALF_OPEN, f"{self.current_open_duration}秒冷却期结束，进入半开状态")
            self._half_open_in_progress = False
        return self._status

    def _open(self, reason: str):
        self.open_until = time.time() + self.current_open_duration
        self._change_state(BreakerStatus.OPEN, reason)
        self.current_open_duration = min(self.current_open_duration * 2, MAX_OPEN_DURATION_SECONDS)

    def _record_internal(self, success: bool, duration_ms: float):
        self.recent_calls.append(CallRecord(
            timestamp=datetime.now(),
            result=CallResult.SUCCESS if success else CallResult.FAILURE,
            duration_ms=duration_ms
        ))

        current_status = self.status

        if current_status == BreakerStatus.CLOSED:
            if not success:
                self.consecutive_failures += 1
                if self.consecutive_failures >= self.failure_threshold:
                    self._open(f"连续{self.consecutive_failures}次失败触发打开")
            else:
                self.consecutive_failures = 0

        elif current_status == BreakerStatus.HALF_OPEN:
            self._half_open_in_progress = False
            if success:
                self.consecutive_failures = 0
                self.current_open_duration = self.open_duration_seconds
                self._change_state(BreakerStatus.CLOSED, "半开探测成功，熔断器关闭")
            else:
                self._open("半开探测失败，重新打开熔断器")

    async def try_allow_call(self) -> bool:
        async with self._lock:
            current_status = self.status
            
            if current_status == BreakerStatus.OPEN:
                return False
            
            if current_status == BreakerStatus.HALF_OPEN:
                if self._half_open_in_progress:
                    return False
                self._half_open_in_progress = True
                return True
            
            return True

    async def record_result(self, success: bool, duration_ms: float):
        async with self._lock:
            self._record_internal(success, duration_ms)

    def to_state_model(self) -> BreakerState:
        current_status = self.status
        return BreakerState(
            status=current_status,
            service_name=self.service_name,
            failure_threshold=self.failure_threshold,
            open_duration_seconds=self.open_duration_seconds,
            current_open_duration=self.current_open_duration if current_status == BreakerStatus.OPEN else None,
            consecutive_failures=self.consecutive_failures,
            last_state_change=self.last_state_change,
            recent_calls=list(self.recent_calls)
        )


class CircuitBreakerManager:
    def __init__(self):
        self.breakers: Dict[str, CircuitBreaker] = {}
        self._lock = asyncio.Lock()

    async def register(self, config: BreakerConfig) -> CircuitBreaker:
        async with self._lock:
            if config.service_name in self.breakers:
                raise HTTPException(
                    status_code=409,
                    detail=f"服务 {config.service_name} 已存在，使用 PUT 接口修改配置"
                )
            breaker = CircuitBreaker(
                service_name=config.service_name,
                failure_threshold=config.failure_threshold,
                open_duration_seconds=config.open_duration_seconds
            )
            self.breakers[config.service_name] = breaker
            return breaker

    async def update_config(self, service_name: str, update: BreakerConfigUpdate) -> CircuitBreaker:
        async with self._lock:
            if service_name not in self.breakers:
                raise HTTPException(
                    status_code=404,
                    detail=f"服务 {service_name} 未注册"
                )
            breaker = self.breakers[service_name]
            breaker.update_config(
                failure_threshold=update.failure_threshold,
                open_duration_seconds=update.open_duration_seconds
            )
            return breaker

    def get_breaker(self, service_name: str) -> Optional[CircuitBreaker]:
        return self.breakers.get(service_name)

    def list_services(self) -> List[str]:
        return list(self.breakers.keys())


app = FastAPI(title="熔断器服务")
manager = CircuitBreakerManager()


@app.post("/breakers", status_code=201, response_model=BreakerState)
async def register_breaker(config: BreakerConfig):
    breaker = await manager.register(config)
    return breaker.to_state_model()


@app.put("/breakers/{service}", response_model=BreakerState)
async def update_breaker(service: str, update: BreakerConfigUpdate):
    breaker = await manager.update_config(service, update)
    return breaker.to_state_model()


@app.get("/breakers", response_model=BreakerListResponse)
async def list_breakers():
    return BreakerListResponse(services=manager.list_services())


@app.get("/breakers/{service}")
async def get_breaker_status(service: str):
    breaker = manager.get_breaker(service)
    if breaker is None:
        return JSONResponse(
            status_code=200,
            content={
                "status": "unknown",
                "service_name": service,
                "message": f"服务 {service} 还没注册过熔断配置"
            }
        )
    return breaker.to_state_model()


@app.post("/breakers/{service}/call")
async def record_service_call(service: str, success: bool = True, duration_ms: float = 0.0):
    breaker = manager.get_breaker(service)
    if breaker is None:
        raise HTTPException(
            status_code=404,
            detail=f"服务 {service} 未注册"
        )
    
    allowed = await breaker.try_allow_call()
    
    if not allowed:
        current_status = breaker.status
        status_msg = "已打开" if current_status == BreakerStatus.OPEN else "处于半开状态，正在探测中"
        return JSONResponse(
            status_code=503,
            content={
                "service_name": service,
                "status": current_status.value,
                "message": f"服务 {service} 熔断器{status_msg}，请求被拒绝"
            }
        )
    
    await breaker.record_result(success, duration_ms)
    return breaker.to_state_model()


if __name__ == "__main__":
    import uvicorn
    port = int(os.environ.get("PORT", "8901"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
