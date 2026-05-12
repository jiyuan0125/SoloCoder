from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from .database import Base, engine, SessionLocal
from . import models
from .routers import (
    agent_services,
    berths,
    berthing,
    materials,
    waste,
    todos,
    settlements,
)


def init_db():
    Base.metadata.create_all(bind=engine)
    db = SessionLocal()
    try:
        existing_berths = db.query(models.Berth).count()
        if existing_berths == 0:
            default_berths = [
                models.Berth(name="A1", location="东港区", max_length=200, max_draft=12, is_available=True),
                models.Berth(name="A2", location="东港区", max_length=200, max_draft=12, is_available=True),
                models.Berth(name="B1", location="西港区", max_length=300, max_draft=15, is_available=True),
                models.Berth(name="B2", location="西港区", max_length=300, max_draft=15, is_available=True),
                models.Berth(name="C1", location="南港区", max_length=150, max_draft=10, is_available=True),
            ]
            db.add_all(default_berths)
            db.commit()
    finally:
        db.close()


@asynccontextmanager
async def lifespan(app: FastAPI):
    init_db()
    yield


app = FastAPI(
    title="船舶代理业务管理系统",
    description="管理代理服务、靠泊申请、物料补给、垃圾回收、费用结算等业务",
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(agent_services.router)
app.include_router(berths.router)
app.include_router(berthing.router)
app.include_router(materials.router)
app.include_router(waste.router)
app.include_router(todos.router)
app.include_router(settlements.router)


@app.get("/")
def root():
    return {
        "name": "船舶代理业务管理系统",
        "version": "1.0.0",
        "docs": "/docs",
    }


@app.get("/health")
def health():
    return {"status": "ok"}
