from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime
from app.database import get_db
from app import models, schemas
from app.utils import is_flood_season
import uuid

router = APIRouter(prefix="/inspections", tags=["inspections"])

def generate_inspection_code():
    return f"INSP-{uuid.uuid4().hex[:8].upper()}"

@router.post("/", response_model=schemas.InspectionResponse)
def create_inspection(inspection: schemas.InspectionCreate, db: Session = Depends(get_db)):
    device = db.query(models.Device).filter(models.Device.id == inspection.device_id).first()
    if not device:
        raise HTTPException(status_code=404, detail="设备不存在")
    
    inspector = db.query(models.Inspector).filter(models.Inspector.id == inspection.inspector_id).first()
    if not inspector:
        raise HTTPException(status_code=404, detail="巡检员不存在")
    
    if inspection.inspection_type == models.InspectionType.FLOOD and not is_flood_season():
        raise HTTPException(status_code=400, detail="非汛期（11月-次年3月）不允许创建防洪检查任务")
    
    inspection_dict = inspection.model_dump()
    inspection_dict['inspection_code'] = generate_inspection_code()
    inspection_dict['status'] = models.InspectionStatus.PENDING
    
    db_inspection = models.Inspection(**inspection_dict)
    db.add(db_inspection)
    db.commit()
    db.refresh(db_inspection)
    
    return db_inspection

@router.get("/", response_model=List[schemas.InspectionResponse])
def get_inspections(
    status: models.InspectionStatus = None,
    inspection_type: models.InspectionType = None,
    db: Session = Depends(get_db)
):
    query = db.query(models.Inspection)
    if status:
        query = query.filter(models.Inspection.status == status)
    if inspection_type:
        query = query.filter(models.Inspection.inspection_type == inspection_type)
    return query.all()

@router.get("/{inspection_code}", response_model=schemas.InspectionResponse)
def get_inspection(inspection_code: str, db: Session = Depends(get_db)):
    inspection = db.query(models.Inspection).filter(models.Inspection.inspection_code == inspection_code).first()
    if not inspection:
        raise HTTPException(status_code=404, detail="巡检任务不存在")
    return inspection

@router.put("/{inspection_code}/start", response_model=schemas.InspectionResponse)
def start_inspection(inspection_code: str, db: Session = Depends(get_db)):
    inspection = db.query(models.Inspection).filter(models.Inspection.inspection_code == inspection_code).first()
    if not inspection:
        raise HTTPException(status_code=404, detail="巡检任务不存在")
    
    if inspection.status != models.InspectionStatus.PENDING:
        raise HTTPException(status_code=400, detail="只有待开始的任务可以开始")
    
    inspection.status = models.InspectionStatus.IN_PROGRESS
    inspection.actual_start_time = datetime.utcnow()
    db.commit()
    db.refresh(inspection)
    
    return inspection

@router.put("/{inspection_code}/complete", response_model=schemas.InspectionResponse)
def complete_inspection(
    inspection_code: str, 
    result: schemas.InspectionComplete,
    db: Session = Depends(get_db)
):
    inspection = db.query(models.Inspection).filter(models.Inspection.inspection_code == inspection_code).first()
    if not inspection:
        raise HTTPException(status_code=404, detail="巡检任务不存在")
    
    if inspection.status not in [models.InspectionStatus.IN_PROGRESS, models.InspectionStatus.PENDING]:
        raise HTTPException(status_code=400, detail="只有进行中或待开始的任务可以完成")
    
    now = datetime.utcnow()
    if now > inspection.planned_end_time and inspection.status == models.InspectionStatus.IN_PROGRESS:
        inspection.status = models.InspectionStatus.OVERDUE_COMPLETED
    else:
        inspection.status = models.InspectionStatus.COMPLETED
    
    inspection.actual_end_time = now
    if inspection.actual_start_time is None:
        inspection.actual_start_time = now
    inspection.problem_level = result.problem_level
    inspection.handling_suggestion = result.handling_suggestion
    
    db.commit()
    db.refresh(inspection)
    
    return inspection

@router.put("/{inspection_code}/cancel", response_model=schemas.InspectionResponse)
def cancel_inspection(inspection_code: str, db: Session = Depends(get_db)):
    inspection = db.query(models.Inspection).filter(models.Inspection.inspection_code == inspection_code).first()
    if not inspection:
        raise HTTPException(status_code=404, detail="巡检任务不存在")
    
    if inspection.status not in [models.InspectionStatus.PENDING, models.InspectionStatus.IN_PROGRESS]:
        raise HTTPException(status_code=400, detail="只有待开始或进行中的任务可以取消")
    
    inspection.status = models.InspectionStatus.CANCELLED
    db.commit()
    db.refresh(inspection)
    
    return inspection
