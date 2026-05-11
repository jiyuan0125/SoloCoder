import sys
from pathlib import Path
import signal

BASE_DIR = Path(__file__).resolve().parent.parent.parent
sys.path.insert(0, str(BASE_DIR))

from fastapi import FastAPI
from contextlib import asynccontextmanager
from src.core.database import init_db
from src.core.config import settings
from src.server.routers import (
    pipelines,
    segments,
    pressure,
    alarms,
    patrol,
    integrity,
    reports
)
from src.server.scheduler import start_scheduler, stop_scheduler


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    start_scheduler()
    yield
    stop_scheduler()


app = FastAPI(
    title="油气管道巡线监测系统",
    description="油气管道巡线和压力监测后端系统",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(pipelines.router)
app.include_router(segments.router)
app.include_router(pressure.router)
app.include_router(alarms.router)
app.include_router(patrol.router)
app.include_router(integrity.router)
app.include_router(reports.router)


@app.get("/")
def root():
    return {
        "name": "油气管道巡线监测系统",
        "version": "1.0.0",
        "status": "running"
    }


@app.get("/health")
def health():
    return {"status": "healthy"}
