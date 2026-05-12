from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager

from .database import Base, engine
from .routers import stations, data, redtide, farmers
from .config import PORT


@asynccontextmanager
async def lifespan(app: FastAPI):
    Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(
    title="海洋环境监测系统",
    description="基于FastAPI的海洋环境监测系统，支持岸基站和浮标站数据采集、海浪预警和赤潮监测",
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

app.include_router(stations.router, prefix="/api")
app.include_router(data.router, prefix="/api")
app.include_router(redtide.router, prefix="/api")
app.include_router(farmers.router, prefix="/api")


@app.get("/")
def root():
    return {
        "name": "海洋环境监测系统",
        "version": "1.0.0",
        "status": "running",
        "docs": "/docs"
    }


@app.get("/health")
def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "server.main:app",
        host="0.0.0.0",
        port=PORT,
        reload=False
    )
