import os
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .database import init_db
from .routers import stores, ingredients, purchase_orders, wastages, alerts, metrics


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


app = FastAPI(title="连锁餐饮后厨管理系统", lifespan=lifespan)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(stores.router, prefix="/api")
app.include_router(ingredients.router, prefix="/api")
app.include_router(purchase_orders.router, prefix="/api")
app.include_router(wastages.router, prefix="/api")
app.include_router(alerts.router, prefix="/api")
app.include_router(metrics.router, prefix="/api")


@app.get("/")
def root():
    return {"message": "连锁餐饮后厨管理系统 API", "version": "1.0.0"}


def get_app():
    return app


def run():
    import uvicorn

    port = int(os.getenv("PORT", "8000"))
    host = os.getenv("HOST", "0.0.0.0")
    uvicorn.run("src.server.main:app", host=host, port=port, reload=False)
