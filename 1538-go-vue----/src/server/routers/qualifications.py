from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from src.core.database import get_db
from src.core.models import Qualification, Unit
from src.core.schemas import QualificationCreate, QualificationUpdate, QualificationResponse
from src.core.enums import UnitType

router = APIRouter()


@router.post("", response_model=QualificationResponse, status_code=status.HTTP_201_CREATED)
def create_qualification(q_data: QualificationCreate, db: Session = Depends(get_db)):
    disposer = db.query(Unit).filter(
        Unit.id == q_data.disposer_id,
        Unit.unit_type == UnitType.DISPOSER
    ).first()
    if not disposer:
        raise HTTPException(status_code=400, detail="处置单位不存在或不是处置单位")

    if q_data.valid_from > q_data.valid_until:
        raise HTTPException(status_code=400, detail="开始日期不能晚于结束日期")

    qualification = Qualification(**q_data.model_dump())
    db.add(qualification)
    db.commit()
    db.refresh(qualification)
    return qualification


@router.get("", response_model=List[QualificationResponse])
def list_qualifications(
    disposer_id: int = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Qualification)
    if disposer_id:
        query = query.filter(Qualification.disposer_id == disposer_id)
    return query.offset(skip).limit(limit).all()


@router.get("/{q_id}", response_model=QualificationResponse)
def get_qualification(q_id: int, db: Session = Depends(get_db)):
    q = db.query(Qualification).filter(Qualification.id == q_id).first()
    if not q:
        raise HTTPException(status_code=404, detail="资质不存在")
    return q


@router.put("/{q_id}", response_model=QualificationResponse)
def update_qualification(
    q_id: int,
    q_data: QualificationUpdate,
    db: Session = Depends(get_db)
):
    qualification = db.query(Qualification).filter(Qualification.id == q_id).first()
    if not qualification:
        raise HTTPException(status_code=404, detail="资质不存在")

    update_data = q_data.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(qualification, key, value)

    db.commit()
    db.refresh(qualification)
    return qualification


@router.delete("/{q_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_qualification(q_id: int, db: Session = Depends(get_db)):
    qualification = db.query(Qualification).filter(Qualification.id == q_id).first()
    if not qualification:
        raise HTTPException(status_code=404, detail="资质不存在")
    db.delete(qualification)
    db.commit()
