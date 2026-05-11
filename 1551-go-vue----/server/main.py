from fastapi import FastAPI
from .config import APP_NAME, APP_VERSION
from .database import init_db
from .routes import router

init_db()

app = FastAPI(title=APP_NAME, version=APP_VERSION)
app.include_router(router, prefix="/api")

@app.get("/")
def root():
    return {"name": APP_NAME, "version": APP_VERSION}
