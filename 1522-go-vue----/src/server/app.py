import os
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from src.server.routers.projects import router as projects_router
from src.server.routers.boreholes import router as boreholes_router
from src.server.routers.samples import router as samples_router
from src.server.routers.todos import router as todos_router
from src.server.routers.exports import router as exports_router


app = FastAPI(
    title="地质勘探项目管理系统",
    description="地质勘探公司项目管理后端系统",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(projects_router, prefix="/api")
app.include_router(boreholes_router, prefix="/api")
app.include_router(samples_router, prefix="/api")
app.include_router(todos_router, prefix="/api")
app.include_router(exports_router, prefix="/api")


@app.get("/health")
def health_check():
    return {"status": "healthy"}
