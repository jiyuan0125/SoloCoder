import os
from contextlib import asynccontextmanager

from fastapi import FastAPI

from scheduler.api import router as tasks_router
from scheduler.models import init_db
from scheduler.scheduler import scheduler


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    scheduler.restore_on_startup()
    scheduler.start()
    yield
    scheduler.stop()


app = FastAPI(title="事件调度器", lifespan=lifespan)

app.include_router(tasks_router)


@app.get("/health")
async def health_check():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
