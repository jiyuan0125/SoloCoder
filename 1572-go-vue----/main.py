from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.routers import trains, routes, crews, schedules, operations
from app.database import init_db

app = FastAPI(title="铁路客运调度管理系统", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(trains.router, prefix="/api/trains", tags=["列车管理"])
app.include_router(routes.router, prefix="/api/routes", tags=["交路管理"])
app.include_router(crews.router, prefix="/api/crews", tags=["乘务组管理"])
app.include_router(schedules.router, prefix="/api/schedules", tags=["排班管理"])
app.include_router(operations.router, prefix="/api/operations", tags=["运营管理"])


@app.on_event("startup")
async def startup_event():
    init_db()


@app.get("/")
def root():
    return {"message": "铁路客运调度管理系统", "version": "1.0.0"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host="0.0.0.0", port=8502, reload=True)
