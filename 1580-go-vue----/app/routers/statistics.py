from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from sqlalchemy import func
from datetime import datetime
from dateutil.relativedelta import relativedelta
from app.database import get_db
from app import models, schemas

router = APIRouter(prefix="/statistics", tags=["statistics"])

@router.get("/", response_model=schemas.StatisticsResponse)
def get_statistics(db: Session = Depends(get_db)):
    now = datetime.utcnow()
    start_of_month = now.replace(day=1, hour=0, minute=0, second=0, microsecond=0)
    
    total_plans = db.query(func.count(models.Inspection.id)).filter(
        models.Inspection.planned_start_time >= start_of_month,
        models.Inspection.status != models.InspectionStatus.CANCELLED
    ).scalar() or 0
    
    completed = db.query(func.count(models.Inspection.id)).filter(
        models.Inspection.planned_start_time >= start_of_month,
        models.Inspection.status.in_([
            models.InspectionStatus.COMPLETED, 
            models.InspectionStatus.OVERDUE_COMPLETED
        ])
    ).scalar() or 0
    
    completion_rate = (completed / total_plans * 100) if total_plans > 0 else 0.0
    
    problems_by_level = {}
    for level in [models.ProblemLevel.MINOR, models.ProblemLevel.MAJOR, models.ProblemLevel.SERIOUS]:
        count = db.query(func.count(models.Inspection.id)).filter(
            models.Inspection.planned_start_time >= start_of_month,
            models.Inspection.problem_level == level
        ).scalar() or 0
        problems_by_level[level.value] = count
    
    device_status_distribution = {}
    for status in [models.DeviceStatus.GOOD, models.DeviceStatus.WARNING, models.DeviceStatus.CRITICAL]:
        count = db.query(func.count(models.Device.id)).filter(
            models.Device.status == status
        ).scalar() or 0
        device_status_distribution[status.value] = count
    
    return schemas.StatisticsResponse(
        completion_rate=round(completion_rate, 2),
        problems_by_level=problems_by_level,
        device_status_distribution=device_status_distribution
    )
