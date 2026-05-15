import os
import uvicorn
from fastapi import FastAPI
from contextlib import asynccontextmanager

from app.core.database import engine, Base
from app.core.scheduler import init_scheduler, shutdown_scheduler
from app.api.routes import router as api_router

Base.metadata.create_all(bind=engine)


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_scheduler()
    yield
    shutdown_scheduler()


app = FastAPI(
    title="Event Scheduler with Audit",
    description="A cron-based event scheduler with complete execution audit trail",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(api_router, prefix="/api/v1")


@app.get("/health")
async def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    port = int(os.getenv("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)
