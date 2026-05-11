import os
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .routers import (
    zone_limits,
    monitoring_points,
    noise_data,
    alerts,
    construction_permits,
    statistics,
)


def create_app() -> FastAPI:
    app = FastAPI(
        title="噪声监测点位管理系统",
        description="城市噪声监测点位管理的后端服务",
        version="1.0.0",
    )
    
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )
    
    app.include_router(zone_limits.router)
    app.include_router(monitoring_points.router)
    app.include_router(noise_data.router)
    app.include_router(alerts.router)
    app.include_router(construction_permits.router)
    app.include_router(statistics.router)
    
    return app


app = create_app()


@app.get("/")
def root():
    return {
        "message": "噪声监测点位管理系统 API",
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc",
    }


if __name__ == "__main__":
    import uvicorn
    
    host = os.getenv("HOST", "0.0.0.0")
    port = int(os.getenv("PORT", "8000"))
    
    uvicorn.run("src.server.main:app", host=host, port=port, reload=False)
