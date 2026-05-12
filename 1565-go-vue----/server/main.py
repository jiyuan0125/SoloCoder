from fastapi import FastAPI
from .database import engine, Base
from .routers import waypoints, routes, flights, simulation

Base.metadata.create_all(bind=engine)

app = FastAPI(title="空管调度辅助系统", version="1.0.0")

app.include_router(waypoints.router)
app.include_router(routes.router)
app.include_router(flights.router)
app.include_router(simulation.router)


@app.get("/")
def root():
    return {
        "name": "空管调度辅助系统",
        "version": "1.0.0",
        "endpoints": {
            "waypoints": "/waypoints/",
            "routes": "/routes/",
            "flights": "/flights/",
            "simulation": "/simulation/run"
        }
    }
