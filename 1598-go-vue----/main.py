from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from datetime import date
from database import Base, engine
from config import settings
from routers import management, feeding, veterinary, todos
from models import *
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

Base.metadata.create_all(bind=engine)

app = FastAPI(
    title="动物园综合管理系统",
    description="Zoo Management System with FastAPI backend",
    version="1.0.0"
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(management.router)
app.include_router(feeding.router)
app.include_router(veterinary.router)
app.include_router(todos.router)


@app.get("/")
def root():
    return {
        "name": "动物园综合管理系统",
        "version": "1.0.0",
        "status": "running",
        "port": settings.PORT
    }


@app.get("/health")
def health_check():
    return {"status": "healthy"}


@app.get("/dashboard")
def dashboard():
    from database import SessionLocal
    from sqlalchemy import func
    
    db = SessionLocal()
    try:
        total_animals = db.query(Animal).count()
        sick_animals = db.query(Animal).filter(Animal.health_status == AnimalStatus.SICK).count()
        isolation_animals = db.query(Animal).filter(Animal.is_isolation == True).count()
        low_stock_feeds = db.query(Feed).filter(Feed.current_stock < Feed.safety_stock).count()
        
        today = date.today()
        today_feeding = db.query(FeedingRecord).filter(FeedingRecord.feeding_date == today).count()
        
        pending_todos = db.query(Todo).filter(Todo.status == TodoStatus.PENDING).count()
        
        return {
            "total_animals": total_animals,
            "sick_animals": sick_animals,
            "isolation_animals": isolation_animals,
            "low_stock_feeds": low_stock_feeds,
            "today_feeding_records": today_feeding,
            "pending_todos": pending_todos
        }
    finally:
        db.close()


if __name__ == "__main__":
    import uvicorn
    uvicorn.run("main:app", host=settings.HOST, port=settings.PORT, reload=True)
