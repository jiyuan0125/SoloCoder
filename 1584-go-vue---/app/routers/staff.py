from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models import Staff, Qualification
from app.schemas import StaffCreate, StaffResponse, QualificationCreate, QualificationResponse

router = APIRouter()


@router.post("/", response_model=StaffResponse)
def create_staff(staff: StaffCreate, db: Session = Depends(get_db)):
    existing = db.query(Staff).filter(Staff.employee_id == staff.employee_id).first()
    if existing:
        raise HTTPException(status_code=400, detail="工号已存在")
    
    db_staff = Staff(
        name=staff.name,
        employee_id=staff.employee_id,
        join_date=staff.join_date,
        department=staff.department
    )
    db.add(db_staff)
    db.commit()
    db.refresh(db_staff)
    return db_staff


@router.get("/", response_model=List[StaffResponse])
def list_staff(db: Session = Depends(get_db)):
    return db.query(Staff).all()


@router.get("/{staff_id}", response_model=StaffResponse)
def get_staff(staff_id: int, db: Session = Depends(get_db)):
    staff = db.query(Staff).filter(Staff.id == staff_id).first()
    if not staff:
        raise HTTPException(status_code=404, detail="人员不存在")
    return staff


@router.post("/{staff_id}/qualifications", response_model=QualificationResponse)
def add_qualification(
    staff_id: int,
    qualification: QualificationCreate,
    db: Session = Depends(get_db)
):
    staff = db.query(Staff).filter(Staff.id == staff_id).first()
    if not staff:
        raise HTTPException(status_code=404, detail="人员不存在")
    
    if qualification.staff_id != staff_id:
        raise HTTPException(status_code=400, detail="人员ID不匹配")
    
    db_qual = Qualification(
        staff_id=staff_id,
        type=qualification.type,
        valid_from=qualification.valid_from,
        valid_to=qualification.valid_to
    )
    db.add(db_qual)
    db.commit()
    db.refresh(db_qual)
    return db_qual


@router.get("/{staff_id}/qualifications", response_model=List[QualificationResponse])
def list_qualifications(staff_id: int, db: Session = Depends(get_db)):
    staff = db.query(Staff).filter(Staff.id == staff_id).first()
    if not staff:
        raise HTTPException(status_code=404, detail="人员不存在")
    
    return db.query(Qualification).filter(Qualification.staff_id == staff_id).all()


@router.delete("/qualifications/{qualification_id}")
def delete_qualification(qualification_id: int, db: Session = Depends(get_db)):
    qual = db.query(Qualification).filter(Qualification.id == qualification_id).first()
    if not qual:
        raise HTTPException(status_code=404, detail="资质记录不存在")
    
    db.delete(qual)
    db.commit()
    return {"message": "资质记录已删除"}
