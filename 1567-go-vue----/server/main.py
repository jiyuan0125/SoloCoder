from fastapi import FastAPI
from contextlib import asynccontextmanager

from .database import init_db
from .routers import router


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


app = FastAPI(
    title="机场行李全程追踪管理系统",
    description="机场行李追踪系统，支持托运、安检、分拣、装载、运输、到达、传送带和提取等环节",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(router, prefix="/api/v1")


@app.get("/health")
def health_check():
    return {"status": "ok"}
