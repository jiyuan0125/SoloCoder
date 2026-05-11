from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from server.database import get_db
from server.models import RiverKeeper, River
from server.schemas import (
    RiverKeeperCreate, RiverKeeperUpdate, 
    RiverKeeper as RiverKeeperSchema, PatrolFrequency
)
from server.services import PATROL_FREQUENCIES

router = APIRouter(prefix="/keepers", tags=["keepers"])

@router.get("/frequencies", response_model=List[PatrolFrequency])
def get_patrol_frequencies():
    return list(PATROL_FREQUENCIES.values())

@router.get("/", response_model=List[RiverKeeperSchema])
def list_keepers(db: Session = Depends(get_db)):
    return db.query(RiverKeeper).all()

@router.post("/", response_model=RiverKeeperSchema)
def create_keeper(keeper: RiverKeeperCreate, db: Session = Depends(get_db)):
    if keeper.river_id:
        river = db.query(River).filter(River.id == keeper.river_id).first()
        if not river:
            raise HTTPException(status_code=400, detail="关联的河流不存在")
    db_keeper = RiverKeeper(**keeper.model_dump())
    db.add(db_keeper)
    db.commit()
    db.refresh(db_keeper)
    return db_keeper

@router.get("/{keeper_id}", response_model=RiverKeeperSchema)
def get_keeper(keeper_id: int, db: Session = Depends(get_db)):
    keeper = db.query(RiverKeeper).filter(RiverKeeper.id == keeper_id).first()
    if not keeper:
        raise HTTPException(status_code=404, detail="河长不存在")
    return keeper

@router.put("/{keeper_id}", response_model=RiverKeeperSchema)
def update_keeper(keeper_id: int, keeper: RiverKeeperUpdate, db: Session = Depends(get_db)):
    db_keeper = db.query(RiverKeeper).filter(RiverKeeper.id == keeper_id).first()
    if not db_keeper:
        raise HTTPException(status_code=404, detail="河长不存在")
    update_data = keeper.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_keeper, key, value)
    db.commit()
    db.refresh(db_keeper)
    return db_keeper

@router.delete("/{keeper_id}")
def delete_keeper(keeper_id: int, db: Session = Depends(get_db)):
    db_keeper = db.query(RiverKeeper).filter(RiverKeeper.id == keeper_id).first()
    if not db_keeper:
        raise HTTPException(status_code=404, detail="河长不存在")
    db.delete(db_keeper)
    db.commit()
    return {"message": "删除成功"}
