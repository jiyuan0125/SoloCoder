from fastapi import FastAPI
from contextlib import asynccontextmanager

from app.database import engine, Base
from app.config import settings
from app.routers import cards, transactions, faults


@asynccontextmanager
async def lifespan(app: FastAPI):
    Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(
    title=settings.app_name,
    description="地铁自动售检票管理系统后端 API",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(cards.router)
app.include_router(transactions.router)
app.include_router(faults.router)


@app.get("/")
def root():
    return {
        "name": settings.app_name,
        "version": "1.0.0",
        "status": "running",
        "docs": "/docs",
        "redoc": "/redoc"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=settings.port, reload=True)
