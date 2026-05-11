import os
from contextlib import asynccontextmanager

from fastapi import FastAPI

from src.server.deps import init_container
from src.server.routes import (
    appointments_router,
    medication_router,
    pets_router,
    registrations_router,
    vaccines_router,
)


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_container()
    yield


app = FastAPI(
    title="宠物医院门诊管理系统",
    description="宠物医院门诊管理后端服务",
    version="1.0.0",
    lifespan=lifespan,
)

app.include_router(pets_router)
app.include_router(registrations_router)
app.include_router(vaccines_router)
app.include_router(appointments_router)
app.include_router(medication_router)


@app.get("/")
def root():
    return {
        "name": "宠物医院门诊管理系统",
        "version": "1.0.0",
        "docs": "/docs",
    }


def main():
    import uvicorn

    port = int(os.environ.get("PORT", "8000"))
    host = os.environ.get("HOST", "0.0.0.0")
    uvicorn.run("src.server.app:app", host=host, port=port, reload=False)


if __name__ == "__main__":
    main()
