from fastapi import FastAPI, Depends
from contextlib import asynccontextmanager
from app.config import settings
from app.database import engine, Base
from app.routers import books, readers, borrowings, seats, events


@asynccontextmanager
async def lifespan(app: FastAPI):
    Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(
    title=settings.APP_NAME,
    lifespan=lifespan,
)


app.include_router(books.router, prefix=settings.API_V1_STR)
app.include_router(readers.router, prefix=settings.API_V1_STR)
app.include_router(borrowings.router, prefix=settings.API_V1_STR)
app.include_router(seats.router, prefix=settings.API_V1_STR)
app.include_router(events.router, prefix=settings.API_V1_STR)


@app.get("/")
def root():
    return {
        "name": settings.APP_NAME,
        "api_version": "v1",
        "docs": "/docs",
    }


@app.get("/health")
def health():
    return {"status": "healthy"}
