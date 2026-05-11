import os
from fastapi import FastAPI
from contextlib import asynccontextmanager

from src.server.api import reactors, recipes, sensors, batches, statistics, scheduler


@asynccontextmanager
async def lifespan(app: FastAPI):
    from src.core.utils.scheduler import get_scheduler
    scheduler = get_scheduler()
    scheduler.start()
    yield
    scheduler.stop()


app = FastAPI(
    title="化工厂反应釜监控与投料管理系统",
    description="FastAPI backend for reactor monitoring and feeding management",
    version="1.0.0",
    lifespan=lifespan
)


app.include_router(reactors.router, prefix="/api/v1")
app.include_router(recipes.router, prefix="/api/v1")
app.include_router(sensors.router, prefix="/api/v1")
app.include_router(batches.router, prefix="/api/v1")
app.include_router(statistics.router, prefix="/api/v1")
app.include_router(scheduler.router, prefix="/api/v1")


@app.get("/")
def root():
    return {
        "name": "化工厂反应釜监控与投料管理系统",
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc"
    }


@app.get("/health")
def health():
    return {"status": "healthy"}
