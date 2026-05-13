import os
import json
import logging
from typing import Dict, Any
from aiohttp import web
from connection_pool import (
    FailoverConnectionPool,
    PoolConfig,
    AddressConfig,
)

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


class ConnectionPoolManager:
    def __init__(self):
        self._pools: Dict[str, FailoverConnectionPool] = {}

    def create_pool(self, config: PoolConfig) -> FailoverConnectionPool:
        if config.name in self._pools:
            raise ValueError(f"Pool '{config.name}' already exists")
        
        pool = FailoverConnectionPool(config)
        self._pools[config.name] = pool
        return pool

    def get_pool(self, name: str) -> FailoverConnectionPool:
        if name not in self._pools:
            raise KeyError(f"Pool '{name}' not found")
        return self._pools[name]

    async def delete_pool(self, name: str) -> None:
        if name not in self._pools:
            raise KeyError(f"Pool '{name}' not found")
        
        pool = self._pools.pop(name)
        await pool.close()

    def list_pools(self) -> Dict[str, FailoverConnectionPool]:
        return self._pools


pool_manager = ConnectionPoolManager()


async def handle_create_pool(request: web.Request) -> web.Response:
    try:
        data = await request.json()
    except json.JSONDecodeError:
        return web.json_response({"error": "Invalid JSON"}, status=400)

    name = data.get("name")
    addresses_data = data.get("addresses", [])
    
    if not name or not addresses_data:
        return web.json_response(
            {"error": "Missing required fields: 'name' and 'addresses'"},
            status=400
        )

    try:
        addresses = []
        for addr_data in addresses_data:
            addr = AddressConfig(
                host=addr_data["host"],
                port=addr_data["port"],
                user=addr_data["user"],
                password=addr_data["password"],
                database=addr_data["database"],
            )
            addresses.append(addr)

        config = PoolConfig(
            name=name,
            addresses=addresses,
            max_connections=data.get("max_connections", 10),
            connection_timeout=data.get("connection_timeout", 5),
            failure_threshold=data.get("failure_threshold", 3),
            cooldown_period=data.get("cooldown_period", 60),
            all_failed_wait_period=data.get("all_failed_wait_period", 30),
        )

        pool = pool_manager.create_pool(config)
        logger.info(f"Created pool '{name}' with {len(addresses)} addresses")
        
        return web.json_response({
            "message": f"Pool '{name}' created successfully",
            "status": pool.get_status()
        }, status=201)

    except KeyError as e:
        return web.json_response({"error": f"Missing address field: {e}"}, status=400)
    except ValueError as e:
        return web.json_response({"error": str(e)}, status=409)
    except Exception as e:
        logger.exception(f"Error creating pool: {e}")
        return web.json_response({"error": str(e)}, status=500)


async def handle_delete_pool(request: web.Request) -> web.Response:
    name = request.match_info["name"]
    
    try:
        await pool_manager.delete_pool(name)
        logger.info(f"Deleted pool '{name}'")
        return web.json_response({
            "message": f"Pool '{name}' deleted successfully"
        })
    except KeyError as e:
        return web.json_response({"error": str(e)}, status=404)
    except Exception as e:
        logger.exception(f"Error deleting pool: {e}")
        return web.json_response({"error": str(e)}, status=500)


async def handle_get_pool_status(request: web.Request) -> web.Response:
    name = request.match_info["name"]
    
    try:
        pool = pool_manager.get_pool(name)
        return web.json_response(pool.get_status())
    except KeyError as e:
        return web.json_response({"error": str(e)}, status=404)
    except Exception as e:
        logger.exception(f"Error getting pool status: {e}")
        return web.json_response({"error": str(e)}, status=500)


async def handle_update_pool_config(request: web.Request) -> web.Response:
    name = request.match_info["name"]
    
    try:
        data = await request.json()
    except json.JSONDecodeError:
        return web.json_response({"error": "Invalid JSON"}, status=400)

    try:
        pool = pool_manager.get_pool(name)
        
        await pool.update_config(
            max_connections=data.get("max_connections"),
            connection_timeout=data.get("connection_timeout"),
            failure_threshold=data.get("failure_threshold"),
            cooldown_period=data.get("cooldown_period"),
            all_failed_wait_period=data.get("all_failed_wait_period"),
        )
        
        logger.info(f"Updated config for pool '{name}'")
        return web.json_response({
            "message": f"Pool '{name}' config updated successfully",
            "status": pool.get_status()
        })
    except KeyError as e:
        return web.json_response({"error": str(e)}, status=404)
    except Exception as e:
        logger.exception(f"Error updating pool config: {e}")
        return web.json_response({"error": str(e)}, status=500)


async def handle_list_pools(request: web.Request) -> web.Response:
    try:
        pools = pool_manager.list_pools()
        return web.json_response({
            "pools": [
                {
                    "name": name,
                    "status": pool.get_status()
                }
                for name, pool in pools.items()
            ]
        })
    except Exception as e:
        logger.exception(f"Error listing pools: {e}")
        return web.json_response({"error": str(e)}, status=500)


async def handle_test_connection(request: web.Request) -> web.Response:
    name = request.match_info["name"]
    
    try:
        pool = pool_manager.get_pool(name)
        
        try:
            conn, addr_idx = await pool.acquire()
            try:
                async with conn.cursor() as cur:
                    await cur.execute("SELECT 1")
                    result = await cur.fetchone()
                
                return web.json_response({
                    "message": "Connection successful",
                    "result": result,
                    "address_index": addr_idx,
                })
            finally:
                await pool.release(conn, addr_idx)
        
        except RuntimeError as e:
            return web.json_response({
                "error": str(e),
                "status": pool.get_status()
            }, status=503)
            
    except KeyError as e:
        return web.json_response({"error": str(e)}, status=404)
    except Exception as e:
        logger.exception(f"Error testing connection: {e}")
        return web.json_response({"error": str(e)}, status=500)


def create_app() -> web.Application:
    app = web.Application()
    
    app.router.add_post("/pools", handle_create_pool)
    app.router.add_delete("/pools/{name}", handle_delete_pool)
    app.router.add_get("/pools/{name}/status", handle_get_pool_status)
    app.router.add_patch("/pools/{name}/config", handle_update_pool_config)
    app.router.add_get("/pools", handle_list_pools)
    app.router.add_get("/pools/{name}/test", handle_test_connection)
    
    return app


async def on_cleanup(app: web.Application) -> None:
    logger.info("Cleaning up all pools...")
    for name in list(pool_manager._pools.keys()):
        await pool_manager.delete_pool(name)
    logger.info("All pools cleaned up")


if __name__ == "__main__":
    app = create_app()
    app.on_cleanup.append(on_cleanup)
    
    port = int(os.environ.get("PORT", 8080))
    logger.info(f"Starting server on port {port}")
    web.run_app(app, port=port, host="0.0.0.0")
