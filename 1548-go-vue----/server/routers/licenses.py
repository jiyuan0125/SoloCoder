from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional

from ..database import get_db
from ..models import License, LicenseStatus, AuditLog
from ..schemas import (
    LicenseCreate, LicenseUpdate, LicenseResponse,
    InspectRequest, PublishRequest, AuditLogResponse
)
from ..services import (
    create_license, submit_license, accept_license,
    inspect_license, publish_license, issue_license
)

router = APIRouter(prefix="/licenses", tags=["许可证管理"])


@router.post("", response_model=LicenseResponse)
def create_new_license(
    license_data: LicenseCreate,
    db: Session = Depends(get_db)
):
    return create_license(db, license_data)


@router.get("", response_model=dict)
def list_licenses(
    status: Optional[LicenseStatus] = None,
    company_name: Optional[str] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=100),
    db: Session = Depends(get_db)
):
    query = db.query(License)
    if status:
        query = query.filter(License.status == status)
    if company_name:
        query = query.filter(License.company_name.contains(company_name))
    
    total = query.count()
    licenses = query.offset(skip).limit(limit).all()
    return {"total": total, "items": licenses}


@router.get("/{license_id}", response_model=LicenseResponse)
def get_license(license_id: int, db: Session = Depends(get_db)):
    license = db.query(License).filter(License.id == license_id).first()
    if not license:
        raise HTTPException(status_code=404, detail="许可证不存在")
    return license


@router.put("/{license_id}", response_model=LicenseResponse)
def update_license(
    license_id: int,
    update_data: LicenseUpdate,
    db: Session = Depends(get_db)
):
    license = db.query(License).filter(License.id == license_id).first()
    if not license:
        raise HTTPException(status_code=404, detail="许可证不存在")
    if license.status != LicenseStatus.DRAFT:
        raise HTTPException(status_code=400, detail="只能在草稿状态修改")
    
    for key, value in update_data.model_dump(exclude_unset=True).items():
        setattr(license, key, value)
    db.commit()
    db.refresh(license)
    return license


@router.post("/{license_id}/submit", response_model=LicenseResponse)
def submit_license_for_approval(license_id: int, db: Session = Depends(get_db)):
    license = submit_license(db, license_id)
    if not license:
        raise HTTPException(status_code=400, detail="提交失败，请确认许可证状态")
    return license


@router.post("/{license_id}/accept", response_model=LicenseResponse)
def accept_license_application(license_id: int, db: Session = Depends(get_db)):
    license = accept_license(db, license_id)
    if not license:
        raise HTTPException(status_code=400, detail="受理失败，请确认许可证状态")
    return license


@router.post("/{license_id}/inspect", response_model=LicenseResponse)
def inspect_license_application(
    license_id: int,
    inspect_data: InspectRequest,
    db: Session = Depends(get_db)
):
    license = inspect_license(db, license_id, inspect_data.inspect_result, inspect_data.passed)
    if not license:
        raise HTTPException(status_code=400, detail="核查失败，请确认许可证状态")
    return license


@router.post("/{license_id}/publish", response_model=LicenseResponse)
def publish_license_application(
    license_id: int,
    publish_data: PublishRequest,
    db: Session = Depends(get_db)
):
    license = publish_license(db, license_id, publish_data.publish_days)
    if not license:
        raise HTTPException(status_code=400, detail="公示失败，请确认许可证状态")
    return license


@router.post("/{license_id}/issue", response_model=LicenseResponse)
def issue_license_application(license_id: int, db: Session = Depends(get_db)):
    license = issue_license(db, license_id)
    if not license:
        raise HTTPException(status_code=400, detail="发证失败，请确认许可证状态")
    return license


@router.get("/{license_id}/audit-logs", response_model=List[AuditLogResponse])
def get_license_audit_logs(
    license_id: int,
    db: Session = Depends(get_db)
):
    logs = db.query(AuditLog).filter(
        AuditLog.entity_type == "License",
        AuditLog.entity_id == license_id
    ).order_by(AuditLog.created_at.desc()).all()
    return logs
