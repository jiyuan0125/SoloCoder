from __future__ import annotations

import os
from contextlib import asynccontextmanager

from fastapi import FastAPI

from app.executor import start_workers
from app.routes import router


@asynccontextmanager
async def lifespan(app: FastAPI):
    workers = await start_workers(worker_count=4)
    try:
        yield
    finally:
        for worker in workers:
            worker.cancel()


def create_app() -> FastAPI:
    app = FastAPI(
        title="Retry Service",
        lifespan=lifespan,
    )
    app.include_router(router)
    return app


app = create_app()

if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
