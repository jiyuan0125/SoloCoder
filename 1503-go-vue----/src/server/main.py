import os
from fastapi import FastAPI
from contextlib import asynccontextmanager

from src.core.database import init_db
from src.server.routers import users, areas, tasks, reports, todos, statistics


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


app = FastAPI(
    title="林火巡护管理系统",
    description="林业局林火巡护管理系统后端服务",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(users.router, prefix="/api")
app.include_router(areas.router, prefix="/api")
app.include_router(tasks.router, prefix="/api")
app.include_router(reports.router, prefix="/api")
app.include_router(todos.router, prefix="/api")
app.include_router(statistics.router, prefix="/api")


@app.get("/")
def root():
    return {"message": "林火巡护管理系统", "status": "running"}


if __name__ == "__main__":
    import uvicorn

    port = int(os.environ.get("PORT", 8000))
    host = os.environ.get("HOST", "0.0.0.0")
    uvicorn.run("src.server.main:app", host=host, port=port, reload=False)
