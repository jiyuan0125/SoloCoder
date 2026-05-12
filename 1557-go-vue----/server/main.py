from fastapi import FastAPI
from contextlib import asynccontextmanager

from .config import PORT
from .database import engine, Base
from .routers import vessels, inspections, certificates, dockings, todos, operations


def init_db():
    Base.metadata.create_all(bind=engine)


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


app = FastAPI(
    title="船舶检验管理系统",
    description="船舶检验、证书管理、坞修安排和待办提醒系统",
    version="1.0.0",
    lifespan=lifespan,
)


app.include_router(vessels.router, prefix="/api/v1/vessels", tags=["船舶管理"])
app.include_router(inspections.router, prefix="/api/v1/inspections", tags=["检验管理"])
app.include_router(certificates.router, prefix="/api/v1/certificates", tags=["证书管理"])
app.include_router(dockings.router, prefix="/api/v1/dockings", tags=["坞修管理"])
app.include_router(todos.router, prefix="/api/v1/todos", tags=["待办提醒"])
app.include_router(operations.router, prefix="/api/v1/operations", tags=["营运记录"])


@app.get("/", tags=["系统"])
def root():
    return {
        "name": "船舶检验管理系统",
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc",
    }


@app.get("/health", tags=["系统"])
def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn

    uvicorn.run("server.main:app", host="0.0.0.0", port=PORT, reload=True)
