from fastapi import FastAPI
from contextlib import asynccontextmanager

from app.config import get_settings
from app.database import Base, engine
from app.routers import ventilation, lighting, drainage, alarms

settings = get_settings()


@asynccontextmanager
async def lifespan(app: FastAPI):
    Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(
    title=settings.app_name,
    description="隧道综合监控系统 - 通风、照明、排水集成管理平台",
    version="1.0.0",
    lifespan=lifespan
)


app.include_router(ventilation.router)
app.include_router(lighting.router)
app.include_router(drainage.router)
app.include_router(alarms.router)


@app.get("/")
def root():
    return {
        "name": settings.app_name,
        "version": "1.0.0",
        "endpoints": {
            "ventilation": "/ventilation/",
            "lighting": "/lighting/",
            "drainage": "/drainage/",
            "alarms": "/alarms/"
        }
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "app.main:app",
        host=settings.api_host,
        port=settings.api_port,
        reload=True
    )
