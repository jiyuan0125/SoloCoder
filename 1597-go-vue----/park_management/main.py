from fastapi import FastAPI
from .database import Base, engine
from .routers import green_zones, facilities, activities, scheduler

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="公园综合管理系统",
    description="公园绿化区域、设施和活动的综合管理系统",
    version="1.0.0"
)

app.include_router(green_zones.router, tags=["绿化区域"])
app.include_router(facilities.router, tags=["设施管理"])
app.include_router(activities.router, tags=["活动管理"])
app.include_router(scheduler.router, tags=["调度器"])


@app.get("/", tags=["系统"])
def root():
    return {
        "message": "公园综合管理系统 API",
        "version": "1.0.0",
        "docs": "/docs"
    }


@app.get("/health", tags=["系统"])
def health_check():
    return {"status": "healthy"}
