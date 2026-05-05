from fastapi import FastAPI, Request, status
from fastapi.responses import JSONResponse

from shared.constants import ErrorCode
from shared.models import ErrorResponse
from server.routers import outbound, returns, defective, statistics

app = FastAPI(
    title="退货入库管理系统",
    version="0.1.0",
    description="退货入库管理API服务",
)


@app.exception_handler(ValueError)
async def value_error_handler(request: Request, exc: ValueError) -> JSONResponse:
    return JSONResponse(
        status_code=status.HTTP_400_BAD_REQUEST,
        content=ErrorResponse(
            code=ErrorCode.INTERNAL_ERROR,
            message=str(exc),
        ).model_dump(),
    )


@app.exception_handler(Exception)
async def general_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content=ErrorResponse(
            code=ErrorCode.INTERNAL_ERROR,
            message="内部服务器错误",
            detail=str(exc),
        ).model_dump(),
    )


app.include_router(outbound.router, prefix="/api/v1/outbound", tags=["出库单管理"])
app.include_router(returns.router, prefix="/api/v1/returns", tags=["退货管理"])
app.include_router(defective.router, prefix="/api/v1/defective", tags=["瑕疵品管理"])
app.include_router(statistics.router, prefix="/api/v1/statistics", tags=["统计分析"])


@app.get("/health")
async def health_check() -> dict[str, str]:
    return {"status": "healthy"}
