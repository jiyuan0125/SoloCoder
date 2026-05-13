import asyncio
import json
import logging
from typing import Any, Dict, Optional

from aiohttp import web

from .config import DatasourceConfig, PoolConfig, ProtocolType
from .manager import DatasourceManager, DatasourceNotFoundError, DatasourceAlreadyExistsError
from .pool import PoolAcquireTimeoutError, PoolDrainingError

logger = logging.getLogger(__name__)


def json_response(data: Dict[str, Any], status: int = 200) -> web.Response:
    return web.json_response(data, status=status, dumps=lambda obj: json.dumps(obj, ensure_ascii=False))


def error_response(message: str, status: int = 500) -> web.Response:
    return json_response({"error": message}, status=status)


def parse_protocol(value: str) -> ProtocolType:
    try:
        return ProtocolType(value.lower())
    except ValueError:
        raise ValueError(f"Invalid protocol: {value}")


def parse_pool_config(body: Dict[str, Any]) -> PoolConfig:
    config = PoolConfig()
    if "max_size" in body:
        config.max_size = int(body["max_size"])
    if "min_idle" in body:
        config.min_idle = int(body["min_idle"])
    if "acquire_timeout" in body:
        config.acquire_timeout = float(body["acquire_timeout"])
    if "idle_timeout" in body:
        config.idle_timeout = float(body["idle_timeout"])
    if "health_check_interval" in body:
        config.health_check_interval = float(body["health_check_interval"])
    if "health_check_fail_threshold" in body:
        config.health_check_fail_threshold = int(body["health_check_fail_threshold"])
    if "leak_detection_threshold" in body:
        config.leak_detection_threshold = float(body["leak_detection_threshold"])
    if "refill_rate_limit" in body:
        config.refill_rate_limit = int(body["refill_rate_limit"])
    config.validate()
    return config


def parse_datasource_config(body: Dict[str, Any]) -> DatasourceConfig:
    if "name" not in body:
        raise ValueError("Field 'name' is required")
    if "protocol" not in body:
        raise ValueError("Field 'protocol' is required")
    if "host" not in body:
        raise ValueError("Field 'host' is required")
    if "port" not in body:
        raise ValueError("Field 'port' is required")

    name = body["name"]
    protocol = parse_protocol(body["protocol"])
    host = body["host"]
    port = int(body["port"])

    pool_config = parse_pool_config(body.get("pool_config", {}))

    config = DatasourceConfig(
        name=name,
        protocol=protocol,
        host=host,
        port=port,
        pool_config=pool_config,
    )
    config.validate()
    return config


class ApiHandler:
    def __init__(self, manager: DatasourceManager):
        self._manager: DatasourceManager = manager

    async def health(self, request: web.Request) -> web.Response:
        return json_response({"status": "ok"})

    async def list_pools(self, request: web.Request) -> web.Response:
        pools = await self._manager.list_pools()
        return json_response({"pools": pools})

    async def get_stats(self, request: web.Request) -> web.Response:
        name: Optional[str] = request.match_info.get("name")
        try:
            stats = await self._manager.get_stats(name)
            return json_response({"stats": stats})
        except DatasourceNotFoundError as e:
            return error_response(str(e), status=404)

    async def register(self, request: web.Request) -> web.Response:
        try:
            body = await request.json()
        except json.JSONDecodeError:
            return error_response("Invalid JSON", status=400)

        try:
            config = parse_datasource_config(body)
        except (ValueError, TypeError) as e:
            return error_response(f"Invalid configuration: {e}", status=400)

        try:
            await self._manager.register(config)
        except DatasourceAlreadyExistsError as e:
            return error_response(str(e), status=409)

        return json_response(
            {
                "message": f"Datasource '{config.name}' registered",
                "config": {
                    "name": config.name,
                    "protocol": config.protocol.value,
                    "host": config.host,
                    "port": config.port,
                    "pool_config": {
                        "max_size": config.pool_config.max_size,
                        "min_idle": config.pool_config.min_idle,
                        "acquire_timeout": config.pool_config.acquire_timeout,
                    },
                },
            },
            status=201,
        )

    async def unregister(self, request: web.Request) -> web.Response:
        name = request.match_info["name"]
        wait = request.query.get("wait", "true").lower() == "true"

        try:
            await self._manager.unregister(name, wait=wait)
        except DatasourceNotFoundError as e:
            return error_response(str(e), status=404)

        return json_response({"message": f"Datasource '{name}' unregistered"})

    async def update_config(self, request: web.Request) -> web.Response:
        name = request.match_info["name"]

        try:
            body = await request.json()
        except json.JSONDecodeError:
            return error_response("Invalid JSON", status=400)

        try:
            new_config = parse_pool_config(body)
        except (ValueError, TypeError) as e:
            return error_response(f"Invalid configuration: {e}", status=400)

        try:
            await self._manager.update_pool_config(name, new_config)
        except DatasourceNotFoundError as e:
            return error_response(str(e), status=404)

        return json_response({"message": f"Config updated for '{name}'", "new_config": {
            "max_size": new_config.max_size,
            "min_idle": new_config.min_idle,
            "acquire_timeout": new_config.acquire_timeout,
        }})

    async def acquire(self, request: web.Request) -> web.Response:
        name = request.match_info["name"]

        try:
            conn = await self._manager.acquire_connection(name)
        except DatasourceNotFoundError:
            return error_response(f"Datasource '{name}' not found", status=404)
        except PoolAcquireTimeoutError as e:
            return error_response(str(e), status=504)
        except PoolDrainingError:
            return error_response(f"Datasource '{name}' is currently being drained", status=503)
        except Exception as e:
            logger.exception("Error acquiring connection")
            return error_response(f"Internal error: {e}", status=500)

        await self._manager.release_connection(name, conn)

        return json_response({
            "message": f"Connection acquired and released successfully from '{name}'",
            "connection_id": conn.id,
        })


def create_app(manager: DatasourceManager) -> web.Application:
    handler = ApiHandler(manager)
    app = web.Application()

    app.router.add_get("/health", handler.health)
    app.router.add_get("/pools", handler.list_pools)
    app.router.add_get("/stats", handler.get_stats)
    app.router.add_get("/stats/{name}", handler.get_stats)
    app.router.add_post("/datasources", handler.register)
    app.router.add_delete("/datasources/{name}", handler.unregister)
    app.router.add_patch("/datasources/{name}/config", handler.update_config)
    app.router.add_post("/datasources/{name}/acquire", handler.acquire)

    return app
