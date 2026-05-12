from fastapi import FastAPI
from contextlib import asynccontextmanager
from app.config import settings
from app.database import engine, Base
from app.routers import stations, routes, containers, vehicles, orders


@asynccontextmanager
async def lifespan(app: FastAPI):
    Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(
    title="多式联运调度系统",
    description="Intermodal Transportation Dispatch System",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(stations.router)
app.include_router(routes.router)
app.include_router(containers.router)
app.include_router(vehicles.router)
app.include_router(orders.router)


@app.get("/")
def root():
    return {
        "name": "多式联运调度系统",
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host=settings.HOST,
        port=settings.PORT,
        reload=True
    )
