from fastapi import FastAPI
from contextlib import asynccontextmanager
from apscheduler.schedulers.background import BackgroundScheduler
from datetime import date

from server.config import settings
from server.database import engine, Base, SessionLocal
from server.routers import canals, gates, water_plans, water_usage, dispatch, warnings, coefficient
from server import crud

Base.metadata.create_all(bind=engine)

scheduler = BackgroundScheduler()


def scheduled_dispatch():
    db = SessionLocal()
    try:
        today = date.today()
        crud.generate_daily_dispatch(db, target_date=today)
    finally:
        db.close()


@asynccontextmanager
async def lifespan(app: FastAPI):
    scheduler.add_job(
        scheduled_dispatch,
        'cron',
        hour=6,
        minute=0,
        id='daily_dispatch',
        replace_existing=True
    )
    scheduler.start()
    yield
    scheduler.shutdown()


app = FastAPI(
    title=settings.APP_NAME,
    description="灌区水利设施和用水调度系统 API",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(canals.router)
app.include_router(gates.router)
app.include_router(water_plans.router)
app.include_router(water_usage.router)
app.include_router(dispatch.router)
app.include_router(warnings.router)
app.include_router(coefficient.router)


@app.get("/", tags=["根路径"])
def root():
    return {
        "name": settings.APP_NAME,
        "version": "1.0.0",
        "docs": "/docs"
    }


@app.get("/health", tags=["健康检查"])
def health():
    return {"status": "healthy"}
