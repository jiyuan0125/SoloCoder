from fastapi import FastAPI
from contextlib import asynccontextmanager

from app.database import engine, Base
from app.config import settings
from app.routers import performances, trainings, venues, education


@asynccontextmanager
async def lifespan(app: FastAPI):
    Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(
    title=settings.APP_NAME,
    description="文化馆综合管理系统后端API",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(performances.router)
app.include_router(trainings.router)
app.include_router(venues.router)
app.include_router(education.router)


@app.get("/")
def root():
    return {
        "name": settings.APP_NAME,
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=settings.PORT, reload=True)
