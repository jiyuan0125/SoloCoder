from contextlib import asynccontextmanager
from typing import AsyncGenerator

from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse

from shared import protocols
from server.app.exceptions import PackingSystemException
from server.app.routers import orders, packages, box_types, templates
from server.app.services.box_type_service import initialize_default_box_types


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    initialize_default_box_types()
    yield


app = FastAPI(
    title="装箱管理API",
    description="仓库出库装箱管理系统 - 跨系统多模块架构",
    version="0.1.0",
    lifespan=lifespan,
)


@app.exception_handler(PackingSystemException)
async def packing_exception_handler(
    request: Request, exc: PackingSystemException
) -> JSONResponse:
    return JSONResponse(
        status_code=exc.status_code,
        content={
            "success": False,
            "error_code": exc.error_code,
            "message": exc.message,
        },
    )


app.include_router(orders.router, tags=["订单管理"])
app.include_router(packages.router, tags=["包裹管理"])
app.include_router(box_types.router, tags=["箱型管理"])
app.include_router(templates.router, tags=["装箱模板"])


@app.get("/")
async def root() -> dict[str, str]:
    return {
        "name": "装箱管理API",
        "version": "0.1.0",
        "docs": "/docs",
    }


@app.get("/health")
async def health_check() -> dict[str, str]:
    return {"status": "healthy"}
