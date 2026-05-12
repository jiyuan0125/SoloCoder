import os
from fastapi import FastAPI
from server.database import Base, engine
from server.routers import shops, sales, settlements

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="机场免税店综合管理系统",
    description="FastAPI + SQLite 的机场免税店管理后端",
    version="1.0.0",
)

app.include_router(shops.router)
app.include_router(sales.router)
app.include_router(settlements.router)


@app.get("/")
def root():
    return {
        "name": "机场免税店综合管理系统",
        "version": "1.0.0",
        "docs": "/docs",
    }
