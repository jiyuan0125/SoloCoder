from fastapi import FastAPI
from contextlib import asynccontextmanager

from app.database import init_db
from app.config import settings
from app.routers import venue, user, booking, tournament, training
from app.scheduler import start_scheduler


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    scheduler = start_scheduler()
    yield
    if scheduler.running:
        scheduler.shutdown()


app = FastAPI(
    title=settings.APP_NAME,
    version=settings.APP_VERSION,
    lifespan=lifespan,
)

app.include_router(venue.router)
app.include_router(user.router)
app.include_router(booking.router)
app.include_router(tournament.router)
app.include_router(training.router)


@app.get("/")
def root():
    return {
        "name": settings.APP_NAME,
        "version": settings.APP_VERSION,
        "status": "running",
    }


@app.get("/health")
def health_check():
    return {"status": "healthy"}
