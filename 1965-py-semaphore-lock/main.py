import os
import uvicorn
from fastapi import FastAPI

from semaphore_manager import SemaphoreManager
from routes import create_routes


DEFAULT_WAIT_TIMEOUT = 30.0
DEFAULT_PORT = 8000


def create_app(default_wait_timeout: float = DEFAULT_WAIT_TIMEOUT) -> FastAPI:
    app = FastAPI(
        title="Distributed Semaphore API",
        description="A FastAPI-based distributed semaphore service for controlling concurrent access to limited resources",
        version="1.0.0"
    )

    semaphore_manager = SemaphoreManager(default_wait_timeout=default_wait_timeout)
    router = create_routes(semaphore_manager)
    
    app.include_router(router, prefix="/api/v1")
    app.include_router(router)

    return app


app = create_app()


if __name__ == "__main__":
    port_str = os.environ.get("PORT", str(DEFAULT_PORT))
    try:
        port = int(port_str)
    except ValueError:
        port = DEFAULT_PORT

    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=port,
        reload=False
    )
