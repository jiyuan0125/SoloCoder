from fastapi import FastAPI
from contextlib import asynccontextmanager
from app.database import init_db, SessionLocal
from app.models import Cabin, MaintenanceTask, MaintenanceType
from app.routers import queue, cabin, sensor, maintenance, stats
from app.config import get_settings

settings = get_settings()


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    
    db = SessionLocal()
    try:
        if not db.query(Cabin).count():
            default_cabin = Cabin(
                cabin_number="C001",
                max_weight=1000.0
            )
            db.add(default_cabin)
        
        db.commit()
    finally:
        db.close()
    
    yield


app = FastAPI(
    title="景区缆车运营管理系统",
    description="往复式缆车运营管理 - 排队、调度、维护、统计",
    version="0.1.0",
    lifespan=lifespan
)

app.include_router(queue.router)
app.include_router(cabin.router)
app.include_router(sensor.router)
app.include_router(maintenance.router)
app.include_router(stats.router)


@app.get("/", tags=["系统"])
def root():
    return {
        "service": "景区缆车运营管理系统",
        "version": "0.1.0",
        "status": "running",
        "endpoints": [
            "排队: GET /queue/realtime, GET /queue/estimate",
            "发车: POST /cabin/{轿厢号}/dispatch, POST /cabin/{轿厢号}/arrive",
            "传感器: POST /cabin/{轿厢号}/sensor/weight, POST /cabin/{轿厢号}/sensor/status",
            "维护: POST /maintenance/{任务ID}/start, POST /maintenance/{任务ID}/complete",
            "统计: GET /stats/daily, GET /stats/shifts"
        ]
    }


@app.get("/health", tags=["系统"])
def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=settings.port,
        reload=True
    )
