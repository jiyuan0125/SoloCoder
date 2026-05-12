from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from sqlalchemy import select
from typing import List
from server.database import get_db
from server.models import Pilot
from server.schemas import PilotCreate, PilotResponse
from server.services import create_pilot as service_create_pilot

router = APIRouter(
    prefix="/pilots",
    tags=["pilots"]
)


@router.post("", response_model=PilotResponse)
def create_pilot(pilot_data: PilotCreate, db: Session = Depends(get_db)):
    return service_create_pilot(db, pilot_data)


@router.get("", response_model=List[PilotResponse])
def get_pilots(db: Session = Depends(get_db)):
    return db.execute(select(Pilot)).scalars().all()


@router.get("/{pilot_id}", response_model=PilotResponse)
def get_pilot(pilot_id: int, db: Session = Depends(get_db)):
    pilot = db.execute(
        select(Pilot).where(Pilot.id == pilot_id)
    ).scalar()
    
    if not pilot:
        raise HTTPException(status_code=404, detail="引航员不存在")
    
    return pilot


@router.put("/{pilot_id}/duty", response_model=PilotResponse)
def toggle_duty_status(pilot_id: int, db: Session = Depends(get_db)):
    pilot = db.execute(
        select(Pilot).where(Pilot.id == pilot_id)
    ).scalar()
    
    if not pilot:
        raise HTTPException(status_code=404, detail="引航员不存在")
    
    pilot.is_on_duty = not pilot.is_on_duty
    db.commit()
    db.refresh(pilot)
    
    return pilot
