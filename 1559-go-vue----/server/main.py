from fastapi import FastAPI
from .database import init_db
from .routers import router


app = FastAPI(
    title="Search and Rescue Coordination System",
    description="搜救协调管理系统 - 报警管理、事件流程、力量调度、结果记录",
    version="1.0.0"
)


@app.on_event("startup")
def startup_event():
    init_db()


app.include_router(router, prefix="/api/v1", tags=["All API"])


@app.get("/health")
def health_check():
    return {"status": "ok"}
