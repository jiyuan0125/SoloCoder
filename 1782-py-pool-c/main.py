import asyncio
import os
import json
import logging
from aiohttp import web
from pool import PoolConfig, ConnectionPool, PoolManager, PoolState

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

pool_manager = PoolManager()

async def handle_create_pool(request):
    try:
        data = await request.json()
    except json.JSONDecodeError:
        return web.json_response({"error": "Invalid JSON"}, status=400)
    
    try:
        config = PoolConfig(
            name=data["name"],
            target_url=data["target_url"],
            max_connections=data.get("max_connections", 10),
            wait_timeout=data.get("wait_timeout", 30.0),
            create_retry_count=data.get("create_retry_count", 3),
            create_retry_delay=data.get("create_retry_delay", 1.0),
            leak_timeout=data.get("leak_timeout", 60.0),
            recovery_interval=data.get("recovery_interval", 30.0)
        )
    except (KeyError, ValueError) as e:
        return web.json_response({"error": str(e)}, status=400)
    
    try:
        await pool_manager.create_pool(config)
    except ValueError as e:
        return web.json_response({"error": str(e)}, status=400)
    
    logger.info(f"Created pool: {config.name}")
    return web.json_response({"status": "ok", "pool": config.name}, status=201)

async def handle_list_pools(request):
    pools = pool_manager.get_all_pools()
    overview = [
        {
            "name": name,
            "state": pool.state.value,
            "max_connections": pool.config.max_connections,
            "idle_connections": pool.idle_count,
            "leased_connections": pool.leased_count
        }
        for name, pool in pools.items()
    ]
    return web.json_response({"pools": overview})

async def handle_pool_stats(request):
    pool_name = request.match_info["name"]
    pool = pool_manager.get_pool(pool_name)
    
    if not pool:
        return web.json_response({"error": f"Pool '{pool_name}' not found"}, status=404)
    
    return web.json_response(pool.get_stats())

async def handle_delete_pool(request):
    pool_name = request.match_info["name"]
    pool = pool_manager.get_pool(pool_name)
    
    if not pool:
        return web.json_response({"error": f"Pool '{pool_name}' not found"}, status=404)
    
    await pool_manager.remove_pool(pool_name)
    logger.info(f"Deleted pool: {pool_name}")
    return web.json_response({"status": "ok"})

async def handle_acquire_connection(request):
    pool_name = request.match_info["name"]
    pool = pool_manager.get_pool(pool_name)
    
    if not pool:
        return web.json_response({"error": f"Pool '{pool_name}' not found"}, status=404)
    
    try:
        data = await request.json()
    except json.JSONDecodeError:
        data = {}
    
    wait_timeout = data.get("wait_timeout")
    caller_info = data.get("caller_info", f"API request from {request.remote}")
    
    try:
        conn_id = await pool.acquire(wait_timeout=wait_timeout, caller_info=caller_info)
        return web.json_response({"connection_id": conn_id, "status": "ok"})
    except RuntimeError as e:
        error_msg = str(e)
        if "degraded" in error_msg.lower() or "backend unreachable" in error_msg.lower():
            return web.json_response({
                "error": "后端不可达",
                "details": error_msg,
                "suggestion": "请使用降级逻辑"
            }, status=503)
        elif "full" in error_msg.lower() or "timeout" in error_msg.lower():
            return web.json_response({"error": error_msg}, status=503)
        else:
            return web.json_response({"error": error_msg}, status=500)

async def handle_release_connection(request):
    pool_name = request.match_info["name"]
    conn_id = request.match_info["conn_id"]
    pool = pool_manager.get_pool(pool_name)
    
    if not pool:
        return web.json_response({"error": f"Pool '{pool_name}' not found"}, status=404)
    
    await pool.release(conn_id)
    return web.json_response({"status": "ok"})

async def handle_health(request):
    return web.json_response({"status": "ok"})

async def on_shutdown(app):
    logger.info("Shutting down server...")
    await pool_manager.close_all()
    logger.info("Server shutdown complete")

def create_app():
    app = web.Application()
    app.router.add_post("/api/pools", handle_create_pool)
    app.router.add_get("/api/pools", handle_list_pools)
    app.router.add_get("/api/pools/{name}/stats", handle_pool_stats)
    app.router.add_delete("/api/pools/{name}", handle_delete_pool)
    app.router.add_post("/api/pools/{name}/acquire", handle_acquire_connection)
    app.router.add_post("/api/pools/{name}/release/{conn_id}", handle_release_connection)
    app.router.add_get("/health", handle_health)
    app.on_shutdown.append(on_shutdown)
    return app

def main():
    port = int(os.environ.get("PORT", 8080))
    logger.info(f"Starting connection pool server on port {port}")
    
    app = create_app()
    web.run_app(app, host="0.0.0.0", port=port)

if __name__ == "__main__":
    main()
