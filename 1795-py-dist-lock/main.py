import os
import asyncio
from contextlib import asynccontextmanager
from typing import Optional

from fastapi import FastAPI, HTTPException, Response
from pydantic import BaseModel

from lock_manager import lock_manager


class AcquireRequest(BaseModel):
    name: str
    ttl: int = 30


class ReleaseRequest(BaseModel):
    name: str
    lock_id: str


class RenewRequest(BaseModel):
    name: str
    lock_id: str
    extend_ttl: int = 30


async def cleanup_task():
    while True:
        await lock_manager.cleanup_expired()
        await asyncio.sleep(1)


@asynccontextmanager
async def lifespan(app: FastAPI):
    task = asyncio.create_task(cleanup_task())
    yield
    task.cancel()
    try:
        await task
    except asyncio.CancelledError:
        pass


app = FastAPI(lifespan=lifespan)


@app.post("/locks/acquire")
async def acquire_lock(request: AcquireRequest):
    if request.ttl <= 0:
        raise HTTPException(status_code=400, detail="TTL must be positive")
    
    waiter = await lock_manager.acquire(request.name, request.ttl)
    result = await waiter.future
    
    return {
        "lock_id": result["lock_id"],
        "expires_at": result["expires_at"],
        "name": request.name
    }


@app.post("/locks/release")
async def release_lock(request: ReleaseRequest):
    result = await lock_manager.release(request.name, request.lock_id)
    
    if result == 403:
        raise HTTPException(status_code=403, detail="Not owner of the lock")
    elif result == 404:
        raise HTTPException(status_code=404, detail="锁已过期")
    
    return {"status": "released", "name": request.name}


@app.post("/locks/renew")
async def renew_lock(request: RenewRequest):
    if request.extend_ttl <= 0:
        raise HTTPException(status_code=400, detail="Extend TTL must be positive")
    
    expires_at = await lock_manager.renew(
        request.name, 
        request.lock_id, 
        request.extend_ttl
    )
    
    if expires_at is None:
        raise HTTPException(
            status_code=404, 
            detail="Lock not found, already expired, or exceed max extension limit"
        )
    
    return {
        "name": request.name,
        "lock_id": request.lock_id,
        "expires_at": expires_at
    }


@app.get("/locks")
async def list_locks():
    locks = lock_manager.get_all_locks()
    return {"locks": locks}


@app.post("/locks/force-release/{name}")
async def force_release_lock(name: str):
    success = await lock_manager.force_release(name)
    
    if not success:
        raise HTTPException(status_code=404, detail="Lock not found")
    
    return {"status": "force_released", "name": name}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)