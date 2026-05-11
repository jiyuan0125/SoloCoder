from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date
from ..database import get_db
from ..models import Fisherman, ViolationCase, CaseStatus
from ..schemas import (
    FishermanCreate, FishermanUpdate, FishermanResponse, MessageResponse
)

router = APIRouter(prefix="/api/fishermen", tags=["渔民信息管理"])


@router.get("/", response_model=List[FishermanResponse])
def list_fishermen(
    skip: int = 0,
    limit: int = 100,
    name: Optional[str] = None,
    id_card: Optional[str] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Fisherman)
    if name:
        query = query.filter(Fisherman.name.contains(name))
    if id_card:
        query = query.filter(Fisherman.id_card.contains(id_card))
    return query.offset(skip).limit(limit).all()


@router.get("/{fisherman_id}", response_model=FishermanResponse)
def get_fisherman(fisherman_id: int, db: Session = Depends(get_db)):
    fisherman = db.query(Fisherman).filter(Fisherman.id == fisherman_id).first()
    if not fisherman:
        raise HTTPException(status_code=404, detail="渔民信息不存在")
    return fisherman


@router.post("/", response_model=FishermanResponse)
def create_fisherman(fisherman: FishermanCreate, db: Session = Depends(get_db)):
    existing = db.query(Fisherman).filter(
        Fisherman.id_card == fisherman.id_card
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="该身份证已注册")
    
    if fisherman.vessel_registration:
        existing_vessel = db.query(Fisherman).filter(
            Fisherman.vessel_registration == fisherman.vessel_registration
        ).first()
        if existing_vessel:
            raise HTTPException(status_code=400, detail="该船舶登记号已存在")
    
    db_fisherman = Fisherman(**fisherman.model_dump())
    db.add(db_fisherman)
    db.commit()
    db.refresh(db_fisherman)
    return db_fisherman


@router.put("/{fisherman_id}", response_model=FishermanResponse)
def update_fisherman(
    fisherman_id: int,
    fisherman: FishermanUpdate,
    db: Session = Depends(get_db)
):
    db_fisherman = db.query(Fisherman).filter(Fisherman.id == fisherman_id).first()
    if not db_fisherman:
        raise HTTPException(status_code=404, detail="渔民信息不存在")
    
    update_data = fisherman.model_dump(exclude_unset=True)
    
    if update_data.get("vessel_registration"):
        existing = db.query(Fisherman).filter(
            Fisherman.vessel_registration == update_data["vessel_registration"],
            Fisherman.id != fisherman_id
        ).first()
        if existing:
            raise HTTPException(status_code=400, detail="该船舶登记号已存在")
    
    for key, value in update_data.items():
        setattr(db_fisherman, key, value)
    
    db.commit()
    db.refresh(db_fisherman)
    return db_fisherman


@router.delete("/{fisherman_id}", response_model=MessageResponse)
def delete_fisherman(fisherman_id: int, db: Session = Depends(get_db)):
    db_fisherman = db.query(Fisherman).filter(Fisherman.id == fisherman_id).first()
    if not db_fisherman:
        raise HTTPException(status_code=404, detail="渔民信息不存在")
    
    db.delete(db_fisherman)
    db.commit()
    return {"message": "渔民信息已删除"}
