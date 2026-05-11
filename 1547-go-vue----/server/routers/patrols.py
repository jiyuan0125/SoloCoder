from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from server.database import get_db
from server.models import Patrol, River, RiverKeeper
from server.schemas import (
    PatrolCreate, PatrolUpdate, 
    Patrol as PatrolSchema, PatrolDetail
)

router = APIRouter(prefix="/patrols", tags=["patrols"])

@router.get("/", response_model=List[PatrolSchema])
def list_patrols(db: Session = Depends(get_db)):
    return db.query(Patrol).order_by(Patrol.patrol_date.desc()).all()

@router.post("/", response_model=PatrolSchema)
def create_patrol(patrol: PatrolCreate, db: Session = Depends(get_db)):
    river = db.query(River).filter(River.id == patrol.river_id).first()
    if not river:
        raise HTTPException(status_code=400, detail="河流不存在")
    keeper = db.query(RiverKeeper).filter(RiverKeeper.id == patrol.keeper_id).first()
    if not keeper:
        raise HTTPException(status_code=400, detail="河长不存在")
    db_patrol = Patrol(**patrol.model_dump())
    db.add(db_patrol)
    db.commit()
    db.refresh(db_patrol)
    return db_patrol

@router.get("/{patrol_id}", response_model=PatrolDetail)
def get_patrol(patrol_id: int, db: Session = Depends(get_db)):
    patrol = db.query(Patrol).filter(Patrol.id == patrol_id).first()
    if not patrol:
        raise HTTPException(status_code=404, detail="巡河记录不存在")
    return patrol

@router.put("/{patrol_id}", response_model=PatrolSchema)
def update_patrol(patrol_id: int, patrol: PatrolUpdate, db: Session = Depends(get_db)):
    db_patrol = db.query(Patrol).filter(Patrol.id == patrol_id).first()
    if not db_patrol:
        raise HTTPException(status_code=404, detail="巡河记录不存在")
    update_data = patrol.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_patrol, key, value)
    db.commit()
    db.refresh(db_patrol)
    return db_patrol

@router.delete("/{patrol_id}")
def delete_patrol(patrol_id: int, db: Session = Depends(get_db)):
    db_patrol = db.query(Patrol).filter(Patrol.id == patrol_id).first()
    if not db_patrol:
        raise HTTPException(status_code=404, detail="巡河记录不存在")
    db.delete(db_patrol)
    db.commit()
    return {"message": "删除成功"}
