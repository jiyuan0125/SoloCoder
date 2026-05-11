from fastapi import FastAPI
from contextlib import asynccontextmanager
from server.database import init_db
from server.routes import router
from server.scheduler import start_scheduler, stop_scheduler


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    start_scheduler()
    yield
    stop_scheduler()


app = FastAPI(
    title="流域防洪调度系统",
    description="FastAPI + SQLite 实现的流域防洪调度系统",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(router, prefix="/api", tags=["flood-control"])


@app.get("/")
def read_root():
    return {"name": "流域防洪调度系统", "version": "1.0.0", "docs": "/docs"}


@app.get("/health")
def health_check():
    return {"status": "ok"}
