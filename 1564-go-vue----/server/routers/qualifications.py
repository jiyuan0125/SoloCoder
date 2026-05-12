from typing import List, Optional
from datetime import date
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Pilot as PilotModel, Qualification as QualificationModel
from ..schemas import Qualification as QualificationSchema, QualificationCreate, QualificationUpdate

router = APIRouter(prefix="/api/qualifications", tags=["qualifications"])


@router.get("", response_model=List[QualificationSchema])
def list_qualifications(
    pilot_id: Optional[int] = None,
    is_valid: Optional[bool] = None,
    db: Session = Depends(get_db)
):
    query = db.query(QualificationModel)
    if pilot_id:
        query = query.filter(QualificationModel.pilot_id == pilot_id)
    if is_valid is not None:
        query = query.filter(QualificationModel.is_valid == is_valid)
    return query.all()


@router.post("", response_model=QualificationSchema)
def create_qualification(qualification: QualificationCreate, db: Session = Depends(get_db)):
    pilot = db.query(PilotModel).filter(PilotModel.id == qualification.pilot_id).first() if hasattr(qualification, 'pilot_id') else None
    db_qual = QualificationModel(**qualification.model_dump())
    db.add(db_qual)
    db.commit()
    db.refresh(db_qual)
    return db_qual


@router.post("/pilots/{pilot_id}", response_model=QualificationSchema)
def create_pilot_qualification(
    pilot_id: int,
    qualification: QualificationCreate,
    db: Session = Depends(get_db)
):
    pilot = db.query(PilotModel).filter(PilotModel.id == pilot_id).first()
    if not pilot:
        raise HTTPException(status_code=404, detail="飞行员不存在")
    
    db_qual = QualificationModel(pilot_id=pilot_id, **qualification.model_dump())
    db.add(db_qual)
    db.commit()
    db.refresh(db_qual)
    return db_qual


@router.get("/{qual_id}", response_model=QualificationSchema)
def get_qualification(qual_id: int, db: Session = Depends(get_db)):
    qual = db.query(QualificationModel).filter(QualificationModel.id == qual_id).first()
    if not qual:
        raise HTTPException(status_code=404, detail="资质不存在")
    return qual


@router.put("/{qual_id}", response_model=QualificationSchema)
def update_qualification(qual_id: int, qual_update: QualificationUpdate, db: Session = Depends(get_db)):
    qual = db.query(QualificationModel).filter(QualificationModel.id == qual_id).first()
    if not qual:
        raise HTTPException(status_code=404, detail="资质不存在")
    
    update_data = qual_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(qual, key, value)
    
    db.commit()
    db.refresh(qual)
    return qual


@router.delete("/{qual_id}")
def delete_qualification(qual_id: int, db: Session = Depends(get_db)):
    qual = db.query(QualificationModel).filter(QualificationModel.id == qual_id).first()
    if not qual:
        raise HTTPException(status_code=404, detail="资质不存在")
    
    db.delete(qual)
    db.commit()
    return {"message": "删除成功"}
