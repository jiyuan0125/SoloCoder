from fastapi import FastAPI
from server.database import engine, Base
from server.routers import router

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="航线网络管理系统",
    description="航空公司航线网络管理系统，支持航线生命周期管理、收益分析和智能优化建议",
    version="1.0.0"
)

app.include_router(router, prefix="/api")

@app.get("/")
def root():
    return {
        "name": "航线网络管理系统",
        "version": "1.0.0",
        "docs": "/docs"
    }
