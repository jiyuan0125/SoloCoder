from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from server.database import Base, engine
from server.routers import projects

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="水土保持项目管理系统",
    description="全流程水土保持项目管理系统 - 项目、措施、进度、监测、验收、资金、待办",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(projects.router)


@app.get("/")
def root():
    return {
        "name": "水土保持项目管理系统",
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc",
    }


@app.get("/health")
def health():
    return {"status": "ok"}
