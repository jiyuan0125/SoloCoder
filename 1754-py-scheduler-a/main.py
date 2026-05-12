import os
from contextlib import asynccontextmanager
from fastapi import FastAPI
from app.routes import router
from app.scheduler import start_scheduler


@asynccontextmanager
async def lifespan(app: FastAPI):
    start_scheduler()
    yield


app = FastAPI(
    title="Unified Task Scheduler",
    description="A centralized task scheduler supporting cron expressions and one-time delayed tasks",
    version="1.0.0",
    lifespan=lifespan,
)

app.include_router(router, prefix="/api", tags=["tasks"])


@app.get("/")
def root():
    return {
        "service": "unified-task-scheduler",
        "status": "running",
    }


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=False)
