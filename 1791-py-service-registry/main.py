import os
import uuid
import time
import asyncio
from typing import Optional, Dict, List, Any
from datetime import datetime, timezone

from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import JSONResponse
from pydantic import BaseModel, Field
import semantic_version


class ServiceRegisterRequest(BaseModel):
    name: str
    address: str
    metadata: Dict[str, Any] = Field(default_factory=dict)
    version: Optional[str] = None


class ServiceInstance(BaseModel):
    instance_id: str
    name: str
    address: str
    metadata: Dict[str, Any]
    version: Optional[semantic_version.Version] = None
    registered_at: float
    last_heartbeat_at: float

    class Config:
        arbitrary_types_allowed = True


class InstanceStatus(str):
    HEALTHY = "healthy"
    EXPIRED = "expired"


class ServiceRegistry:
    def __init__(self, expiration_seconds: int = 30):
        self._instances: Dict[str, ServiceInstance] = {}
        self._expiration_seconds = expiration_seconds

    def register(self, name: str, address: str, metadata: Dict[str, Any], version: Optional[str] = None) -> str:
        instance_id = str(uuid.uuid4())
        now = time.time()
        
        parsed_version = None
        if version is not None:
            try:
                parsed_version = semantic_version.Version(version)
            except ValueError:
                raise HTTPException(
                    status_code=400,
                    detail=f"Invalid version format: {version}. Expected semantic version like '2.1.3'"
                )
        
        instance = ServiceInstance(
            instance_id=instance_id,
            name=name,
            address=address,
            metadata=metadata or {},
            version=parsed_version,
            registered_at=now,
            last_heartbeat_at=now
        )
        self._instances[instance_id] = instance
        return instance_id

    def heartbeat(self, instance_id: str) -> bool:
        if instance_id not in self._instances:
            return False
        self._instances[instance_id].last_heartbeat_at = time.time()
        return True

    def deregister(self, instance_id: str) -> bool:
        if instance_id not in self._instances:
            return False
        del self._instances[instance_id]
        return True

    def is_expired(self, instance: ServiceInstance) -> bool:
        return (time.time() - instance.last_heartbeat_at) > self._expiration_seconds

    def get_instance(self, instance_id: str) -> Optional[ServiceInstance]:
        return self._instances.get(instance_id)

    def get_all_instances(self) -> List[ServiceInstance]:
        return list(self._instances.values())

    def cleanup_expired(self) -> int:
        expired_ids = [
            instance_id
            for instance_id, instance in self._instances.items()
            if self.is_expired(instance)
        ]
        for instance_id in expired_ids:
            del self._instances[instance_id]
        return len(expired_ids)

    def query(
        self,
        name: Optional[str] = None,
        labels: Optional[Dict[str, str]] = None,
        min_version: Optional[str] = None,
        max_version: Optional[str] = None
    ) -> List[Dict[str, Any]]:
        if min_version is not None:
            try:
                parsed_min = semantic_version.Version(min_version)
            except ValueError:
                raise HTTPException(
                    status_code=400,
                    detail=f"Invalid min_version format: {min_version}"
                )
        else:
            parsed_min = None

        if max_version is not None:
            try:
                parsed_max = semantic_version.Version(max_version)
            except ValueError:
                raise HTTPException(
                    status_code=400,
                    detail=f"Invalid max_version format: {max_version}"
                )
        else:
            parsed_max = None

        has_version_filter = min_version is not None or max_version is not None

        results = []
        for instance in self._instances.values():
            if name is not None and instance.name != name:
                continue

            if labels:
                metadata_labels = instance.metadata.get("labels", {})
                if not isinstance(metadata_labels, dict):
                    continue
                if not all(metadata_labels.get(k) == v for k, v in labels.items()):
                    continue

            if has_version_filter:
                if instance.version is None:
                    continue
                if parsed_min is not None and instance.version < parsed_min:
                    continue
                if parsed_max is not None and instance.version >= parsed_max:
                    continue

            results.append({
                "instance_id": instance.instance_id,
                "name": instance.name,
                "address": instance.address,
                "metadata": instance.metadata,
                "version": str(instance.version) if instance.version else None
            })

        return results


app = FastAPI(title="Service Registry")
registry = ServiceRegistry(expiration_seconds=30)


@app.on_event("startup")
async def startup_event():
    asyncio.create_task(cleanup_task())


async def cleanup_task():
    while True:
        expired_count = registry.cleanup_expired()
        if expired_count > 0:
            print(f"Cleaned up {expired_count} expired instances")
        await asyncio.sleep(5)


@app.post("/register", status_code=201)
async def register(request: ServiceRegisterRequest):
    instance_id = registry.register(
        name=request.name,
        address=request.address,
        metadata=request.metadata,
        version=request.version
    )
    return {"instance_id": instance_id}


@app.post("/heartbeat/{instance_id}", status_code=200)
async def heartbeat(instance_id: str):
    if not registry.heartbeat(instance_id):
        raise HTTPException(status_code=404, detail="Instance not found")
    return {"status": "ok"}


@app.delete("/deregister/{instance_id}", status_code=204)
async def deregister(instance_id: str):
    if not registry.deregister(instance_id):
        raise HTTPException(status_code=404, detail="Instance not found")
    return JSONResponse(status_code=204, content={})


def parse_labels(labels_param: List[str]) -> Dict[str, str]:
    labels = {}
    for label in labels_param:
        if "=" not in label:
            raise HTTPException(
                status_code=400,
                detail=f"Invalid label format: {label}. Expected 'key=value'"
            )
        key, value = label.split("=", 1)
        labels[key] = value
    return labels


@app.get("/query")
async def query(
    name: Optional[str] = Query(None, description="Service name to filter"),
    label: List[str] = Query([], description="Label filter in format 'key=value'. Multiple labels must all match"),
    min_version: Optional[str] = Query(None, description="Minimum version (>=), e.g., 2.0.0"),
    max_version: Optional[str] = Query(None, description="Maximum version (<), e.g., 3.0.0")
):
    labels = parse_labels(label)
    results = registry.query(
        name=name,
        labels=labels if labels else None,
        min_version=min_version,
        max_version=max_version
    )
    return {"services": results, "count": len(results)}


@app.get("/health")
async def get_all_services_health():
    instances = registry.get_all_instances()
    
    grouped: Dict[str, List[Dict[str, Any]]] = {}
    for instance in instances:
        if instance.name not in grouped:
            grouped[instance.name] = []
        
        status = InstanceStatus.EXPIRED if registry.is_expired(instance) else InstanceStatus.HEALTHY
        grouped[instance.name].append({
            "instance_id": instance.instance_id,
            "address": instance.address,
            "status": status,
            "version": str(instance.version) if instance.version else None,
            "metadata": instance.metadata,
            "last_heartbeat_at": datetime.fromtimestamp(
                instance.last_heartbeat_at, tz=timezone.utc
            ).isoformat()
        })
    
    return {"services": grouped}


if __name__ == "__main__":
    import uvicorn
    port = int(os.environ.get("PORT", "8000"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
