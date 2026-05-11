from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager
from server.database import init_db
from server.routers import rivers, keepers, patrols, issues, reports

@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield

app = FastAPI(
    title="河长制综合管理系统",
    description="FastAPI 实现的河长制综合管理系统后端服务",
    version="1.0.0",
    lifespan=lifespan
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(rivers.router)
app.include_router(keepers.router)
app.include_router(patrols.router)
app.include_router(issues.router)
app.include_router(reports.router)

@app.get("/")
def root():
    return {
        "message": "河长制综合管理系统",
        "version": "1.0.0",
        "docs": "/docs"
    }

@app.get("/health")
def health_check():
    return {"status": "healthy"}
