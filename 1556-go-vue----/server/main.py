from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from .config import engine, Base, PORT
from .routers import crew, certificates, ships, attendance

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="船员管理系统",
    description="船员信息、证书管理、培训记录、考勤工资、船舶分配管理系统",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(crew.router)
app.include_router(certificates.router)
app.include_router(ships.router)
app.include_router(attendance.router)


@app.get("/")
def root():
    return {
        "message": "船员管理系统 API",
        "version": "1.0.0",
        "docs": "/docs",
        "redoc": "/redoc"
    }


@app.get("/health")
def health_check():
    return {"status": "healthy"}


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("server.main:app", host="0.0.0.0", port=PORT, reload=True)
