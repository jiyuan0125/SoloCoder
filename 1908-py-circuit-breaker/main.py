from __future__ import annotations

import os
from typing import List, Optional

from fastapi import FastAPI, HTTPException, status
from pydantic import BaseModel, Field

from breaker_manager import manager
from circuit_breaker import CircuitConfig

app = FastAPI(title="Circuit Breaker API")


class ConfigUpdate(BaseModel):
    failure_rate_threshold: Optional[float] = None
    slow_call_rate_threshold: Optional[float] = None
    slow_call_threshold_ms: Optional[int] = None
    window_size: Optional[int] = Field(default=None, ge=1)
    half_open_probes: Optional[int] = Field(default=None, ge=1)


class EventResponse(BaseModel):
    service_name: str
    from_state: str
    to_state: str
    trigger_reason: str
    timestamp: float


class ServiceResponse(BaseModel):
    service_name: str
    status: str
    config: Optional[dict] = None
    events: List[EventResponse] = []


@app.get("/health")
def health_check():
    return {"status": "ok"}


@app.get("/breakers")
def list_breakers():
    services = manager.list_services()
    result = []
    for name, state in services.items():
        breaker = manager.get_or_create(name)
        config = breaker.config if breaker.is_configured else None
        result.append({
            "service_name": name,
            "status": state,
            "config": config.__dict__ if config else None,
        })
    return {"services": result}


@app.get("/breakers/{service_name}")
def get_breaker(service_name: str):
    breaker = manager.get_or_create(service_name)
    if not breaker.is_configured:
        return ServiceResponse(
            service_name=service_name,
            status="unconfigured",
            config=None,
            events=[],
        )
    events = [
        EventResponse(
            service_name=e.service_name,
            from_state=e.from_state,
            to_state=e.to_state,
            trigger_reason=e.trigger_reason,
            timestamp=e.timestamp,
        )
        for e in breaker.events
    ]
    return ServiceResponse(
        service_name=service_name,
        status=breaker.state.value,
        config=breaker.config.__dict__,
        events=events,
    )


@app.put("/breakers/{service_name}/config", status_code=status.HTTP_200_OK)
def update_config(service_name: str, config_update: ConfigUpdate):
    existing = manager.get_or_create(service_name)
    if existing.is_configured:
        current = existing.config
        new_config = CircuitConfig(
            failure_rate_threshold=config_update.failure_rate_threshold or current.failure_rate_threshold,
            slow_call_rate_threshold=config_update.slow_call_rate_threshold or current.slow_call_rate_threshold,
            slow_call_threshold_ms=config_update.slow_call_threshold_ms or current.slow_call_threshold_ms,
            window_size=config_update.window_size or current.window_size,
            half_open_probes=config_update.half_open_probes or current.half_open_probes,
        )
    else:
        default = CircuitConfig()
        new_config = CircuitConfig(
            failure_rate_threshold=config_update.failure_rate_threshold or default.failure_rate_threshold,
            slow_call_rate_threshold=config_update.slow_call_rate_threshold or default.slow_call_rate_threshold,
            slow_call_threshold_ms=config_update.slow_call_threshold_ms or default.slow_call_threshold_ms,
            window_size=config_update.window_size or default.window_size,
            half_open_probes=config_update.half_open_probes or default.half_open_probes,
        )
    manager.configure(service_name, new_config)
    breaker = manager.get_or_create(service_name)
    return {
        "service_name": service_name,
        "config": breaker.config.__dict__,
    }


@app.post("/breakers/{service_name}/call")
def record_call(service_name: str, is_success: bool, duration_ms: float):
    breaker = manager.get_or_create(service_name)
    if not breaker.is_configured:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Service '{service_name}' is not configured",
        )
    if not breaker.can_accept_call():
        raise HTTPException(
            status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
            detail=f"Circuit breaker for service '{service_name}' is open",
        )
    breaker.record_call(is_success=is_success, duration_ms=duration_ms)
    return {
        "service_name": service_name,
        "status": breaker.state.value,
    }


def get_app():
    return app


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8400"))
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        reload=True,
    )
