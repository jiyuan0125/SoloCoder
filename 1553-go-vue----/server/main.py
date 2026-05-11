from fastapi import FastAPI

from server.config import settings
from server.database import init_db
from server.routes import router

app = FastAPI(title=settings.app_name, version="1.0.0")

app.include_router(router, prefix="/api")


@app.on_event("startup")
def on_startup():
    init_db()


@app.get("/", tags=["health"])
def root():
    return {"app": settings.app_name, "status": "running"}


@app.get("/health", tags=["health"])
def health():
    return {"status": "healthy"}


def run_server():
    import uvicorn
    uvicorn.run(
        "server.main:app",
        host="0.0.0.0",
        port=settings.port,
        reload=False
    )
