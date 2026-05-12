from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models import Locomotive
from app.schemas import (
    LocomotiveCreate, LocomotiveResponse, LocomotiveUpdateKm,
    LocomotiveWithNextMaintenance
)

router = APIRouter(prefix="/api/locomotives", tags=["机车管理"])


@router.post("/", response_model=LocomotiveResponse)
def create_locomotive(locomotive: LocomotiveCreate, db: Session = Depends(get_db)):
    existing = db.query(Locomotive).filter(
        Locomotive.locomotive_number == locomotive.locomotive_number
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="机车号已存在")
    
    db_locomotive = Locomotive(
        locomotive_number=locomotive.locomotive_number,
        section_cycle_km=locomotive.section_cycle_km
    )
    db.add(db_locomotive)
    db.commit()
    db.refresh(db_locomotive)
    return db_locomotive


@router.get("/", response_model=List[LocomotiveResponse])
def get_locomotives(db: Session = Depends(get_db)):
    return db.query(Locomotive).all()


@router.get("/{locomotive_id}", response_model=LocomotiveResponse)
def get_locomotive(locomotive_id: int, db: Session = Depends(get_db)):
    locomotive = db.query(Locomotive).filter(Locomotive.id == locomotive_id).first()
    if not locomotive:
        raise HTTPException(status_code=404, detail="机车不存在")
    return locomotive


@router.put("/{locomotive_id}/km", response_model=LocomotiveWithNextMaintenance)
def update_locomotive_km(
    locomotive_id: int,
    km_update: LocomotiveUpdateKm,
    db: Session = Depends(get_db)
):
    locomotive = db.query(Locomotive).filter(Locomotive.id == locomotive_id).first()
    if not locomotive:
        raise HTTPException(status_code=404, detail="机车不存在")
    
    if km_update.current_km < locomotive.current_km:
        raise HTTPException(status_code=400, detail="走行公里数不能小于当前值")
    
    locomotive.current_km = km_update.current_km
    db.commit()
    db.refresh(locomotive)
    
    return _calculate_next_maintenance(locomotive)


@router.get("/{locomotive_id}/next-maintenance", response_model=LocomotiveWithNextMaintenance)
def get_next_maintenance(locomotive_id: int, db: Session = Depends(get_db)):
    locomotive = db.query(Locomotive).filter(Locomotive.id == locomotive_id).first()
    if not locomotive:
        raise HTTPException(status_code=404, detail="机车不存在")
    
    return _calculate_next_maintenance(locomotive)


def _calculate_next_maintenance(locomotive: Locomotive) -> LocomotiveWithNextMaintenance:
    next_section_km = locomotive.last_section_maintenance_km + locomotive.section_cycle_km
    next_factory_km = locomotive.last_factory_maintenance_km + 2400000.0
    
    is_section_overdue = locomotive.current_km > next_section_km
    is_factory_overdue = locomotive.current_km > next_factory_km
    
    overdue_km = 0.0
    if is_section_overdue:
        overdue_km = max(overdue_km, locomotive.current_km - next_section_km)
    if is_factory_overdue:
        overdue_km = max(overdue_km, locomotive.current_km - next_factory_km)
    
    return LocomotiveWithNextMaintenance(
        locomotive=LocomotiveResponse.model_validate(locomotive),
        next_section_km=next_section_km,
        next_factory_km=next_factory_km,
        is_section_overdue=is_section_overdue,
        is_factory_overdue=is_factory_overdue,
        overdue_km=overdue_km
    )
