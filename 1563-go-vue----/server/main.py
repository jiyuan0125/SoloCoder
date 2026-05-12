import os
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .database import init_db
from .routers import parts, requisitions, repairs, purchases, inventory


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


app = FastAPI(
    title="航材管理系统 API",
    description="FastAPI 后端服务，管理航材目录、领用、送修、采购和库存盘点",
    version="1.0.0",
    lifespan=lifespan
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(parts.router, prefix="/api")
app.include_router(requisitions.router, prefix="/api")
app.include_router(repairs.router, prefix="/api")
app.include_router(purchases.router, prefix="/api")
app.include_router(inventory.router, prefix="/api")


@app.get("/")
def root():
    return {"message": "航材管理系统 API", "docs": "/docs"}


@app.get("/health")
def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("PORT", 8000))
    uvicorn.run("server.main:app", host="0.0.0.0", port=port, reload=True)
