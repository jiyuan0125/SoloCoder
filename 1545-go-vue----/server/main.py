from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from contextlib import asynccontextmanager

from .config import settings
from .database import engine, Base
from .routers import auth, sections, monitoring, evaluations, export, audit


@asynccontextmanager
async def lifespan(app: FastAPI):
    async with engine.begin() as conn:
        await conn.run_sync(Base.metadata.create_all)
    
    from sqlalchemy.ext.asyncio import AsyncSession
    from sqlalchemy import select
    from .models import User, UserRole
    from .auth import get_password_hash
    
    async with AsyncSession(engine) as session:
        result = await session.execute(select(User).where(User.username == "admin"))
        admin = result.scalar_one_or_none()
        if not admin:
            admin = User(
                username="admin",
                hashed_password=get_password_hash("admin123"),
                full_name="系统管理员",
                role=UserRole.ADMIN
            )
            session.add(admin)
            await session.commit()
    
    yield


app = FastAPI(
    title="水质监测管理系统",
    description="环保局水质监测管理系统 - FastAPI后端",
    version="1.0.0",
    lifespan=lifespan
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(auth.router)
app.include_router(sections.router)
app.include_router(monitoring.router)
app.include_router(evaluations.router)
app.include_router(export.router)
app.include_router(audit.router)


@app.get("/")
async def root():
    return {
        "name": "水质监测管理系统",
        "version": "1.0.0",
        "docs": "/docs"
    }


@app.get("/health")
async def health_check():
    return {"status": "healthy"}
