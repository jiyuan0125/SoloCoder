from fastapi import FastAPI, Depends, HTTPException
from sqlalchemy.orm import Session
from app.database import engine, Base, get_db
from app.routers import devices, inspections, todos, statistics
from app import models, schemas
import os

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="高铁设备巡检管理系统",
    description="FastAPI 后端服务，支持线路、桥梁和隧道设备的巡检管理",
    version="1.0.0"
)

app.include_router(devices.router)
app.include_router(inspections.router)
app.include_router(todos.router)
app.include_router(statistics.router)

@app.get("/")
def read_root():
    return {"message": "高铁设备巡检管理系统 API 服务运行中"}

@app.post("/inspectors/", response_model=schemas.InspectorResponse)
def create_inspector(inspector: schemas.InspectorCreate, db: Session = Depends(get_db)):
    db_inspector = models.Inspector(**inspector.model_dump())
    db.add(db_inspector)
    db.commit()
    db.refresh(db_inspector)
    return db_inspector

@app.get("/inspectors/", response_model=list[schemas.InspectorResponse])
def get_inspectors(db: Session = Depends(get_db)):
    return db.query(models.Inspector).all()

if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", "8000"))
    uvicorn.run(app, host="0.0.0.0", port=port)
