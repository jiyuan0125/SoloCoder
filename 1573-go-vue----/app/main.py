from fastapi import FastAPI
from app.database import init_db
from app.routers import switches, interlocking, blocks, semaphores, stats
import os

app = FastAPI(title="铁路信号监控系统", version="1.0.0")

init_db()

app.include_router(switches.router, prefix="/signals", tags=["道岔"])
app.include_router(interlocking.router, prefix="/interlocking", tags=["联锁"])
app.include_router(blocks.router, prefix="/block", tags=["闭塞分区"])
app.include_router(semaphores.router, prefix="/semaphore", tags=["信号机"])
app.include_router(stats.router, prefix="/stats", tags=["统计"])


@app.get("/")
def root():
    return {"message": "铁路信号监控系统 API", "version": "1.0.0"}
