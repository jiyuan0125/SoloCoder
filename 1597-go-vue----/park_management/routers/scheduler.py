from datetime import datetime
from typing import List
from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session
from ..database import get_db
from ..utils.scheduler import schedule_tasks_by_zone, get_overdue_tasks, get_zone_task_summary
from ..models import FacilityReport, Facility, Activity, ActivityStatus
from .. import schemas

router = APIRouter(prefix="/scheduler", tags=["scheduler"])


@router.get("/schedule", response_model=List[schemas.TaskScheduleItem])
def get_scheduled_tasks(
    limit: int = Query(50, ge=1, le=200),
    db: Session = Depends(get_db)
):
    return schedule_tasks_by_zone(db, limit=limit)


@router.get("/overdue")
def get_overdue(db: Session = Depends(get_db)):
    return get_overdue_tasks(db)


@router.get("/zone-summary")
def get_zone_summary(db: Session = Depends(get_db)):
    return get_zone_task_summary(db)


@router.get("/reports/urgent", response_model=List[schemas.FacilityReport])
def list_urgent_reports(db: Session = Depends(get_db)):
    now = datetime.utcnow()
    
    urgent_reports = db.query(FacilityReport).join(Facility).filter(
        FacilityReport.is_handled == False,
        Facility.is_safety_related == True,
        FacilityReport.deadline.isnot(None),
        FacilityReport.deadline > now
    ).order_by(FacilityReport.deadline.asc()).all()
    
    return urgent_reports


@router.get("/reports/overdue", response_model=List[schemas.FacilityReport])
def list_overdue_reports(db: Session = Depends(get_db)):
    now = datetime.utcnow()
    
    overdue_reports = db.query(FacilityReport).join(Facility).filter(
        FacilityReport.is_handled == False,
        FacilityReport.deadline.isnot(None),
        FacilityReport.deadline <= now
    ).order_by(FacilityReport.deadline.asc()).all()
    
    return overdue_reports


@router.get("/activities/requires-security", response_model=List[schemas.Activity])
def list_activities_requiring_security(db: Session = Depends(get_db)):
    return db.query(Activity).filter(
        Activity.requires_security == True,
        Activity.status.in_([ActivityStatus.PENDING, ActivityStatus.APPROVED])
    ).order_by(Activity.start_time.asc()).all()
