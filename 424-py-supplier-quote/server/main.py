from contextlib import asynccontextmanager
from typing import Any, AsyncIterator

from fastapi import FastAPI, Request, status
from fastapi.exceptions import RequestValidationError
from fastapi.responses import JSONResponse

from shared.constants.error_codes import ErrorCode
from shared.protocols.common import ApiResponse
from server.routers import (
    supplier_router,
    purchase_router,
    quote_router,
    order_router,
    report_router,
)
from server.repositories.database import initialize_repositories
from server.services.exceptions import BusinessException


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    initialize_repositories()
    yield


app = FastAPI(
    title="供应商报价管理系统",
    description="采购流程中供应商报价管理 API",
    version="0.1.0",
    lifespan=lifespan,
)


@app.exception_handler(BusinessException)
async def business_exception_handler(
    request: Request, exc: BusinessException
) -> JSONResponse:
    response: ApiResponse[None] = ApiResponse.create_error(code=exc.code, message=exc.message)
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content=response.model_dump(mode="json"),
    )


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(
    request: Request, exc: RequestValidationError
) -> JSONResponse:
    error_details = "; ".join([f"{e['loc']}: {e['msg']}" for e in exc.errors()])
    response: ApiResponse[None] = ApiResponse.create_error(
        code=ErrorCode.INVALID_INPUT,
        message=f"参数验证失败: {error_details}",
    )
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content=response.model_dump(mode="json"),
    )


@app.exception_handler(Exception)
async def general_exception_handler(
    request: Request, exc: Exception
) -> JSONResponse:
    response: ApiResponse[None] = ApiResponse.create_error(
        code=ErrorCode.INTERNAL_ERROR,
        message=str(exc),
    )
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content=response.model_dump(mode="json"),
    )


app.include_router(supplier_router, prefix="/api/v1")
app.include_router(purchase_router, prefix="/api/v1")
app.include_router(quote_router, prefix="/api/v1")
app.include_router(order_router, prefix="/api/v1")
app.include_router(report_router, prefix="/api/v1")


@app.get("/health", response_model=ApiResponse[str])
async def health_check() -> ApiResponse[str]:
    return ApiResponse.create_success(data="ok", message="服务正常运行")


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
