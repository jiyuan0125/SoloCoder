from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime
from app.database import get_db
from app import models, schemas

router = APIRouter(prefix="/todos", tags=["todos"])

@router.get("/{inspector_id}/pending", response_model=List[schemas.InspectionResponse])
def get_pending_todos(inspector_id: int, db: Session = Depends(get_db)):
    inspector = db.query(models.Inspector).filter(models.Inspector.id == inspector_id).first()
    if not inspector:
        raise HTTPException(status_code=404, detail="巡检员不存在")
    
    now = datetime.utcnow()
    pending = db.query(models.Inspection).filter(
        models.Inspection.inspector_id == inspector_id,
        models.Inspection.status == models.InspectionStatus.PENDING,
        models.Inspection.planned_start_time <= now
    ).all()
    
    return pending

@router.get("/{inspector_id}/overdue", response_model=List[schemas.InspectionResponse])
def get_overdue_todos(inspector_id: int, db: Session = Depends(get_db)):
    inspector = db.query(models.Inspector).filter(models.Inspector.id == inspector_id).first()
    if not inspector:
        raise HTTPException(status_code=404, detail="巡检员不存在")
    
    now = datetime.utcnow()
    overdue = db.query(models.Inspection).filter(
        models.Inspection.inspector_id == inspector_id,
        models.Inspection.status.in_([models.InspectionStatus.PENDING, models.InspectionStatus.IN_PROGRESS]),
        models.Inspection.planned_end_time < now
    ).all()
    
    return overdue
