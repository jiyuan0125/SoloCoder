from fastapi import FastAPI
from app.database import Base, engine
from app.routes import locomotives, maintenance, parts, manuals

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="机务段检修管理系统",
    description="FastAPI 实现的机务段检修管理系统，支持检修计划、配件管理和技术手册借阅",
    version="1.0.0"
)

app.include_router(locomotives.router)
app.include_router(maintenance.router)
app.include_router(parts.router)
app.include_router(manuals.router)


@app.get("/")
def read_root():
    return {"message": "机务段检修管理系统 API", "version": "1.0.0"}


@app.get("/health")
def health_check():
    return {"status": "healthy"}
