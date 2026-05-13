import os
from contextlib import asynccontextmanager
from typing import Dict, List, Optional

from fastapi import FastAPI, HTTPException, Query
from pydantic import BaseModel

from models import (
    CircuitBreakerState,
    DegradationState,
    DependencyConfig,
    DependencyRecord,
    DependencyGraph,
)
from circuit_breaker import CircuitBreakerManager


cb_manager = CircuitBreakerManager()


@asynccontextmanager
async def lifespan(app: FastAPI):
    yield


app = FastAPI(title="Circuit Breaker Coordinator", lifespan=lifespan)


class FailureRecord(BaseModel):
    service_name: str


class SuccessRecord(BaseModel):
    service_name: str


@app.post("/dependencies", response_model=DependencyRecord)
async def create_dependency(config: DependencyConfig):
    if not config.upstream:
        raise HTTPException(status_code=400, detail="upstream cannot be empty")
    if not config.downstream:
        raise HTTPException(status_code=400, detail="downstream list cannot be empty")
    if config.degraded_timeout <= 0:
        raise HTTPException(
            status_code=400, detail="degraded_timeout must be positive"
        )

    record = cb_manager.add_dependency(config)
    return record


@app.delete("/dependencies/{dep_id}")
async def delete_dependency(dep_id: str):
    success = cb_manager.remove_dependency(dep_id)
    if not success:
        raise HTTPException(status_code=404, detail=f"Dependency {dep_id} not found")
    return {"message": "Dependency removed successfully", "id": dep_id}


@app.get("/status")
async def get_all_status():
    services_status = cb_manager.get_all_services_status()
    dependency_graph = cb_manager.get_dependency_graph()

    services_dict = {}
    for name, status in services_status.items():
        services_dict[name] = status.model_dump()

    return {
        "services": services_dict,
        "dependency_graph": dependency_graph,
    }


@app.get("/status/{service_name}")
async def get_service_status(service_name: str):
    status = cb_manager.get_service_status(service_name)
    if status is None:
        raise HTTPException(
            status_code=404, detail=f"Service {service_name} not found"
        )
    return status.model_dump()


@app.post("/record/failure")
async def record_failure(record: FailureRecord):
    await cb_manager.record_failure(record.service_name)
    status = cb_manager.get_service_status(record.service_name)
    return {
        "service": record.service_name,
        "circuit_breaker_state": status.circuit_breaker_state,
        "degradation_state": status.degradation_state,
        "failure_count": status.failure_count,
    }


@app.post("/record/success")
async def record_success(record: SuccessRecord):
    await cb_manager.record_success(record.service_name)
    status = cb_manager.get_service_status(record.service_name)
    return {
        "service": record.service_name,
        "circuit_breaker_state": status.circuit_breaker_state,
        "degradation_state": status.degradation_state,
    }


@app.get("/can-call/{service_name}")
async def can_call_service(service_name: str):
    can_call = cb_manager.can_call(service_name)
    current_timeout = cb_manager.get_current_timeout(service_name)
    return {
        "service": service_name,
        "can_call": can_call,
        "current_timeout_seconds": current_timeout,
    }


if __name__ == "__main__":
    import uvicorn

    port = int(os.environ.get("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
