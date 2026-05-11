from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .config import settings
from .database import Base, engine
from . import models
from .routers import units, sources, inspections, approvals, audits

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="放射源全生命周期管理系统",
    description="FastAPI + SQLite 实现的放射源全生命周期管理系统",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(units.router, prefix="/api")
app.include_router(sources.router, prefix="/api")
app.include_router(inspections.router, prefix="/api")
app.include_router(approvals.router, prefix="/api")
app.include_router(audits.router, prefix="/api")


@app.get("/", tags=["系统"])
def root():
    return {
        "message": "放射源全生命周期管理系统 API",
        "version": "1.0.0",
        "docs": "/docs",
    }


@app.get("/health", tags=["系统"])
def health():
    return {"status": "ok"}
