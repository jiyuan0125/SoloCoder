from fastapi import FastAPI
from src.server.routes import router

app = FastAPI(
    title="养殖场数字化管理系统",
    description="养殖场管理系统后端服务，支持批次管理、繁育管理、防疫管理、出栏统计等功能",
    version="1.0.0"
)

app.include_router(router, prefix="/api")

@app.get("/")
async def root():
    return {
        "name": "养殖场数字化管理系统",
        "version": "1.0.0",
        "status": "running"
    }
