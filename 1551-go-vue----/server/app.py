import os
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from server.database import init_db
from server.routes import router

app = FastAPI(
    title="港口泊位与装卸调度系统",
    description="Port Berth and Handling Scheduling System",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(router, prefix="/api/v1", tags=["port"])


@app.on_event("startup")
def startup_event():
    init_db()


@app.get("/", tags=["root"])
def read_root():
    return {
        "message": "港口泊位与装卸调度系统 API 服务运行中",
        "docs": "/docs",
        "api_prefix": "/api/v1"
    }


@app.get("/health", tags=["health"])
def health_check():
    return {"status": "healthy"}
