import asyncio
import time
from typing import Optional

import httpx
from fastapi import FastAPI, Request, status
from fastapi.responses import JSONResponse, Response, StreamingResponse
from starlette.middleware.base import BaseHTTPMiddleware

from app.config import settings
from app.models import ColoringRule, HeaderMatch, IPRange, RouteRule
from app.routers import coloring_router, routes_router, stats_router
from app.services import coloring_service, mirror_service, routing_service, stats_service


class ColoringMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        headers = {k.lower(): v for k, v in request.headers.items()}
        cookies = dict(request.cookies)
        client_ip = request.client.host if request.client else "127.0.0.1"

        xff = request.headers.get("x-forwarded-for")
        if xff:
            client_ip = xff.split(",")[0].strip()

        tag = coloring_service.get_tag(headers, cookies, client_ip)

        request.state.color_tag = tag
        request.scope["color_tag"] = tag

        response = await call_next(request)

        response.headers["X-Color-Tag"] = tag
        return response


class RoutingMiddleware(BaseHTTPMiddleware):
    def __init__(self, app):
        super().__init__(app)
        self._client: Optional[httpx.AsyncClient] = None

    async def _get_client(self) -> httpx.AsyncClient:
        if self._client is None:
            self._client = httpx.AsyncClient(timeout=30.0)
        return self._client

    async def dispatch(self, request: Request, call_next):
        path = request.url.path
        method = request.method

        if path.startswith("/api/") or path == "/health" or path.startswith("/docs") or path.startswith("/openapi.json") or path.startswith("/redoc"):
            return await call_next(request)

        parts = path.lstrip("/").split("/", 1)
        if not parts or not parts[0]:
            return await call_next(request)

        service_name = parts[0]
        service_path = "/" + (parts[1] if len(parts) > 1 else "")

        if request.url.query:
            service_path += f"?{request.url.query}"

        tag = getattr(request.state, "color_tag", "default")

        upstream = routing_service.get_upstream(service_name, tag)
        if not upstream:
            return JSONResponse(
                status_code=status.HTTP_503_SERVICE_UNAVAILABLE,
                content={"error": f"No upstream configured for service '{service_name}' with tag '{tag}'"},
            )

        start_time = time.time()

        body = await request.body()

        asyncio.create_task(
            mirror_service.mirror_request(
                method=method,
                url=service_path,
                headers=dict(request.headers),
                body=body,
                tag=tag,
            )
        )

        try:
            client = await self._get_client()
            upstream_headers = {k: v for k, v in request.headers.items() if k.lower() != "host"}
            upstream_headers["X-Color-Tag"] = tag

            upstream_response = await client.request(
                method=method,
                url=f"{upstream}{service_path}",
                headers=upstream_headers,
                content=body,
            )

            latency_ms = (time.time() - start_time) * 1000
            is_error = upstream_response.status_code >= 500

            stats_service.record_request(service_name, tag, latency_ms, is_error)

            response_headers = dict(upstream_response.headers)
            response_headers.pop("transfer-encoding", None)
            response_headers.pop("content-encoding", None)
            response_headers["X-Upstream"] = upstream
            response_headers["X-Color-Tag"] = tag
            response_headers["X-Latency-MS"] = str(latency_ms)

            return StreamingResponse(
                content=upstream_response.aiter_bytes(),
                status_code=upstream_response.status_code,
                headers=response_headers,
            )

        except httpx.HTTPError as e:
            latency_ms = (time.time() - start_time) * 1000
            stats_service.record_request(service_name, tag, latency_ms, is_error=True)
            return JSONResponse(
                status_code=status.HTTP_502_BAD_GATEWAY,
                content={"error": f"Upstream error: {str(e)}"},
            )
        except Exception as e:
            latency_ms = (time.time() - start_time) * 1000
            stats_service.record_request(service_name, tag, latency_ms, is_error=True)
            return JSONResponse(
                status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                content={"error": str(e)},
            )


def init_default_data():
    for rule in settings.default_coloring_rules:
        coloring_service.add_rule(rule)

    for service_name, service_routes in settings.default_service_routes.items():
        for rule in service_routes.rules.values():
            routing_service.add_route_rule(service_name, rule)


def create_app() -> FastAPI:
    app = FastAPI(title="Traffic Tag Gateway", version="1.0.0")

    init_default_data()

    app.add_middleware(ColoringMiddleware)
    app.add_middleware(RoutingMiddleware)

    app.include_router(routes_router)
    app.include_router(stats_router)
    app.include_router(coloring_router)

    @app.on_event("shutdown")
    async def shutdown_event():
        await mirror_service.close()

    @app.get("/health")
    async def health() -> dict:
        return {"status": "ok"}

    return app


app = create_app()
