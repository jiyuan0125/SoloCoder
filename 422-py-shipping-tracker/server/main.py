from contextlib import asynccontextmanager
from typing import Any, AsyncIterator

from fastapi import FastAPI, Request, status
from fastapi.encoders import jsonable_encoder
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

from server.routers import router


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    yield


app = FastAPI(
    title="物流跟踪 API",
    description="物流跟踪系统 - 提供运单管理、节点跟踪、异常告警等功能",
    version="0.1.0",
    lifespan=lifespan,
)


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(
    request: Request, exc: RequestValidationError
) -> JSONResponse:
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content=jsonable_encoder(
            {
                "error_code": -1,
                "message": "请求参数验证失败",
                "details": exc.errors(),
            }
        ),
    )


app.include_router(router, prefix="/api/v1")


@app.get("/", response_model=dict[str, Any])
async def root() -> dict[str, Any]:
    return {
        "name": "物流跟踪 API",
        "version": "0.1.0",
        "docs": "/docs",
        "openapi": "/openapi.json",
    }
