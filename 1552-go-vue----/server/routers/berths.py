from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from .. import models, schemas, services
from ..database import get_db

router = APIRouter(prefix="/api/berths", tags=["berths"])


@router.post("/", response_model=schemas.Berth)
def create_berth(berth: schemas.BerthCreate, db: Session = Depends(get_db)):
    return services.create_berth(db, berth)


@router.get("/", response_model=List[schemas.Berth])
def list_berths(available: bool = None, db: Session = Depends(get_db)):
    query = db.query(models.Berth)
    if available is not None:
        query = query.filter(models.Berth.is_available == available)
    return query.all()


@router.get("/{berth_id}", response_model=schemas.Berth)
def get_berth(berth_id: int, db: Session = Depends(get_db)):
    berth = db.query(models.Berth).filter(models.Berth.id == berth_id).first()
    if not berth:
        raise HTTPException(status_code=404, detail="泊位不存在")
    return berth


@router.put("/{berth_id}", response_model=schemas.Berth)
def update_berth(berth_id: int, data: schemas.BerthUpdate, db: Session = Depends(get_db)):
    berth = db.query(models.Berth).filter(models.Berth.id == berth_id).first()
    if not berth:
        raise HTTPException(status_code=404, detail="泊位不存在")
    for key, value in data.model_dump(exclude_unset=True).items():
        setattr(berth, key, value)
    db.commit()
    db.refresh(berth)
    return berth


@router.delete("/{berth_id}")
def delete_berth(berth_id: int, db: Session = Depends(get_db)):
    berth = db.query(models.Berth).filter(models.Berth.id == berth_id).first()
    if not berth:
        raise HTTPException(status_code=404, detail="泊位不存在")
    db.delete(berth)
    db.commit()
    return {"message": "已删除"}
