import os
from contextlib import asynccontextmanager
from typing import Dict, Any

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from pool_manager import PoolManager
from pool_config import PoolConfig, PoolStats


pool_manager = PoolManager()


@asynccontextmanager
async def lifespan(app: FastAPI):
    await pool_manager.create_pool(
        name="default",
        config=PoolConfig(
            max_connections=10,
            idle_timeout=300,
            wait_timeout=30,
            min_usage_count=0,
            min_connections=1,
            idle_recycle_interval=60
        )
    )
    
    yield
    
    await pool_manager.close_all()


app = FastAPI(title="Connection Pool Manager", lifespan=lifespan)


class PoolStatsResponse(BaseModel):
    active_connections: int
    idle_connections: int
    waiting_requests: int
    total_created: int
    total_destroyed: int


class AllPoolsStatsResponse(BaseModel):
    pools: Dict[str, PoolStatsResponse]


@app.get("/health", response_model=Dict[str, str])
async def health():
    return {"status": "healthy"}


@app.get("/health/pools", response_model=AllPoolsStatsResponse)
async def get_all_pool_stats():
    stats = pool_manager.get_all_stats()
    return AllPoolsStatsResponse(
        pools={
            name: PoolStatsResponse(
                active_connections=s.active_connections,
                idle_connections=s.idle_connections,
                waiting_requests=s.waiting_requests,
                total_created=s.total_created,
                total_destroyed=s.total_destroyed
            )
            for name, s in stats.items()
        }
    )


@app.get("/health/pools/{pool_name}", response_model=PoolStatsResponse)
async def get_pool_stats(pool_name: str):
    pool = pool_manager.get_pool(pool_name)
    if not pool:
        raise HTTPException(status_code=404, detail=f"Pool '{pool_name}' not found")
    
    stats = pool.get_stats()
    return PoolStatsResponse(
        active_connections=stats.active_connections,
        idle_connections=stats.idle_connections,
        waiting_requests=stats.waiting_requests,
        total_created=stats.total_created,
        total_destroyed=stats.total_destroyed
    )


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)
