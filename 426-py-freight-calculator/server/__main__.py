import uvicorn
from typing import Any
from fastapi import FastAPI, Request, status
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError

from shared.responses import ErrorResponse, ErrorDetail
from shared.errors import ErrorCode

from server.routes import calculate, statistics, config


app = FastAPI(
    title="Freight Calculator API",
    description="跨系统多模块运费计算服务",
    version="1.0.0",
)


@app.exception_handler(RequestValidationError)
async def validation_exception_handler(
    request: Request,
    exc: RequestValidationError,
) -> JSONResponse:
    error_response = ErrorResponse(
        error=ErrorDetail(
            code=ErrorCode.INVALID_REQUEST.value,
            message="请求参数验证失败",
            details={"errors": exc.errors()},
        )
    )
    return JSONResponse(
        status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
        content=error_response.model_dump(),
    )


@app.exception_handler(Exception)
async def global_exception_handler(
    request: Request,
    exc: Exception,
) -> JSONResponse:
    error_response = ErrorResponse(
        error=ErrorDetail(
            code=ErrorCode.INTERNAL_ERROR.value,
            message=str(exc),
        )
    )
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content=error_response.model_dump(),
    )


app.include_router(calculate.router, prefix="/api/v1")
app.include_router(statistics.router, prefix="/api/v1")
app.include_router(config.router, prefix="/api/v1")


@app.get("/health")
async def health_check() -> dict[str, Any]:
    return {
        "status": "healthy",
        "service": "freight-calculator",
        "version": "1.0.0",
    }


def main() -> None:
    uvicorn.run(
        "server.__main__:app",
        host="127.0.0.1",
        port=8000,
        reload=True,
    )


if __name__ == "__main__":
    main()
