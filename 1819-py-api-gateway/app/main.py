import os
import httpx
from contextlib import asynccontextmanager
from typing import List, Optional
from fastapi import FastAPI, Request, HTTPException, status
from fastapi.responses import JSONResponse, Response
from fastapi.exception_handlers import http_exception_handler

from .config import get_settings, RouteConfig, MatchType, AuthType
from .router import RouteMatcher
from .auth import AuthManager
from .aggregate import AggregateRequest, aggregate_services

settings = get_settings()
router = RouteMatcher()
auth_manager = AuthManager(settings)

@asynccontextmanager
async def lifespan(app: FastAPI):
    for route in settings.default_routes:
        await router.add_route(route)
    yield

app = FastAPI(
    title="API Gateway",
    description="FastAPI-based API Gateway with routing, authentication, and aggregation",
    version="1.0.0",
    lifespan=lifespan
)

@app.exception_handler(HTTPException)
async def custom_http_exception_handler(request: Request, exc: HTTPException):
    if exc.status_code == status.HTTP_404_NOT_FOUND:
        all_routes = await router.get_all_routes()
        available_routes = []
        for r in all_routes:
            available_routes.append({
                "id": r.id,
                "path": r.path,
                "match_type": r.match_type.value,
                "auth_type": r.auth_type.value,
                "description": r.description
            })
        return JSONResponse(
            status_code=status.HTTP_404_NOT_FOUND,
            content={
                "detail": "Route not found",
                "available_routes": available_routes
            }
        )
    return await http_exception_handler(request, exc)

@app.post("/aggregate", tags=["Aggregation"])
async def aggregate_endpoint(request: AggregateRequest):
    return await aggregate_services(request, settings.aggregate_timeout)

@app.get("/routes", tags=["Route Management"])
async def list_routes():
    routes = await router.get_all_routes()
    return {
        "total": len(routes),
        "routes": [r.dict() for r in routes]
    }

@app.get("/routes/{route_id}", tags=["Route Management"])
async def get_route(route_id: str):
    route = await router.get_route(route_id)
    if not route:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Route {route_id} not found"
        )
    return route

@app.post("/routes", tags=["Route Management"])
async def create_route(route: RouteConfig):
    success = await router.add_route(route)
    if not success:
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=f"Route {route.id} already exists"
        )
    return {"message": "Route created", "route": route.dict()}

@app.put("/routes/{route_id}", tags=["Route Management"])
async def update_route(route_id: str, route: RouteConfig):
    if route.id != route_id:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail="Route ID in path does not match body"
        )
    success = await router.update_route(route)
    if not success:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Route {route_id} not found"
        )
    return {"message": "Route updated", "route": route.dict()}

@app.delete("/routes/{route_id}", tags=["Route Management"])
async def delete_route(route_id: str):
    success = await router.remove_route(route_id)
    if not success:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Route {route_id} not found"
        )
    return {"message": f"Route {route_id} marked for deletion, will be removed after active requests complete"}

async def _proxy_request(request: Request, target_url: str) -> Response:
    path = request.url.path
    query_params = request.url.query
    method = request.method
    headers = dict(request.headers)
    
    headers_to_remove = ["host", "content-length", "connection"]
    for h in headers_to_remove:
        headers.pop(h, None)
    
    full_url = target_url.rstrip("/") + path
    if query_params:
        full_url += "?" + query_params
    
    body = await request.body()
    
    async with httpx.AsyncClient() as client:
        response = await client.request(
            method=method,
            url=full_url,
            headers=headers,
            content=body,
            timeout=30.0
        )
        
        content = response.content
        response_headers = dict(response.headers)
        
        headers_to_remove = ["content-encoding", "content-length", "transfer-encoding"]
        for h in headers_to_remove:
            response_headers.pop(h, None)
        
        return Response(
            content=content,
            status_code=response.status_code,
            headers=response_headers,
            media_type=response.headers.get("content-type")
        )

@app.api_route("/{path:path}", methods=["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"])
async def gateway_handler(request: Request, path: str):
    full_path = "/" + path if not path.startswith("/") else path
    
    matched_route, route_state = await router.match_route(full_path)
    
    if not matched_route:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail="Route not found"
        )
    
    try:
        await auth_manager.authenticate(request, matched_route.auth_type)
        
        response = await _proxy_request(request, matched_route.target_url)
        return response
    finally:
        if route_state:
            await router.release_route(route_state)

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=settings.port,
        reload=False
    )
