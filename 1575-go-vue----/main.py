from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from app.database import init_db
from app.routers import stations, security, checkin, passenger_flow

app = FastAPI(title="高铁车站客运管理系统", version="1.0.0")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.on_event("startup")
async def startup_event():
    init_db()

app.include_router(stations.router, prefix="/api/v1", tags=["车站管理"])
app.include_router(security.router, prefix="/api/v1", tags=["安检管理"])
app.include_router(checkin.router, prefix="/api/v1", tags=["检票管理"])
app.include_router(passenger_flow.router, prefix="/api/v1", tags=["客流管理"])

@app.get("/")
async def root():
    return {"message": "高铁车站客运管理系统运行中", "version": "1.0.0"}
