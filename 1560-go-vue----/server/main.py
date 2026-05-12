from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from server.core.config import settings
from server.core.database import init_db
from server.routers import inventory_router, lighthouse_router, maintenance_router, monitoring_router

app = FastAPI(
    title=settings.APP_NAME,
    version="1.0.0",
    description="灯塔综合管理系统 - 实时监控灯光状态、能源管理、维护计划、备件库存及工单流程",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.on_event("startup")
def on_startup():
    init_db()


@app.get("/health")
def health_check():
    return {"status": "ok", "app": settings.APP_NAME}


app.include_router(lighthouse_router)
app.include_router(monitoring_router)
app.include_router(maintenance_router)
app.include_router(inventory_router)
