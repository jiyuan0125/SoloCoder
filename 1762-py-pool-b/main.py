import os
from contextlib import asynccontextmanager
from typing import Any, Dict, List, Optional

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from db_pool import (
    PoolConfig,
    PoolFullStrategy,
    PoolFullError,
    PoolManager,
    PoolStats,
    PoolTimeoutError,
)


pool_manager = PoolManager()


class CreatePoolRequest(BaseModel):
    name: str
    db_url: str
    min_size: int = 1
    max_size: int = 10
    max_idle_time: float = 300.0
    pool_full_strategy: str = "wait"
    wait_timeout: float = 10.0
    health_check_interval: float = 60.0


class PoolConfigResponse(BaseModel):
    name: str
    db_url: str
    min_size: int
    max_size: int
    max_idle_time: float
    pool_full_strategy: str
    wait_timeout: float
    health_check_interval: float


class PoolStatsResponse(BaseModel):
    name: str
    total_connections: int
    idle_connections: int
    in_use_connections: int
    min_size: int
    max_size: int
    connections_created: int
    connections_destroyed: int
    borrows: int
    borrow_timeouts: int
    health_check_failures: int
    max_idle_time: float
    last_health_check_at: Optional[float]


class QueryRequest(BaseModel):
    pool_name: str
    query: str
    params: List[Any] = []


def stats_to_response(stats: PoolStats) -> PoolStatsResponse:
    return PoolStatsResponse(
        name=stats.name,
        total_connections=stats.total_connections,
        idle_connections=stats.idle_connections,
        in_use_connections=stats.in_use_connections,
        min_size=stats.min_size,
        max_size=stats.max_size,
        connections_created=stats.connections_created,
        connections_destroyed=stats.connections_destroyed,
        borrows=stats.borrows,
        borrow_timeouts=stats.borrow_timeouts,
        health_check_failures=stats.health_check_failures,
        max_idle_time=stats.max_idle_time,
        last_health_check_at=stats.last_health_check_at,
    )


@asynccontextmanager
async def lifespan(app: FastAPI):
    default_pools = [
        {
            "name": "primary",
            "db_url": "primary.db",
            "min_size": 2,
            "max_size": 10,
            "pool_full_strategy": PoolFullStrategy.WAIT,
        },
        {
            "name": "cache",
            "db_url": "cache.db",
            "min_size": 1,
            "max_size": 5,
            "max_idle_time": 60.0,
            "pool_full_strategy": PoolFullStrategy.REJECT,
        },
    ]
    for config in default_pools:
        await pool_manager.create_pool(
            PoolConfig(**config)
        )
    yield
    await pool_manager.close_all()


app = FastAPI(
    title="Database Connection Pool Service",
    description="FastAPI service with multiple named database connection pools",
    version="1.0.0",
    lifespan=lifespan,
)


@app.get("/", response_model=Dict[str, Any])
async def root():
    return {
        "service": "Database Connection Pool Service",
        "version": "1.0.0",
        "pools": pool_manager.list_pools(),
    }


@app.get("/pools", response_model=List[PoolStatsResponse])
async def list_pools():
    stats = pool_manager.get_all_stats()
    return [stats_to_response(s) for s in stats.values()]


@app.post("/pools", response_model=PoolStatsResponse)
async def create_pool(request: CreatePoolRequest):
    try:
        strategy = PoolFullStrategy(request.pool_full_strategy.lower())
    except ValueError:
        raise HTTPException(
            status_code=400,
            detail=f"Invalid pool_full_strategy. Use 'wait' or 'reject'"
        )
    config = PoolConfig(
        name=request.name,
        db_url=request.db_url,
        min_size=request.min_size,
        max_size=request.max_size,
        max_idle_time=request.max_idle_time,
        pool_full_strategy=strategy,
        wait_timeout=request.wait_timeout,
        health_check_interval=request.health_check_interval,
    )
    try:
        await pool_manager.create_pool(config)
    except ValueError as e:
        raise HTTPException(status_code=409, detail=str(e))
    return stats_to_response(pool_manager.get_pool(request.name).get_stats())


@app.get("/pools/{pool_name}", response_model=PoolStatsResponse)
async def get_pool_stats(pool_name: str):
    pool = pool_manager.get_pool(pool_name)
    if not pool:
        raise HTTPException(status_code=404, detail=f"Pool '{pool_name}' not found")
    return stats_to_response(pool.get_stats())


@app.delete("/pools/{pool_name}")
async def remove_pool(pool_name: str):
    pool = pool_manager.get_pool(pool_name)
    if not pool:
        raise HTTPException(status_code=404, detail=f"Pool '{pool_name}' not found")
    await pool_manager.remove_pool(pool_name)
    return {"message": f"Pool '{pool_name}' removed successfully"}


@app.post("/pools/{pool_name}/query")
async def execute_query(pool_name: str, request: QueryRequest):
    pool = pool_manager.get_pool(pool_name)
    if not pool:
        raise HTTPException(status_code=404, detail=f"Pool '{pool_name}' not found")
    pooled_conn = None
    try:
        pooled_conn = await pool.borrow()
        connection = pooled_conn._connection
        async with connection.execute(request.query, request.params) as cursor:
            try:
                results = await cursor.fetchall()
                columns = [desc[0] for desc in cursor.description] if cursor.description else []
                rows = [dict(zip(columns, row)) for row in results]
            except Exception:
                rows = []
                columns = []
        await connection.commit()
        return {
            "pool_name": pool_name,
            "query": request.query,
            "rows": rows,
            "columns": columns,
        }
    except PoolTimeoutError as e:
        raise HTTPException(status_code=503, detail=str(e))
    except PoolFullError as e:
        raise HTTPException(status_code=503, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Query error: {str(e)}")
    finally:
        if pooled_conn:
            await pool.release(pooled_conn)


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)
