from fastapi import FastAPI
from contextlib import asynccontextmanager
from server.database import Base, engine
from server.config import get_settings
from server.routers import flights, resources

settings = get_settings()


@asynccontextmanager
async def lifespan(app: FastAPI):
    Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(
    title=settings.app_name,
    lifespan=lifespan
)

app.include_router(flights.router)
app.include_router(resources.router)


@app.get("/")
def root():
    return {"message": settings.app_name, "status": "running"}


@app.get("/health")
def health_check():
    return {"status": "healthy"}
