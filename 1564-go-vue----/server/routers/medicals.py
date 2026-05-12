from typing import List, Optional
from datetime import date
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Pilot as PilotModel, Medical as MedicalModel
from ..schemas import Medical as MedicalSchema, MedicalCreate, MedicalUpdate
from ..validation_service import get_medical_expiry_date

router = APIRouter(prefix="/api/medicals", tags=["medicals"])


@router.get("", response_model=List[MedicalSchema])
def list_medicals(
    pilot_id: Optional[int] = None,
    result: Optional[str] = None,
    db: Session = Depends(get_db)
):
    query = db.query(MedicalModel)
    if pilot_id:
        query = query.filter(MedicalModel.pilot_id == pilot_id)
    if result:
        query = query.filter(MedicalModel.result == result)
    return query.order_by(MedicalModel.examination_date.desc()).all()


@router.post("/pilots/{pilot_id}", response_model=MedicalSchema)
def create_medical(
    pilot_id: int,
    medical: MedicalCreate,
    db: Session = Depends(get_db)
):
    pilot = db.query(PilotModel).filter(PilotModel.id == pilot_id).first()
    if not pilot:
        raise HTTPException(status_code=404, detail="飞行员不存在")
    
    medical_data = medical.model_dump()
    
    if medical.result != "qualified":
        pilot.status = "grounded"
    
    db_medical = MedicalModel(pilot_id=pilot_id, **medical_data)
    db.add(db_medical)
    db.commit()
    db.refresh(db_medical)
    if medical.result != "qualified":
        db.refresh(pilot)
    return db_medical


@router.get("/{medical_id}", response_model=MedicalSchema)
def get_medical(medical_id: int, db: Session = Depends(get_db)):
    medical = db.query(MedicalModel).filter(MedicalModel.id == medical_id).first()
    if not medical:
        raise HTTPException(status_code=404, detail="体检记录不存在")
    return medical


@router.put("/{medical_id}", response_model=MedicalSchema)
def update_medical(medical_id: int, medical_update: MedicalUpdate, db: Session = Depends(get_db)):
    medical = db.query(MedicalModel).filter(MedicalModel.id == medical_id).first()
    if not medical:
        raise HTTPException(status_code=404, detail="体检记录不存在")
    
    update_data = medical_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(medical, key, value)
    
    pilot = db.query(PilotModel).filter(PilotModel.id == medical.pilot_id).first()
    if pilot and "result" in update_data:
        if update_data["result"] == "qualified":
            if pilot.status == "grounded":
                pilot.status = "standby"
        else:
            pilot.status = "grounded"
    
    db.commit()
    db.refresh(medical)
    return medical


@router.delete("/{medical_id}")
def delete_medical(medical_id: int, db: Session = Depends(get_db)):
    medical = db.query(MedicalModel).filter(MedicalModel.id == medical_id).first()
    if not medical:
        raise HTTPException(status_code=404, detail="体检记录不存在")
    
    db.delete(medical)
    db.commit()
    return {"message": "删除成功"}
