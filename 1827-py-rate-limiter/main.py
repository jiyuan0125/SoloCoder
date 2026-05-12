import os
from typing import Optional
from contextlib import asynccontextmanager
from fastapi import FastAPI, HTTPException, Query
from rate_limiter import RateLimiter, RateLimitMiddleware, RateLimitConfig


rate_limiter = RateLimiter()


@asynccontextmanager
async def lifespan(app: FastAPI):
    rate_limiter.start_cleanup_task()
    yield
    rate_limiter.stop_cleanup_task()


app = FastAPI(
    title="Rate Limiter API",
    description="IP 和路由前缀双维度限流中间件",
    lifespan=lifespan
)


app.add_middleware(RateLimitMiddleware, rate_limiter=rate_limiter)


@app.get("/rate-limits")
async def get_all_configs():
    configs = await rate_limiter.get_all_configs()
    return {
        "configs": {
            prefix: {"window_size": config.window_size, "max_requests": config.max_requests}
            for prefix, config in configs.items()
        }
    }


@app.put("/rate-limits/{prefix:path}")
async def update_config(prefix: str, config: RateLimitConfig):
    if not prefix.startswith("/"):
        raise HTTPException(status_code=400, detail="Prefix must start with /")
    
    await rate_limiter.set_config(prefix, config)
    return {
        "prefix": prefix,
        "window_size": config.window_size,
        "max_requests": config.max_requests,
        "message": "Configuration updated successfully"
    }


@app.delete("/rate-limits/{prefix:path}")
async def delete_config(prefix: str):
    deleted = await rate_limiter.delete_config(prefix)
    if not deleted:
        raise HTTPException(status_code=404, detail="Configuration not found")
    return {"message": f"Configuration for '{prefix}' deleted successfully"}


@app.get("/rate-limits/usage")
async def get_usage(
    client_ip: Optional[str] = Query(None, description="按客户端 IP 过滤"),
    prefix: Optional[str] = Query(None, description="按路由前缀过滤"),
    page: int = Query(1, ge=1, description="页码"),
    page_size: int = Query(20, ge=1, le=100, description="每页大小")
):
    usage, total = await rate_limiter.get_usage(client_ip, prefix, page, page_size)
    
    return {
        "items": usage,
        "total": total,
        "page": page,
        "page_size": page_size,
        "total_pages": (total + page_size - 1) // page_size if total > 0 else 0
    }


@app.get("/api/fast/test")
async def test_fast():
    return {"message": "Fast API endpoint"}


@app.get("/api/slow/test")
async def test_slow():
    return {"message": "Slow API endpoint"}


@app.get("/")
async def root():
    return {"message": "Rate Limiter Demo API"}


if __name__ == "__main__":
    import uvicorn
    port = int(os.environ.get("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)
