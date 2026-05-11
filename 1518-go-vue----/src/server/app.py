from fastapi import FastAPI
from contextlib import asynccontextmanager

from core.models import init_db
from server.config import settings
from server.routers import farmers, medicines, routes, visits, todos, dashboard


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


app = FastAPI(
    title=settings.app_name,
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(farmers.router)
app.include_router(medicines.router)
app.include_router(routes.router)
app.include_router(visits.router)
app.include_router(todos.router)
app.include_router(dashboard.router)


@app.get("/health")
def health_check():
    return {"status": "ok"}
