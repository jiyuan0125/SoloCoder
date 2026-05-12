from fastapi import FastAPI
from contextlib import asynccontextmanager

from server.database import init_db
from server.routers import ships, licenses, inspections, rectifications, penalties, audit_logs


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


app = FastAPI(
    title="海事局综合管理系统",
    description="行政审批、执法检查、行政处罚、整改管理综合系统",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(ships.router)
app.include_router(licenses.router)
app.include_router(inspections.router)
app.include_router(rectifications.router)
app.include_router(penalties.router)
app.include_router(audit_logs.router)


@app.get("/")
def root():
    return {
        "name": "海事局综合管理系统",
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc"
    }


@app.get("/health")
def health():
    return {"status": "healthy"}
