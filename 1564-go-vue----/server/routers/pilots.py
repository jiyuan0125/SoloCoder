from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Pilot as PilotModel
from ..schemas import Pilot as PilotSchema, PilotCreate, PilotUpdate, PilotSummary

router = APIRouter(prefix="/api/pilots", tags=["pilots"])


@router.get("", response_model=List[PilotSummary])
def list_pilots(
    skip: int = 0,
    limit: int = 100,
    status: Optional[str] = None,
    db: Session = Depends(get_db)
):
    query = db.query(PilotModel)
    if status:
        query = query.filter(PilotModel.status == status)
    return query.offset(skip).limit(limit).all()


@router.post("", response_model=PilotSchema)
def create_pilot(pilot: PilotCreate, db: Session = Depends(get_db)):
    existing = db.query(PilotModel).filter(PilotModel.employee_id == pilot.employee_id).first()
    if existing:
        raise HTTPException(status_code=400, detail="工号已存在")
    
    db_pilot = PilotModel(**pilot.model_dump())
    db.add(db_pilot)
    db.commit()
    db.refresh(db_pilot)
    return db_pilot


@router.get("/{pilot_id}", response_model=PilotSchema)
def get_pilot(pilot_id: int, db: Session = Depends(get_db)):
    pilot = db.query(PilotModel).filter(PilotModel.id == pilot_id).first()
    if not pilot:
        raise HTTPException(status_code=404, detail="飞行员不存在")
    return pilot


@router.put("/{pilot_id}", response_model=PilotSchema)
def update_pilot(pilot_id: int, pilot_update: PilotUpdate, db: Session = Depends(get_db)):
    pilot = db.query(PilotModel).filter(PilotModel.id == pilot_id).first()
    if not pilot:
        raise HTTPException(status_code=404, detail="飞行员不存在")
    
    update_data = pilot_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(pilot, key, value)
    
    db.commit()
    db.refresh(pilot)
    return pilot


@router.delete("/{pilot_id}")
def delete_pilot(pilot_id: int, db: Session = Depends(get_db)):
    pilot = db.query(PilotModel).filter(PilotModel.id == pilot_id).first()
    if not pilot:
        raise HTTPException(status_code=404, detail="飞行员不存在")
    
    db.delete(pilot)
    db.commit()
    return {"message": "删除成功"}
