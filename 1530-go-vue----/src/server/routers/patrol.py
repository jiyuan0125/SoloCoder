from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date as date_type
from src.core.database import get_db
from src.core.services import PatrolService
from src.server.schemas import (
    PatrolPlanResponse,
    PatrolRecordCreate,
    PatrolRecordResponse
)


router = APIRouter(prefix="/patrol", tags=["patrol"])


@router.post("/plans/generate")
def generate_daily_plan(db: Session = Depends(get_db)):
    service = PatrolService(db)
    plans = service.create_daily_plan()
    return {"message": f"Generated {len(plans)} patrol plans", "plans_count": len(plans)}


@router.get("/plans/", response_model=List[PatrolPlanResponse])
def list_plans(
    plan_date: Optional[date_type] = None,
    db: Session = Depends(get_db)
):
    service = PatrolService(db)
    return service.list_plans(plan_date=plan_date)


@router.get("/plans/{plan_id}", response_model=PatrolPlanResponse)
def get_plan(plan_id: int, db: Session = Depends(get_db)):
    service = PatrolService(db)
    plan = service.get_plan(plan_id)
    if not plan:
        raise HTTPException(status_code=404, detail="Plan not found")
    return plan


@router.post("/records/", response_model=PatrolRecordResponse)
def add_patrol_record(record: PatrolRecordCreate, db: Session = Depends(get_db)):
    service = PatrolService(db)
    return service.add_patrol_record(record.model_dump())


@router.get("/records/", response_model=List[PatrolRecordResponse])
def list_records(
    segment_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    service = PatrolService(db)
    return service.list_records(segment_id=segment_id)
