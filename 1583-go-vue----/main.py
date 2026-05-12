from fastapi import FastAPI
from contextlib import asynccontextmanager

from config import settings
from database import init_db
from routers import inquiries, lost_found, complaints, statistics
from scheduler import start_scheduler, stop_scheduler


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    start_scheduler()
    yield
    stop_scheduler()


app = FastAPI(
    title=settings.APP_NAME,
    description="火车站客服管理系统 - 问询记录、失物招领、投诉管理",
    version="1.0.0",
    lifespan=lifespan
)


app.include_router(inquiries.router, prefix="/api")
app.include_router(lost_found.router, prefix="/api")
app.include_router(complaints.router, prefix="/api")
app.include_router(statistics.router, prefix="/api")


@app.get("/", summary="系统健康检查")
def root():
    return {
        "app": settings.APP_NAME,
        "version": "1.0.0",
        "status": "running",
        "docs": "/docs"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=settings.PORT, reload=False)
