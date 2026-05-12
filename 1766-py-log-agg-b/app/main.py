import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI, status
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from app.config import settings
from app.database import init_db
from app.routers import logs, stats
from app.log_cleanup import cleanup_service

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
)


@asynccontextmanager
async def lifespan(app: FastAPI):
    await init_db()
    cleanup_service.start()
    try:
        yield
    finally:
        cleanup_service.stop()


app = FastAPI(
    title="Centralized Log Aggregator",
    description=(
        "FastAPI-based centralized log aggregation service "
        "with filtering, search, and distributed tracing support."
    ),
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(logs.router, prefix="/api/v1")
app.include_router(stats.router, prefix="/api/v1")


@app.get("/", tags=["root"])
async def root():
    return {
        "service": "Centralized Log Aggregator",
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc",
        "retention_days": settings.LOG_RETENTION_DAYS,
        "cleanup_interval_hours": settings.CLEANUP_INTERVAL_HOURS,
    }


@app.get("/health", tags=["health"])
async def health():
    return {"status": "ok", "timestamp": __import__("datetime").datetime.utcnow().isoformat()}


@app.exception_handler(Exception)
async def global_exception_handler(request, exc):
    logging.exception("Unhandled exception")
    return JSONResponse(
        status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
        content={"detail": "Internal server error", "error": str(exc)},
    )


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "app.main:app",
        host="0.0.0.0",
        port=settings.PORT,
        reload=False,
    )
