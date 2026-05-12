from fastapi import FastAPI
from server.routers import pilots, applications, tasks, system
from server.database import init_db

app = FastAPI(
    title="引航站调度管理系统",
    description="FastAPI 后端服务，支持引航申请、调度推荐、任务管理和统计分析",
    version="1.0.0"
)


@app.on_event("startup")
def startup_event():
    init_db()


app.include_router(pilots.router)
app.include_router(applications.router)
app.include_router(tasks.router)
app.include_router(system.router)


@app.get("/")
def read_root():
    return {
        "name": "引航站调度管理系统",
        "version": "1.0.0",
        "status": "running"
    }


@app.get("/health")
def health_check():
    return {"status": "healthy"}
