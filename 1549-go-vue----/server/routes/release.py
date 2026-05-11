from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from ..database import get_db
from ..models import ReleaseActivity, SeedFarm
from ..schemas import (
    SeedFarmCreate, SeedFarmUpdate, SeedFarmResponse,
    ReleaseActivityCreate, ReleaseActivityResponse,
    MessageResponse
)

router = APIRouter(prefix="/api/release", tags=["增殖放流管理"])


@router.get("/seed-farms", response_model=List[SeedFarmResponse])
def list_seed_farms(
    skip: int = 0,
    limit: int = 100,
    name: Optional[str] = None,
    db: Session = Depends(get_db)
):
    query = db.query(SeedFarm)
    if name:
        query = query.filter(SeedFarm.name.contains(name))
    return query.offset(skip).limit(limit).all()


@router.post("/seed-farms", response_model=SeedFarmResponse)
def create_seed_farm(seed_farm: SeedFarmCreate, db: Session = Depends(get_db)):
    db_seed_farm = SeedFarm(**seed_farm.model_dump())
    db.add(db_seed_farm)
    db.commit()
    db.refresh(db_seed_farm)
    return db_seed_farm


@router.get("/seed-farms/{farm_id}", response_model=SeedFarmResponse)
def get_seed_farm(farm_id: int, db: Session = Depends(get_db)):
    seed_farm = db.query(SeedFarm).filter(SeedFarm.id == farm_id).first()
    if not seed_farm:
        raise HTTPException(status_code=404, detail="苗种场不存在")
    return seed_farm


@router.put("/seed-farms/{farm_id}", response_model=SeedFarmResponse)
def update_seed_farm(
    farm_id: int,
    seed_farm: SeedFarmUpdate,
    db: Session = Depends(get_db)
):
    db_seed_farm = db.query(SeedFarm).filter(SeedFarm.id == farm_id).first()
    if not db_seed_farm:
        raise HTTPException(status_code=404, detail="苗种场不存在")
    
    update_data = seed_farm.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_seed_farm, key, value)
    
    db.commit()
    db.refresh(db_seed_farm)
    return db_seed_farm


@router.delete("/seed-farms/{farm_id}", response_model=MessageResponse)
def delete_seed_farm(farm_id: int, db: Session = Depends(get_db)):
    seed_farm = db.query(SeedFarm).filter(SeedFarm.id == farm_id).first()
    if not seed_farm:
        raise HTTPException(status_code=404, detail="苗种场不存在")
    
    db.delete(seed_farm)
    db.commit()
    return {"message": "苗种场已删除"}


@router.get("/activities", response_model=List[ReleaseActivityResponse])
def list_activities(
    skip: int = 0,
    limit: int = 100,
    species: Optional[str] = None,
    seed_farm_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(ReleaseActivity)
    if species:
        query = query.filter(ReleaseActivity.species.contains(species))
    if seed_farm_id:
        query = query.filter(ReleaseActivity.seed_farm_id == seed_farm_id)
    return query.order_by(ReleaseActivity.release_date.desc()).offset(skip).limit(limit).all()


@router.post("/activities", response_model=ReleaseActivityResponse)
def create_activity(activity: ReleaseActivityCreate, db: Session = Depends(get_db)):
    if activity.seed_farm_id:
        seed_farm = db.query(SeedFarm).filter(
            SeedFarm.id == activity.seed_farm_id
        ).first()
        if not seed_farm:
            raise HTTPException(status_code=404, detail="苗种场不存在")
    
    existing = db.query(ReleaseActivity).filter(
        ReleaseActivity.activity_code == activity.activity_code
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="活动编号已存在")
    
    db_activity = ReleaseActivity(**activity.model_dump())
    db.add(db_activity)
    db.commit()
    db.refresh(db_activity)
    return db_activity


@router.get("/activities/{activity_id}", response_model=ReleaseActivityResponse)
def get_activity(activity_id: int, db: Session = Depends(get_db)):
    activity = db.query(ReleaseActivity).filter(
        ReleaseActivity.id == activity_id
    ).first()
    if not activity:
        raise HTTPException(status_code=404, detail="放流活动不存在")
    return activity


@router.delete("/activities/{activity_id}", response_model=MessageResponse)
def delete_activity(activity_id: int, db: Session = Depends(get_db)):
    activity = db.query(ReleaseActivity).filter(
        ReleaseActivity.id == activity_id
    ).first()
    if not activity:
        raise HTTPException(status_code=404, detail="放流活动不存在")
    
    db.delete(activity)
    db.commit()
    return {"message": "放流活动已删除"}
