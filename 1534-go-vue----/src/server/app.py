from fastapi import FastAPI
from .database import init_db
from .routes import enterprises, facilities, maintenances, emissions, todos, alerts, statistics


def create_app() -> FastAPI:
    app = FastAPI(title="环保设施运维系统", version="1.0.0")
    init_db()
    app.include_router(enterprises.router, prefix="/api/v1")
    app.include_router(facilities.router, prefix="/api/v1")
    app.include_router(maintenances.router, prefix="/api/v1")
    app.include_router(emissions.router, prefix="/api/v1")
    app.include_router(todos.router, prefix="/api/v1")
    app.include_router(alerts.router, prefix="/api/v1")
    app.include_router(statistics.router, prefix="/api/v1")
    return app


app = create_app()
