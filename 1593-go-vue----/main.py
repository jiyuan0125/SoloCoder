import os
from fastapi import FastAPI
from contextlib import asynccontextmanager
from app.database import Base, engine
from app.routers import collections, exhibitions, reservations, meetings, borrowings, todos


@asynccontextmanager
async def lifespan(app: FastAPI):
    Base.metadata.create_all(bind=engine)
    yield


app = FastAPI(
    title="博物馆数字化管理系统",
    description="管理藏品、展览、参观预约、会议预约和外借管理",
    version="1.0.0",
    lifespan=lifespan
)

app.include_router(collections.router, prefix="/api/collections", tags=["藏品管理"])
app.include_router(exhibitions.router, prefix="/api/exhibitions", tags=["展览管理"])
app.include_router(reservations.router, prefix="/api/reservations", tags=["参观预约"])
app.include_router(meetings.router, prefix="/api/meetings", tags=["会议预约"])
app.include_router(borrowings.router, prefix="/api/borrowings", tags=["外借管理"])
app.include_router(todos.router, prefix="/api/todos", tags=["待办事项"])


@app.get("/")
def root():
    return {"message": "博物馆数字化管理系统", "version": "1.0.0"}


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)
