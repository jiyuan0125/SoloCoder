from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from src.core.database import get_db
from src.core.models import Unit
from src.core.schemas import UnitCreate, UnitUpdate, UnitResponse
from src.core.enums import UnitType

router = APIRouter()


@router.post("", response_model=UnitResponse, status_code=status.HTTP_201_CREATED)
def create_unit(unit_data: UnitCreate, db: Session = Depends(get_db)):
    unit = Unit(**unit_data.model_dump())
    db.add(unit)
    db.commit()
    db.refresh(unit)
    return unit


@router.get("", response_model=List[UnitResponse])
def list_units(
    unit_type: UnitType = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Unit)
    if unit_type:
        query = query.filter(Unit.unit_type == unit_type.value)
    return query.offset(skip).limit(limit).all()


@router.get("/{unit_id}", response_model=UnitResponse)
def get_unit(unit_id: int, db: Session = Depends(get_db)):
    unit = db.query(Unit).filter(Unit.id == unit_id).first()
    if not unit:
        raise HTTPException(status_code=404, detail="单位不存在")
    return unit


@router.put("/{unit_id}", response_model=UnitResponse)
def update_unit(
    unit_id: int,
    unit_data: UnitUpdate,
    db: Session = Depends(get_db)
):
    unit = db.query(Unit).filter(Unit.id == unit_id).first()
    if not unit:
        raise HTTPException(status_code=404, detail="单位不存在")

    update_data = unit_data.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(unit, key, value)

    db.commit()
    db.refresh(unit)
    return unit


@router.delete("/{unit_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_unit(unit_id: int, db: Session = Depends(get_db)):
    unit = db.query(Unit).filter(Unit.id == unit_id).first()
    if not unit:
        raise HTTPException(status_code=404, detail="单位不存在")
    db.delete(unit)
    db.commit()
