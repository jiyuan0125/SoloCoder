from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from .database import Base, engine
from .routes import fishermen, licenses, violations, todos, release, ports

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="渔业资源管理系统",
    description="FastAPI + SQLite 渔业资源管理系统，包含渔民、许可证、违规处理、待办事项、增殖放流、渔港管理等功能",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(fishermen.router)
app.include_router(licenses.router)
app.include_router(violations.router)
app.include_router(todos.router)
app.include_router(release.router)
app.include_router(ports.router)


@app.get("/", tags=["系统"])
def root():
    return {
        "name": "渔业资源管理系统",
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc"
    }


@app.get("/health", tags=["系统"])
def health_check():
    return {"status": "healthy"}
