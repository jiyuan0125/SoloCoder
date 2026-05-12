from datetime import date
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from ..config import get_db
from ..models import Crew, Certificate, TrainingRecord
from ..schemas import (
    CertificateCreate, CertificateUpdate, CertificateResponse,
    TrainingRecordCreate, TrainingRecordResponse
)
from ..services import (
    calculate_certificate_status, check_training_interval
)

router = APIRouter(prefix="/api", tags=["certificates"])


@router.post("/certificates/", response_model=CertificateResponse, status_code=status.HTTP_201_CREATED)
def create_certificate(cert_data: CertificateCreate, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == cert_data.crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    existing = db.query(Certificate).filter(
        Certificate.certificate_number == cert_data.certificate_number
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="证书编号已存在")
    
    cert = Certificate(**cert_data.model_dump())
    db.add(cert)
    db.commit()
    db.refresh(cert)
    
    status_str, days_until = calculate_certificate_status(cert.expiry_date, date.today())
    return CertificateResponse(
        id=cert.id,
        crew_id=cert.crew_id,
        certificate_type=cert.certificate_type,
        certificate_number=cert.certificate_number,
        issue_date=cert.issue_date,
        expiry_date=cert.expiry_date,
        status=status_str,
        days_until_expiry=days_until,
        created_at=cert.created_at,
        updated_at=cert.updated_at
    )


@router.get("/certificates/", response_model=List[CertificateResponse])
def list_certificates(crew_id: Optional[int] = None, 
                      warning_level: Optional[str] = None,
                      db: Session = Depends(get_db)):
    query = db.query(Certificate)
    if crew_id:
        query = query.filter(Certificate.crew_id == crew_id)
    
    certificates = query.all()
    today = date.today()
    
    result = []
    for cert in certificates:
        status_str, days_until = calculate_certificate_status(cert.expiry_date, today)
        
        if warning_level:
            if warning_level == "yellow" and status_str != "黄色警告":
                continue
            elif warning_level == "orange" and status_str != "橙色警告":
                continue
            elif warning_level == "expired" and status_str != "过期":
                continue
        
        result.append(CertificateResponse(
            id=cert.id,
            crew_id=cert.crew_id,
            certificate_type=cert.certificate_type,
            certificate_number=cert.certificate_number,
            issue_date=cert.issue_date,
            expiry_date=cert.expiry_date,
            status=status_str,
            days_until_expiry=days_until,
            created_at=cert.created_at,
            updated_at=cert.updated_at
        ))
    
    return result


@router.get("/certificates/{cert_id}", response_model=CertificateResponse)
def get_certificate(cert_id: int, db: Session = Depends(get_db)):
    cert = db.query(Certificate).filter(Certificate.id == cert_id).first()
    if not cert:
        raise HTTPException(status_code=404, detail="证书不存在")
    
    status_str, days_until = calculate_certificate_status(cert.expiry_date, date.today())
    return CertificateResponse(
        id=cert.id,
        crew_id=cert.crew_id,
        certificate_type=cert.certificate_type,
        certificate_number=cert.certificate_number,
        issue_date=cert.issue_date,
        expiry_date=cert.expiry_date,
        status=status_str,
        days_until_expiry=days_until,
        created_at=cert.created_at,
        updated_at=cert.updated_at
    )


@router.put("/certificates/{cert_id}", response_model=CertificateResponse)
def update_certificate(cert_id: int, cert_data: CertificateUpdate, db: Session = Depends(get_db)):
    cert = db.query(Certificate).filter(Certificate.id == cert_id).first()
    if not cert:
        raise HTTPException(status_code=404, detail="证书不存在")
    
    update_data = cert_data.model_dump(exclude_unset=True)
    for field, value in update_data.items():
        setattr(cert, field, value)
    
    db.commit()
    db.refresh(cert)
    
    status_str, days_until = calculate_certificate_status(cert.expiry_date, date.today())
    return CertificateResponse(
        id=cert.id,
        crew_id=cert.crew_id,
        certificate_type=cert.certificate_type,
        certificate_number=cert.certificate_number,
        issue_date=cert.issue_date,
        expiry_date=cert.expiry_date,
        status=status_str,
        days_until_expiry=days_until,
        created_at=cert.created_at,
        updated_at=cert.updated_at
    )


@router.delete("/certificates/{cert_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_certificate(cert_id: int, db: Session = Depends(get_db)):
    cert = db.query(Certificate).filter(Certificate.id == cert_id).first()
    if not cert:
        raise HTTPException(status_code=404, detail="证书不存在")
    
    db.delete(cert)
    db.commit()


@router.post("/training/", response_model=TrainingRecordResponse, status_code=status.HTTP_201_CREATED)
def create_training_record(training_data: TrainingRecordCreate, db: Session = Depends(get_db)):
    crew = db.query(Crew).filter(Crew.id == training_data.crew_id).first()
    if not crew:
        raise HTTPException(status_code=404, detail="船员不存在")
    
    certs = db.query(Certificate).filter(
        Certificate.crew_id == training_data.crew_id,
        Certificate.certificate_type == training_data.training_type
    ).all()
    
    for cert in certs:
        if not check_training_interval(db, training_data.crew_id, training_data.training_type,
                                       training_data.training_date, cert):
            raise HTTPException(
                status_code=400, 
                detail=f"培训间隔超过证书有效期一半，不符合要求"
            )
    
    record = TrainingRecord(**training_data.model_dump())
    db.add(record)
    db.commit()
    db.refresh(record)
    return record


@router.get("/training/", response_model=List[TrainingRecordResponse])
def list_training_records(crew_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(TrainingRecord)
    if crew_id:
        query = query.filter(TrainingRecord.crew_id == crew_id)
    return query.order_by(TrainingRecord.training_date.desc()).all()


@router.get("/training/{record_id}", response_model=TrainingRecordResponse)
def get_training_record(record_id: int, db: Session = Depends(get_db)):
    record = db.query(TrainingRecord).filter(TrainingRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="培训记录不存在")
    return record


@router.delete("/training/{record_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_training_record(record_id: int, db: Session = Depends(get_db)):
    record = db.query(TrainingRecord).filter(TrainingRecord.id == record_id).first()
    if not record:
        raise HTTPException(status_code=404, detail="培训记录不存在")
    
    db.delete(record)
    db.commit()
