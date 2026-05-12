from fastapi import FastAPI
from app.database import engine, Base
from app.routers import work_orders, power_supply, alarms
import os

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="接触网检修管理系统",
    description="供电段接触网检修工单管理与供电数据监控系统",
    version="1.0.0"
)

app.include_router(work_orders.router)
app.include_router(power_supply.router)
app.include_router(alarms.router)


@app.get("/")
def root():
    return {
        "message": "接触网检修管理系统 API",
        "version": "1.0.0",
        "docs": "/docs"
    }


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run("app.main:app", host="0.0.0.0", port=port, reload=True)
