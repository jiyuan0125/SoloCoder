import os
from fastapi import FastAPI
from app.database import engine, Base
from app.routes import router as config_router
from app.watch_routes import router as watch_router

Base.metadata.create_all(bind=engine)

app = FastAPI(title="配置中心")

app.include_router(config_router)
app.include_router(watch_router)


@app.get("/health")
def health():
    return {"status": "ok"}


if __name__ == "__main__":
    import uvicorn
    port = int(os.environ.get("PORT", 8000))
    uvicorn.run("app.main:app", host="0.0.0.0", port=port, reload=True)
