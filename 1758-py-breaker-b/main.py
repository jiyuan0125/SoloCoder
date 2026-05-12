import asyncio
import json
import logging
import os
from datetime import datetime
from typing import Optional

import aiohttp
from aiohttp import web
from dotenv import load_dotenv

from circuit_breaker import (
    CircuitBreakerManager,
    CircuitBreakerOpenError,
    CircuitState,
    ServiceStatus,
    StateChangeRecord,
)
from notifier import Notifier


load_dotenv()

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


def env_int(key: str, default: int) -> int:
    value = os.getenv(key)
    return int(value) if value and value.isdigit() else default


def env_float(key: str, default: float) -> float:
    value = os.getenv(key)
    try:
        return float(value) if value else default
    except ValueError:
        return default


breaker_manager: Optional[CircuitBreakerManager] = None
notifier: Optional[Notifier] = None
http_session: Optional[aiohttp.ClientSession] = None


def status_to_dict(status: ServiceStatus) -> dict:
    return {
        "service": status.service,
        "state": status.state.value,
        "failure_count": status.failure_count,
        "success_count": status.success_count,
        "last_failure_time": (
            datetime.fromtimestamp(status.last_failure_time).isoformat()
            if status.last_failure_time
            else None
        ),
        "last_state_change": status.last_state_change,
    }


def record_to_dict(record: StateChangeRecord) -> dict:
    return {
        "service": record.service,
        "from_state": record.from_state.value,
        "to_state": record.to_state.value,
        "reason": record.reason,
        "timestamp": record.timestamp,
    }


async def on_state_change_wrapper(record: StateChangeRecord):
    if notifier:
        await notifier.notify(record)


def patch_manager_callback(manager: CircuitBreakerManager):
    original_on_state_change = manager._on_state_change

    def new_on_state_change(record: StateChangeRecord):
        asyncio.create_task(on_state_change_wrapper(record))
        original_on_state_change(record)

    manager._on_state_change = new_on_state_change


async def handle_status_all(request: web.Request) -> web.Response:
    statuses = breaker_manager.get_all_statuses()
    data = {
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "services": [status_to_dict(s) for s in statuses],
    }
    return web.json_response(data)


async def handle_status_service(request: web.Request) -> web.Response:
    service = request.match_info["service"]
    breaker = breaker_manager.get_breaker(service)
    status = breaker.get_status()
    data = {
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "service": status_to_dict(status),
    }
    return web.json_response(data)


async def handle_history_all(request: web.Request) -> web.Response:
    records = breaker_manager.get_history()
    data = {
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "count": len(records),
        "history": [record_to_dict(r) for r in records],
    }
    return web.json_response(data)


async def handle_history_service(request: web.Request) -> web.Response:
    service = request.match_info["service"]
    records = breaker_manager.get_history(service)
    data = {
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "service": service,
        "count": len(records),
        "history": [record_to_dict(r) for r in records],
    }
    return web.json_response(data)


async def handle_config(request: web.Request) -> web.Response:
    data = {
        "timestamp": datetime.utcnow().isoformat() + "Z",
        "config": {
            "failure_threshold": breaker_manager.failure_threshold,
            "recovery_timeout": breaker_manager.recovery_timeout,
            "success_threshold": breaker_manager.success_threshold,
            "max_history": breaker_manager.max_history,
        },
    }
    return web.json_response(data)


async def handle_proxy_call(request: web.Request) -> web.Response:
    service = request.match_info["service"]
    try:
        body = await request.json()
    except Exception:
        body = {}

    url = body.get("url")
    method = (body.get("method", "GET") or "GET").upper()
    headers = body.get("headers", {})
    payload = body.get("body")
    fallback = body.get("fallback")

    if not url:
        return web.json_response({"error": "缺少 'url' 参数"}, status=400)

    async def make_request():
        nonlocal payload
        if http_session is None or http_session.closed:
            raise RuntimeError("HTTP 会话未初始化")

        kwargs = {"headers": headers}
        if payload is not None and method in ("POST", "PUT", "PATCH"):
            kwargs["json"] = payload

        async with http_session.request(method, url, **kwargs) as resp:
            resp_body = await resp.text()
            if resp.status >= 400:
                raise RuntimeError(f"下游服务错误: {resp.status} - {resp_body[:200]}")
            try:
                return json.loads(resp_body)
            except Exception:
                return resp_body

    try:
        result = await breaker_manager.call(
            service,
            make_request,
            fallback=fallback,
        )
        return web.json_response({
            "service": service,
            "status": "success",
            "result": result,
        })
    except CircuitBreakerOpenError as e:
        return web.json_response({
            "service": service,
            "status": "circuit_open",
            "error": str(e),
        }, status=503)
    except Exception as e:
        return web.json_response({
            "service": service,
            "status": "error",
            "error": str(e),
        }, status=500)


async def handle_health(request: web.Request) -> web.Response:
    return web.json_response({
        "status": "ok",
        "timestamp": datetime.utcnow().isoformat() + "Z",
    })


async def app_factory() -> web.Application:
    global breaker_manager, notifier, http_session

    port = env_int("PORT", 8080)
    failure_threshold = env_int("FAILURE_THRESHOLD", 5)
    recovery_timeout = env_float("RECOVERY_TIMEOUT", 30.0)
    success_threshold = env_int("SUCCESS_THRESHOLD", 3)
    max_history = env_int("MAX_HISTORY", 100)
    webhook_url = os.getenv("WEBHOOK_URL")

    breaker_manager = CircuitBreakerManager(
        failure_threshold=failure_threshold,
        recovery_timeout=recovery_timeout,
        success_threshold=success_threshold,
        max_history=max_history,
    )
    patch_manager_callback(breaker_manager)

    notifier = Notifier(webhook_url=webhook_url)

    http_session = aiohttp.ClientSession()

    logger.info(f"熔断器服务配置:")
    logger.info(f"  端口: {port}")
    logger.info(f"  失败阈值: {failure_threshold}")
    logger.info(f"  恢复超时: {recovery_timeout}秒")
    logger.info(f"  成功阈值: {success_threshold}")
    logger.info(f"  最大历史记录: {max_history}")
    logger.info(f"  Webhook: {'已配置' if webhook_url else '未配置'}")

    app = web.Application()

    app.router.add_get("/health", handle_health)
    app.router.add_get("/config", handle_config)

    app.router.add_get("/status", handle_status_all)
    app.router.add_get("/status/{service}", handle_status_service)

    app.router.add_get("/history", handle_history_all)
    app.router.add_get("/history/{service}", handle_history_service)

    app.router.add_post("/call/{service}", handle_proxy_call)

    async def on_shutdown(app):
        if http_session and not http_session.closed:
            await http_session.close()
        if notifier:
            await notifier.close()

    app.on_shutdown.append(on_shutdown)

    return app


def main():
    port = env_int("PORT", 8080)
    web.run_app(app_factory(), port=port)


if __name__ == "__main__":
    main()
