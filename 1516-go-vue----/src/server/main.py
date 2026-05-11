import os
from fastapi import FastAPI
from src.server.routes import router

app = FastAPI(title="宠物美容店预约管理系统", version="1.0.0")

app.include_router(router, prefix="/api/v1")


@app.get("/")
def root():
    return {"message": "宠物美容店预约管理系统 API", "version": "1.0.0"}


def run():
    port = int(os.environ.get("PORT", "8000"))
    host = os.environ.get("HOST", "0.0.0.0")
    
    import uvicorn
    uvicorn.run(app, host=host, port=port)


if __name__ == "__main__":
    run()
