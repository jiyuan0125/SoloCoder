import os
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .database import Base, engine
from .routers import licenses_router, mining_router, enforcement_router, patrol_router

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="河道采砂全链条管理系统",
    description="采砂许可证审批、作业报备、过磅数据、执法管理、巡查任务、月报生成等功能",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(licenses_router, prefix="/api")
app.include_router(mining_router, prefix="/api")
app.include_router(enforcement_router, prefix="/api")
app.include_router(patrol_router, prefix="/api")


@app.get("/")
def root():
    return {
        "name": "河道采砂全链条管理系统",
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc"
    }


@app.get("/health")
def health_check():
    return {"status": "healthy"}
