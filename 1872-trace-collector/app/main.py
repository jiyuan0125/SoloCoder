from datetime import datetime, timedelta
from contextlib import asynccontextmanager

from fastapi import FastAPI
from apscheduler.schedulers.background import BackgroundScheduler
from apscheduler.triggers.interval import IntervalTrigger

from .database import init_db, SessionLocal
from .models import SpanModel
from .routes import router
from .config import settings


def cleanup_old_data():
    cutoff_time = datetime.utcnow() - timedelta(days=settings.data_retention_days)
    db = SessionLocal()
    try:
        deleted = (
            db.query(SpanModel)
            .filter(SpanModel.created_at < cutoff_time)
            .delete(synchronize_session=False)
        )
        db.commit()
    except Exception as e:
        db.rollback()
        print(f"Cleanup error: {e}")
    finally:
        db.close()


scheduler = BackgroundScheduler()


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    
    scheduler.add_job(
        cleanup_old_data,
        trigger=IntervalTrigger(hours=1),
        id="cleanup_old_spans",
        replace_existing=True,
    )
    scheduler.start()
    
    try:
        yield
    finally:
        scheduler.shutdown()


app = FastAPI(
    title="Trace Collector",
    description="FastAPI Trace Data Collection Service",
    version="1.0.0",
    lifespan=lifespan,
)

app.include_router(router)
