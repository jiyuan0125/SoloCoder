from datetime import date
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from ..config import get_db
from ..models import Crew, CrewStatus, Certificate, TrainingRecord
from ..schemas import (
    CrewCreate, CrewUpdate, CrewResponse, CrewDetailResponse,
    CertificateResponse, TrainingRecordResponse,
    MonthlyAttendanceSummary, AssignmentResponse, SalaryRecordResponse
)
from ..services import (
    calculate_certificate_status, check_crew_certificates_valid,
    get_monthly_attendance_summary
)

router = APIRouter(prefix="/api/crew", tags=["crew"])


@router.post("/", response_model=CrewResponse, status_code=status.HTTP_201_CREATED)
def create_crew(crew_data: CrewCreate, db: Session = Depends(get_db)):
    existing = db.query(Crew).filter(Crew.id_number == crew_data.id_number).first()
    if existing:
        raise HTTPException(status_code=400, detail="身份证号已存在")
    
    crew = Crew(**crew_data.model_dump())
    db.add(crew)
    db.commit()
    db.refresh(crew)
    return crew


@router.get("/", response_model=List[CrewResponse])
def list_crew(status: Optional[str] = None, db: Session = Depends(get_db)):
    query = db.query(Crew)
    if status:
        query = query.filter(Crew.status == status)
    return query.all()


@router.get("/{crew_id}", response_model=CrewDetailResponse)
def get_crew_detail(crew_id: int, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    today = date.today()
    certs_with_status = []
    for cert in crew.certificates:
        status_str, days_until = calculate_certificate_status(cert.expiry_date, today)
        cert_dict = {
            "id": cert.id,
            "crew_id": cert.crew_id,
            "certificate_type": cert.certificate_type,
            "certificate_number": cert.certificate_number,
            "issue_date": cert.issue_date,
            "expiry_date": cert.expiry_date,
            "status": status_str,
            "days_until_expiry": days_until,
            "created_at": cert.created_at,
            "updated_at": cert.updated_at
        }
        certs_with_status.append(CertificateResponse(**cert_dict))
    
    current_month_summary = get_monthly_attendance_summary(db, crew_id, today.year, today.month)
    
    return CrewDetailResponse(
        id=crew.id,
        name=crew.name,
        id_number=crew.id_number,
        phone=crew.phone,
        email=crew.email,
        is_intern=crew.is_intern,
        status=crew.status,
        created_at=crew.created_at,
        updated_at=crew.updated_at,
        certificates=certs_with_status,
        training_records=[TrainingRecordResponse.model_validate(tr) for tr in crew.training_records],
        assignments=[AssignmentResponse.model_validate(a) for a in crew.assignments],
        attendance_summary=MonthlyAttendanceSummary(**current_month_summary) if current_month_summary else None
    )


@router.put("/{crew_id}", response_model=CrewResponse)
def update_crew(crew_id: int, crew_data: CrewUpdate, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    update_data = crew_data.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(crew, field, value)
    
    db.commit()
    db.refresh(crew)
    return crew


@router.delete("/{crew_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_crew(crew_id: int, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    if crew.status == CrewStatus.ON_BOARD.value:
        raise HTTPException(status_code=400, detail="在船船员不能删除")
    
    db.delete(crew)
    db.commit()


@router.get("/{crew_id}/certificate-status")
def get_crew_certificate_status(crew_id: int, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    all_valid, summary = check_crew_certificates_valid(db, crew_id)
    
    return {
        "crew_id": crew_id,
        "crew_name": crew.name,
        "all_certificates_valid": all_valid,
        "certificates": summary
    }


@router.get("/{crew_id}/salary-history", response_model=List[SalaryRecordResponse])
def get_crew_salary_history(crew_id: int, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    return crew.salary_records


@router.get("/{crew_id}/attendance-summary/{year}/{month}")
def get_crew_attendance_summary(crew_id: int, year: int, month: int, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    return get_monthly_attendance_summary(db, crew_id, year, month)
