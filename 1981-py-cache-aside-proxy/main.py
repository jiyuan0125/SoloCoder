import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request, Response, HTTPException
from fastapi.responses import JSONResponse

from app.config import config, WriteStrategy
from app.cache import cache_manager, CacheManager
from app.proxy import proxy_service
from app.stats import stats_manager
from app.health import health_manager

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info(f"Starting cache-aside proxy on port {config.port}")
    logger.info(f"Backend URL: {config.backend_url}")
    logger.info(f"Default TTL: {config.default_ttl}s")
    yield
    await proxy_service.close()
    logger.info("Cache-aside proxy stopped")


app = FastAPI(
    title="Cache-Aside Proxy",
    description="FastAPI-based cache-aside proxy with TTL, LRU eviction, and breakdown protection",
    lifespan=lifespan
)


@app.get("/_admin/health")
async def admin_health():
    return {
        "status": "healthy",
        "backend": health_manager.get_status()
    }


@app.get("/_admin/stats")
async def admin_stats(namespace: str = None):
    if namespace:
        return stats_manager.get_stats(namespace)
    return stats_manager.get_all_stats()


@app.get("/_admin/cache-stats")
async def admin_cache_stats():
    return cache_manager.get_all_stats()


@app.post("/_admin/cache/clear")
async def admin_clear_cache(namespace: str = None):
    if namespace:
        await cache_manager.clear_namespace(namespace)
        logger.info(f"Cleared cache for namespace: {namespace}")
    else:
        await cache_manager.clear_all()
        logger.info("Cleared all cache")
    return {"status": "ok"}


@app.post("/_admin/stats/reset")
async def admin_reset_stats(namespace: str = None):
    stats_manager.reset(namespace)
    logger.info(f"Reset stats{f' for namespace: {namespace}' if namespace else ' all'}")
    return {"status": "ok"}


@app.post("/_admin/health/reset")
async def admin_reset_health():
    health_manager.reset()
    logger.info("Reset health status")
    return {"status": "ok"}


@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"])
async def catch_all(path: str, request: Request):
    return await proxy_service.handle_request(request)


def main():
    import uvicorn
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=config.port,
        reload=False
    )


if __name__ == "__main__":
    main()
