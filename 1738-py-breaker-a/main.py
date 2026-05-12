import os
import time
import asyncio
import httpx
from enum import Enum
from typing import List, Dict, Optional, Any
from datetime import datetime, timedelta
from pydantic import BaseModel, Field
from fastapi import FastAPI, HTTPException
from contextlib import asynccontextmanager


class CircuitState(str, Enum):
    CLOSED = "closed"
    OPEN = "open"
    HALF_OPEN = "half_open"


class CircuitBreakerConfig(BaseModel):
    failure_threshold: int = Field(default=5, ge=1)
    recovery_timeout: int = Field(default=30, ge=1)
    half_open_requests: int = Field(default=3, ge=1)
    callbacks: List[str] = Field(default_factory=list)


class StateChangeRecord(BaseModel):
    service_name: str
    old_state: CircuitState
    new_state: CircuitState
    reason: str
    timestamp: datetime


class CircuitBreakerStatus(BaseModel):
    service_name: str
    state: CircuitState
    failure_count: int
    half_open_attempts: int
    last_state_change: Optional[datetime]
    next_recovery_time: Optional[datetime]


class CircuitBreaker:
    def __init__(self, service_name: str, config: CircuitBreakerConfig):
        self.service_name = service_name
        self.config = config
        self._state: CircuitState = CircuitState.CLOSED
        self._failure_count: int = 0
        self._half_open_attempts: int = 0
        self._last_state_change: Optional[datetime] = None
        self._open_time: Optional[datetime] = None
        self._history: List[StateChangeRecord] = []
        self._lock = asyncio.Lock()

    @property
    def state(self) -> CircuitState:
        return self._state

    @property
    def failure_count(self) -> int:
        return self._failure_count

    @property
    def half_open_attempts(self) -> int:
        return self._half_open_attempts

    @property
    def history(self) -> List[StateChangeRecord]:
        return list(self._history)

    @property
    def next_recovery_time(self) -> Optional[datetime]:
        if self._state == CircuitState.OPEN and self._open_time:
            return self._open_time + timedelta(seconds=self.config.recovery_timeout)
        return None

    def get_status(self) -> CircuitBreakerStatus:
        return CircuitBreakerStatus(
            service_name=self.service_name,
            state=self._state,
            failure_count=self._failure_count,
            half_open_attempts=self._half_open_attempts,
            last_state_change=self._last_state_change,
            next_recovery_time=self.next_recovery_time
        )

    async def _change_state(self, new_state: CircuitState, reason: str):
        old_state = self._state
        if old_state == new_state:
            return

        self._state = new_state
        self._last_state_change = datetime.now()

        record = StateChangeRecord(
            service_name=self.service_name,
            old_state=old_state,
            new_state=new_state,
            reason=reason,
            timestamp=self._last_state_change
        )
        self._history.append(record)

        if len(self._history) > 100:
            self._history = self._history[-100:]

        if new_state == CircuitState.OPEN:
            self._open_time = datetime.now()
        elif new_state == CircuitState.HALF_OPEN:
            self._half_open_attempts = 0
        elif new_state == CircuitState.CLOSED:
            self._failure_count = 0
            self._open_time = None

        if self.config.callbacks:
            asyncio.create_task(self._notify_callbacks(record))

    async def _notify_callbacks(self, record: StateChangeRecord):
        payload = {
            "service_name": record.service_name,
            "old_state": record.old_state.value,
            "new_state": record.new_state.value,
            "reason": record.reason,
            "timestamp": record.timestamp.isoformat()
        }

        async with httpx.AsyncClient(timeout=5.0) as client:
            for callback_url in self.config.callbacks:
                try:
                    await client.post(callback_url, json=payload)
                except Exception:
                    pass

    def _should_allow_request(self) -> bool:
        if self._state == CircuitState.CLOSED:
            return True

        if self._state == CircuitState.OPEN:
            if self._open_time is None:
                return False
            elapsed = (datetime.now() - self._open_time).total_seconds()
            if elapsed >= self.config.recovery_timeout:
                return True
            return False

        if self._state == CircuitState.HALF_OPEN:
            return self._half_open_attempts < self.config.half_open_requests

        return False

    async def on_request_start(self) -> bool:
        async with self._lock:
            if self._state == CircuitState.OPEN:
                if self._open_time is not None:
                    elapsed = (datetime.now() - self._open_time).total_seconds()
                    if elapsed >= self.config.recovery_timeout:
                        await self._change_state(
                            CircuitState.HALF_OPEN,
                            f"Recovery timeout ({self.config.recovery_timeout}s) elapsed"
                        )

            return self._should_allow_request()

    async def on_success(self):
        async with self._lock:
            if self._state == CircuitState.HALF_OPEN:
                self._half_open_attempts += 1
                if self._half_open_attempts >= self.config.half_open_requests:
                    await self._change_state(
                        CircuitState.CLOSED,
                        f"All {self.config.half_open_requests} half-open requests succeeded"
                    )
            elif self._state == CircuitState.CLOSED:
                self._failure_count = 0

    async def on_failure(self):
        async with self._lock:
            if self._state == CircuitState.HALF_OPEN:
                await self._change_state(
                    CircuitState.OPEN,
                    f"Half-open request failed"
                )
                self._open_time = datetime.now()
            elif self._state == CircuitState.CLOSED:
                self._failure_count += 1
                if self._failure_count >= self.config.failure_threshold:
                    await self._change_state(
                        CircuitState.OPEN,
                        f"Failure threshold reached: {self._failure_count}/{self.config.failure_threshold}"
                    )
                    self._open_time = datetime.now()


class CircuitBreakerManager:
    def __init__(self):
        self._breakers: Dict[str, CircuitBreaker] = {}
        self._default_config = CircuitBreakerConfig()
        self._lock = asyncio.Lock()

    def configure(self, service_name: str, config: CircuitBreakerConfig):
        if service_name in self._breakers:
            self._breakers[service_name].config = config
        else:
            self._breakers[service_name] = CircuitBreaker(service_name, config)

    def get_breaker(self, service_name: str) -> CircuitBreaker:
        if service_name not in self._breakers:
            self._breakers[service_name] = CircuitBreaker(service_name, self._default_config)
        return self._breakers[service_name]

    def list_services(self) -> List[str]:
        return list(self._breakers.keys())

    def get_all_statuses(self) -> List[CircuitBreakerStatus]:
        return [breaker.get_status() for breaker in self._breakers.values()]


breaker_manager = CircuitBreakerManager()


@asynccontextmanager
async def lifespan(app: FastAPI):
    breaker_manager.configure(
        "service_a",
        CircuitBreakerConfig(
            failure_threshold=3,
            recovery_timeout=10,
            half_open_requests=2,
            callbacks=[]
        )
    )
    breaker_manager.configure(
        "service_b",
        CircuitBreakerConfig(
            failure_threshold=5,
            recovery_timeout=20,
            half_open_requests=3,
            callbacks=[]
        )
    )
    yield


app = FastAPI(
    title="Circuit Breaker Service",
    description="A circuit breaker service for protecting downstream services",
    version="1.0.0",
    lifespan=lifespan
)


@app.get("/api/breakers", response_model=List[CircuitBreakerStatus])
async def list_all_breakers():
    return breaker_manager.get_all_statuses()


@app.get("/api/breakers/{service_name}", response_model=CircuitBreakerStatus)
async def get_breaker_status(service_name: str):
    if service_name not in breaker_manager.list_services():
        raise HTTPException(status_code=404, detail=f"Service '{service_name}' not found")
    breaker = breaker_manager.get_breaker(service_name)
    return breaker.get_status()


@app.get("/api/breakers/{service_name}/history", response_model=List[StateChangeRecord])
async def get_breaker_history(service_name: str, limit: int = 20):
    if service_name not in breaker_manager.list_services():
        raise HTTPException(status_code=404, detail=f"Service '{service_name}' not found")
    breaker = breaker_manager.get_breaker(service_name)
    history = breaker.history
    return history[-limit:]


@app.post("/api/breakers/{service_name}/configure")
async def configure_breaker(service_name: str, config: CircuitBreakerConfig):
    breaker_manager.configure(service_name, config)
    return {"message": f"Circuit breaker for '{service_name}' configured successfully"}


@app.post("/api/breakers/{service_name}/call")
async def simulate_call(service_name: str, success: bool = True):
    if service_name not in breaker_manager.list_services():
        raise HTTPException(status_code=404, detail=f"Service '{service_name}' not found")
    
    breaker = breaker_manager.get_breaker(service_name)
    
    allowed = await breaker.on_request_start()
    if not allowed:
        raise HTTPException(status_code=503, detail=f"Circuit breaker for '{service_name}' is OPEN")
    
    if success:
        await breaker.on_success()
        return {"service": service_name, "result": "success"}
    else:
        await breaker.on_failure()
        return {"service": service_name, "result": "failure"}


if __name__ == "__main__":
    import uvicorn
    port = int(os.environ.get("PORT", 8416))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)
