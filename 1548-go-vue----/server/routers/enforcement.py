from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date

from ..database import get_db
from ..models import LawEnforcement, License, PenaltyType, PenaltyStatus
from ..schemas import (
    LawEnforcementCreate, LawEnforcementResponse
)
from ..services import (
    create_law_enforcement, check_annual_review_eligibility
)

router = APIRouter(tags=["执法管理"])


@router.post("/law-enforcements", response_model=LawEnforcementResponse)
def create_enforcement(
    enforcement_data: LawEnforcementCreate,
    db: Session = Depends(get_db)
):
    if enforcement_data.license_id:
        license = db.query(License).filter(License.id == enforcement_data.license_id).first()
        if not license:
            raise HTTPException(status_code=404, detail="关联的许可证不存在")
        if enforcement_data.company_name != license.company_name:
            raise HTTPException(status_code=400, detail="企业名称与许可证不符")
    
    return create_law_enforcement(db, enforcement_data)


@router.get("/law-enforcements", response_model=List[LawEnforcementResponse])
def list_enforcements(
    license_id: Optional[int] = None,
    company_name: Optional[str] = None,
    penalty_type: Optional[PenaltyType] = None,
    penalty_status: Optional[PenaltyStatus] = None,
    skip: int = Query(0, ge=0),
    limit: int = Query(20, ge=1, le=100),
    db: Session = Depends(get_db)
):
    query = db.query(LawEnforcement)
    if license_id:
        query = query.filter(LawEnforcement.license_id == license_id)
    if company_name:
        query = query.filter(LawEnforcement.company_name.contains(company_name))
    if penalty_type:
        query = query.filter(LawEnforcement.penalty_type == penalty_type)
    if penalty_status:
        query = query.filter(LawEnforcement.penalty_status == penalty_status)
    
    return query.order_by(LawEnforcement.created_at.desc()).offset(skip).limit(limit).all()


@router.get("/law-enforcements/{enforcement_id}", response_model=LawEnforcementResponse)
def get_enforcement(enforcement_id: int, db: Session = Depends(get_db)):
    enforcement = db.query(LawEnforcement).filter(LawEnforcement.id == enforcement_id).first()
    if not enforcement:
        raise HTTPException(status_code=404, detail="执法记录不存在")
    return enforcement


@router.post("/law-enforcements/{enforcement_id}/execute")
def execute_enforcement(enforcement_id: int, db: Session = Depends(get_db)):
    enforcement = db.query(LawEnforcement).filter(LawEnforcement.id == enforcement_id).first()
    if not enforcement:
        raise HTTPException(status_code=404, detail="执法记录不存在")
    if enforcement.penalty_status == PenaltyStatus.EXECUTED:
        raise HTTPException(status_code=400, detail="处罚已执行")
    
    from datetime import datetime
    enforcement.penalty_status = PenaltyStatus.EXECUTED
    enforcement.penalty_time = datetime.utcnow()
    db.commit()
    db.refresh(enforcement)
    
    return {"message": "处罚已执行", "enforcement_id": enforcement.id}


@router.get("/licenses/{license_id}/annual-review-check")
def check_annual_review(license_id: int, db: Session = Depends(get_db)):
    license = db.query(License).filter(License.id == license_id).first()
    if not license:
        raise HTTPException(status_code=404, detail="许可证不存在")
    
    return check_annual_review_eligibility(db, license_id)


@router.get("/companies/{company_name}/penalty-history")
def get_company_penalty_history(
    company_name: str,
    year: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(LawEnforcement).filter(
        LawEnforcement.company_name == company_name
    )
    if year:
        from sqlalchemy import func, and_
        query = query.filter(
            func.strftime('%Y', LawEnforcement.violation_time) == str(year)
        )
    
    records = query.order_by(LawEnforcement.violation_time.desc()).all()
    
    return {
        "company_name": company_name,
        "year": year,
        "penalty_count": len(records),
        "records": records
    }
