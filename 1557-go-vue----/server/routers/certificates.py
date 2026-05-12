from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from .. import models, schemas, services
from ..database import get_db

router = APIRouter()


@router.post("/", response_model=schemas.Certificate)
def create_certificate(
    certificate: schemas.CertificateCreate, db: Session = Depends(get_db)
):
    vessel = (
        db.query(models.Vessel)
        .filter(models.Vessel.id == certificate.vessel_id)
        .first()
    )
    if not vessel:
        raise HTTPException(status_code=404, detail="船舶不存在")

    if certificate.inspection_id:
        inspection = (
            db.query(models.Inspection)
            .filter(models.Inspection.id == certificate.inspection_id)
            .first()
        )
        if not inspection:
            raise HTTPException(status_code=404, detail="检验记录不存在")

    db_certificate = models.Certificate(**certificate.model_dump())
    db.add(db_certificate)
    db.commit()
    db.refresh(db_certificate)

    services.mark_expired_todos_for_operation_period(db, certificate.vessel_id)

    return db_certificate


@router.get("/", response_model=List[schemas.Certificate])
def list_certificates(
    vessel_id: Optional[int] = Query(None),
    status: Optional[str] = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(models.Certificate)

    if vessel_id:
        query = query.filter(models.Certificate.vessel_id == vessel_id)
    if status:
        query = query.filter(models.Certificate.status == status)

    certificates = query.order_by(models.Certificate.issue_date.desc()).offset(skip).limit(limit).all()
    return certificates


@router.get("/{certificate_id}", response_model=schemas.Certificate)
def get_certificate(certificate_id: int, db: Session = Depends(get_db)):
    certificate = (
        db.query(models.Certificate)
        .filter(models.Certificate.id == certificate_id)
        .first()
    )
    if not certificate:
        raise HTTPException(status_code=404, detail="证书不存在")
    return certificate


@router.put("/{certificate_id}", response_model=schemas.Certificate)
def update_certificate(
    certificate_id: int,
    certificate_update: schemas.CertificateUpdate,
    db: Session = Depends(get_db),
):
    certificate = (
        db.query(models.Certificate)
        .filter(models.Certificate.id == certificate_id)
        .first()
    )
    if not certificate:
        raise HTTPException(status_code=404, detail="证书不存在")

    update_data = certificate_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(certificate, key, value)

    db.commit()
    db.refresh(certificate)

    services.mark_expired_todos_for_operation_period(db, certificate.vessel_id)

    return certificate


@router.post("/generate-reminders")
def generate_reminders(db: Session = Depends(get_db)):
    services.create_certificate_reminders(db)
    return {"message": "证书提醒生成成功"}


@router.post("/update-status")
def update_certificates_status(db: Session = Depends(get_db)):
    services.update_certificate_status(db)
    return {"message": "证书状态更新成功"}
