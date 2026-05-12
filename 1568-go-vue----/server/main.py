from fastapi import FastAPI
from .models import Base
from .database import engine
from .routes import router
from .config import PORT

Base.metadata.create_all(bind=engine)

app = FastAPI(title="机场安检管理系统", version="1.0.0")

app.include_router(router, prefix="/api")


@app.get("/")
def root():
    return {
        "name": "机场安检管理系统",
        "version": "1.0.0",
        "docs": "/docs",
        "openapi": "/openapi.json"
    }


def run():
    import uvicorn
    uvicorn.run("server.main:app", host="0.0.0.0", port=PORT, reload=False)


if __name__ == "__main__":
    run()
