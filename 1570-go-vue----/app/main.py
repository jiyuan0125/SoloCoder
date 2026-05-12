from fastapi import FastAPI

from .database import engine, Base
from .config import APP_PORT
from .routers import flights, cargo, load, dispatchers

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="航空货运管理系统",
    description="航空货运部货物管理后端系统",
    version="1.0.0"
)

app.include_router(flights.router)
app.include_router(cargo.router)
app.include_router(load.router)
app.include_router(dispatchers.router)


@app.get("/")
def read_root():
    return {
        "message": "航空货运管理系统",
        "version": "1.0.0",
        "docs": "/docs"
    }


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("app.main:app", host="0.0.0.0", port=APP_PORT, reload=True)
