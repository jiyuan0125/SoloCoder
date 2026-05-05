from typing import Any, AsyncIterator, Callable, Dict
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from starlette.exceptions import HTTPException as StarletteHTTPException
from contextlib import asynccontextmanager

from server.routes import router
from shared import ApiResponse, ErrorCode


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    yield


app = FastAPI(
    title="Shelf Audit System API",
    description="Warehouse shelf audit and management system",
    version="0.1.0",
    lifespan=lifespan
)

app.include_router(router)


@app.exception_handler(StarletteHTTPException)
async def http_exception_handler(request: Request, exc: StarletteHTTPException) -> JSONResponse:
    if exc.detail and isinstance(exc.detail, dict):
        return JSONResponse(
            status_code=exc.status_code,
            content=exc.detail
        )
    return JSONResponse(
        status_code=exc.status_code,
        content=ApiResponse(
            success=False,
            error_code=ErrorCode.INVALID_REQUEST,
            error_message=str(exc.detail)
        ).model_dump()
    )


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(request: Request, exc: RequestValidationError) -> JSONResponse:
    return JSONResponse(
        status_code=422,
        content=ApiResponse(
            success=False,
            error_code=ErrorCode.INVALID_REQUEST,
            error_message=f"Validation error: {str(exc)}"
        ).model_dump()
    )


@app.exception_handler(Exception)
async def general_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    return JSONResponse(
        status_code=500,
        content=ApiResponse(
            success=False,
            error_code=ErrorCode.INTERNAL_ERROR,
            error_message=str(exc)
        ).model_dump()
    )


@app.get("/health")
async def health_check() -> Dict[str, str]:
    return {"status": "healthy"}
