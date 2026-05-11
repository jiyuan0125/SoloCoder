from fastapi import FastAPI
from .config import settings
from .database import Base, engine
from .routers import router

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title=settings.app_name,
    version=settings.version,
    description="海事局综合管理系统 - 行政审批、执法检查、行政处罚、整改管理一体化平台"
)

app.include_router(router, prefix="/api/v1")
app.include_router(router)


@app.get("/health")
def health_check():
    return {"status": "healthy", "app": settings.app_name, "version": settings.version}
