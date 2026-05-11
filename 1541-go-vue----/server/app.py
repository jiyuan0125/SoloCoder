from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from server.config import settings
from server.database import Base, engine
from server.routes import router

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title=settings.app_name,
    version=settings.version,
    description="水利枢纽运行调度管理系统 - FastAPI 后端"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(router)


@app.get("/")
def root():
    return {
        "name": settings.app_name,
        "version": settings.version,
        "status": "running"
    }


@app.get("/health")
def health():
    return {"status": "healthy"}
