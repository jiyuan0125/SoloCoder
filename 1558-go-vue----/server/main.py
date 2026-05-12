from fastapi import FastAPI

from server.config import Base, engine
from server.routers import ships, declarations, audit

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="危险品运输管理系统",
    description="危险品运输申报、审核、装卸监控和应急处置管理系统",
    version="1.0.0",
)

app.include_router(ships.router)
app.include_router(declarations.router)
app.include_router(audit.router)


@app.get("/", tags=["root"])
def root():
    return {
        "name": "危险品运输管理系统",
        "version": "1.0.0",
        "docs": "/docs",
    }


@app.get("/health", tags=["health"])
def health_check():
    return {"status": "healthy"}
