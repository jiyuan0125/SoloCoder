import os
from fastapi import FastAPI
from contextlib import asynccontextmanager

from server.api import router as api_router
from server.database import init_db


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


app = FastAPI(
    title="煤矿通风瓦斯监测管理系统",
    description="煤矿通风瓦斯监测管理系统后端服务",
    version="1.0.0",
    lifespan=lifespan,
)

app.include_router(api_router, prefix="/api/v1")


@app.get("/")
def root():
    return {
        "name": "煤矿通风瓦斯监测管理系统",
        "version": "1.0.0",
        "status": "running",
    }


@app.get("/health")
def health():
    return {"status": "healthy"}
