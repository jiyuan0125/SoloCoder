from fastapi import FastAPI
from contextlib import asynccontextmanager

from .database import init_db
from .routers import router


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


def create_app() -> FastAPI:
    app = FastAPI(
        title="水文数据采集管理系统",
        description="FastAPI 后端服务，支持水文数据采集、质量检查、数据聚合和设备检定管理",
        version="1.0.0",
        lifespan=lifespan
    )

    app.include_router(router, prefix="/api", tags=["API"])

    @app.get("/", tags=["Health"])
    def health_check():
        return {"status": "ok", "service": "hydrology-management-system"}

    return app
