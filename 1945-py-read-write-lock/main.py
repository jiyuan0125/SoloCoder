import os

from fastapi import FastAPI, HTTPException, status
from pydantic import BaseModel

from lock_manager import LockType, lock_manager


app = FastAPI(title="Distributed Read-Write Lock Service", version="1.0.0")


class AcquireRequest(BaseModel):
    resource_name: str
    lock_type: LockType
    client_id: str
    timeout: float = 10.0


class ReleaseRequest(BaseModel):
    resource_name: str
    client_id: str


class DowngradeRequest(BaseModel):
    resource_name: str
    client_id: str


class LockAcquiredResponse(BaseModel):
    success: bool
    message: str


class LockReleasedResponse(BaseModel):
    success: bool
    message: str


class LockDowngradedResponse(BaseModel):
    success: bool
    message: str


@app.post("/locks/acquire", response_model=LockAcquiredResponse)
async def acquire_lock(request: AcquireRequest):
    success = await lock_manager.acquire(
        resource_name=request.resource_name,
        lock_type=request.lock_type,
        client_id=request.client_id,
        timeout=request.timeout,
    )
    if not success:
        raise HTTPException(
            status_code=status.HTTP_408_REQUEST_TIMEOUT,
            detail=f"Timeout waiting for {request.lock_type.value} lock on resource '{request.resource_name}'",
        )
    return LockAcquiredResponse(
        success=True,
        message=f"{request.lock_type.value.capitalize()} lock acquired on resource '{request.resource_name}'",
    )


@app.post("/locks/release", response_model=LockReleasedResponse)
async def release_lock(request: ReleaseRequest):
    success = await lock_manager.release(
        resource_name=request.resource_name,
        client_id=request.client_id,
    )
    if not success:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"No lock held by client '{request.client_id}' on resource '{request.resource_name}'",
        )
    return LockReleasedResponse(
        success=True,
        message=f"Lock released on resource '{request.resource_name}'",
    )


@app.post("/locks/downgrade", response_model=LockDowngradedResponse)
async def downgrade_lock(request: DowngradeRequest):
    success = await lock_manager.downgrade(
        resource_name=request.resource_name,
        client_id=request.client_id,
    )
    if not success:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Cannot downgrade lock for client '{request.client_id}' on resource '{request.resource_name}'. Make sure you hold a write lock.",
        )
    return LockDowngradedResponse(
        success=True,
        message=f"Write lock downgraded to read lock on resource '{request.resource_name}'",
    )


@app.get("/locks/{resource}")
async def get_lock_status(resource: str):
    return lock_manager.get_status(resource)


@app.get("/health")
async def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
