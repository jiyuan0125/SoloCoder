from fastapi import FastAPI
from contextlib import asynccontextmanager

from core.config import settings
from core.database import engine, Base
from core import models
from server.routes import router
from server.scheduler import start_scheduler, stop_scheduler


models.Base.metadata.create_all(bind=engine)


@asynccontextmanager
async def lifespan(app: FastAPI):
    start_scheduler()
    yield
    stop_scheduler()


app = FastAPI(
    title=settings.APP_NAME,
    version="1.0.0",
    lifespan=lifespan,
)

app.include_router(router, prefix="/api")


@app.get("/health")
def health_check():
    return {"status": "ok"}
