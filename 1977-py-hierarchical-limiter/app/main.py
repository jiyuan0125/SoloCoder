import os
from fastapi import FastAPI, Request, HTTPException, Depends, Header
from fastapi.responses import JSONResponse
from .models import (
    LimiterConfig,
    ConfigUpdate,
    StatusResponse,
    RateLimitExceeded,
)
from .limiter import HierarchicalRateLimiter


DEFAULT_CONFIG = LimiterConfig(
    global_qps=10000,
    tenant_qps={
        "tenant-a": 1000,
        "tenant-b": 500,
    },
    api_qps={
        "/api/heavy-query": 100,
    },
    default_tenant_qps=100,
    default_api_qps=500,
    max_borrow_per_tenant=200,
)

limiter = HierarchicalRateLimiter(DEFAULT_CONFIG)

app = FastAPI(
    title="Hierarchical Rate Limiter",
    description="三级层级限流系统：全局 → 租户 → API",
    version="1.0.0",
)


@app.exception_handler(HTTPException)
async def http_exception_handler(request: Request, exc: HTTPException):
    if exc.status_code == 429:
        return JSONResponse(
            status_code=429,
            content=exc.detail,
        )
    return JSONResponse(
        status_code=exc.status_code,
        content={"detail": exc.detail},
    )


async def rate_limit_middleware(request: Request, call_next):
    path = request.url.path
    method = request.method

    if path.startswith("/limiter/") or path == "/":
        return await call_next(request)

    tenant_id = request.headers.get("X-Tenant-Id")
    if not tenant_id:
        raise HTTPException(
            status_code=400,
            detail="Missing X-Tenant-Id header",
        )

    api_path = f"{method}:{path}"
    result = await limiter.try_acquire(tenant_id, api_path)

    if not result.allowed:
        error_response = RateLimitExceeded(
            level=result.level,
            message=f"Rate limit exceeded at {result.level} level",
            remaining_quota=result.remaining,
            retry_after=result.retry_after,
        )
        response = JSONResponse(
            status_code=429,
            content=error_response.model_dump(),
        )
        response.headers["Retry-After"] = str(result.retry_after)
        return response

    response = await call_next(request)
    response.headers["X-RateLimit-Remaining"] = str(result.remaining)
    return response


app.middleware("http")(rate_limit_middleware)


@app.get("/")
async def root():
    return {
        "name": "Hierarchical Rate Limiter",
        "version": "1.0.0",
        "endpoints": {
            "status": "GET /limiter/status",
            "tenant_stats": "GET /limiter/stats/{tenant_id}",
            "config": "POST /limiter/config",
        },
    }


@app.get("/limiter/status", response_model=StatusResponse)
async def get_status():
    return StatusResponse(
        global_status=limiter.get_global_status(),
        tenants=limiter.get_all_tenants_info(),
        apis=limiter.get_all_apis_info(),
    )


@app.get("/limiter/stats/{tenant_id}")
async def get_tenant_stats(tenant_id: str):
    return limiter.get_tenant_stats(tenant_id)


@app.post("/limiter/config")
async def update_config(config: ConfigUpdate):
    limiter.update_config(config.model_dump())
    return {
        "message": "Config updated",
        "current_config": limiter.config.model_dump(),
    }


@app.get("/api/users")
async def get_users(x_tenant_id: str = Header(...)):
    return {
        "message": "Success",
        "tenant_id": x_tenant_id,
        "data": ["user1", "user2", "user3"],
    }


@app.get("/api/orders")
async def get_orders(x_tenant_id: str = Header(...)):
    return {
        "message": "Success",
        "tenant_id": x_tenant_id,
        "data": ["order1", "order2"],
    }


@app.get("/api/heavy-query")
async def heavy_query(x_tenant_id: str = Header(...)):
    return {
        "message": "Success",
        "tenant_id": x_tenant_id,
        "data": "Heavy processing result",
    }


def run():
    port = int(os.environ.get("PORT", 8000))
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=port)


if __name__ == "__main__":
    run()
