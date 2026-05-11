from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from sqlalchemy.sql import func
from typing import List, Optional
from datetime import date, timedelta
from ..database import get_db
from ..models import (
    ViolationCase, Evidence, Appeal, Fisherman, FishingLicense,
    ViolationType, Severity, PenaltyType, CaseStatus, LicenseStatus
)
from ..schemas import (
    ViolationCaseCreate, ViolationCaseDecision, ViolationCaseResponse,
    EvidenceCreate, EvidenceResponse, AppealCreate, AppealResponse,
    MessageResponse
)

router = APIRouter(prefix="/api/violations", tags=["违规处理系统"])

FINE_STANDARDS = {
    (ViolationType.ILLEGAL_FISHING, Severity.MILD): 1000.0,
    (ViolationType.ILLEGAL_FISHING, Severity.MODERATE): 5000.0,
    (ViolationType.ILLEGAL_FISHING, Severity.SEVERE): 20000.0,
    (ViolationType.ILLEGAL_FISHING, Severity.CRITICAL): 50000.0,
    (ViolationType.CLOSED_SEASON, Severity.MILD): 2000.0,
    (ViolationType.CLOSED_SEASON, Severity.MODERATE): 10000.0,
    (ViolationType.CLOSED_SEASON, Severity.SEVERE): 30000.0,
    (ViolationType.CLOSED_SEASON, Severity.CRITICAL): 80000.0,
    (ViolationType.OVERSIZE_FISH, Severity.MILD): 500.0,
    (ViolationType.OVERSIZE_FISH, Severity.MODERATE): 2000.0,
    (ViolationType.OVERSIZE_FISH, Severity.SEVERE): 8000.0,
    (ViolationType.OVERSIZE_FISH, Severity.CRITICAL): 20000.0,
    (ViolationType.ILLEGAL_EQUIPMENT, Severity.MILD): 3000.0,
    (ViolationType.ILLEGAL_EQUIPMENT, Severity.MODERATE): 8000.0,
    (ViolationType.ILLEGAL_EQUIPMENT, Severity.SEVERE): 25000.0,
    (ViolationType.ILLEGAL_EQUIPMENT, Severity.CRITICAL): 60000.0,
    (ViolationType.NO_LICENSE, Severity.MILD): 2000.0,
    (ViolationType.NO_LICENSE, Severity.MODERATE): 10000.0,
    (ViolationType.NO_LICENSE, Severity.SEVERE): 40000.0,
    (ViolationType.NO_LICENSE, Severity.CRITICAL): 100000.0,
    (ViolationType.OTHER, Severity.MILD): 500.0,
    (ViolationType.OTHER, Severity.MODERATE): 2000.0,
    (ViolationType.OTHER, Severity.SEVERE): 10000.0,
    (ViolationType.OTHER, Severity.CRITICAL): 30000.0,
}

APPEAL_DAYS = 15


def calculate_fine(violation_type: ViolationType, severity: Severity) -> float:
    return FINE_STANDARDS.get((violation_type, severity), 1000.0)


@router.get("/", response_model=List[ViolationCaseResponse])
def list_cases(
    skip: int = 0,
    limit: int = 100,
    fisherman_id: Optional[int] = None,
    status: Optional[CaseStatus] = None,
    violation_type: Optional[ViolationType] = None,
    is_closed: Optional[bool] = None,
    db: Session = Depends(get_db)
):
    query = db.query(ViolationCase)
    if fisherman_id:
        query = query.filter(ViolationCase.fisherman_id == fisherman_id)
    if status:
        query = query.filter(ViolationCase.status == status)
    if violation_type:
        query = query.filter(ViolationCase.violation_type == violation_type)
    if is_closed is not None:
        query = query.filter(ViolationCase.is_closed == is_closed)
    return query.order_by(ViolationCase.created_at.desc()).offset(skip).limit(limit).all()


@router.get("/{case_id}", response_model=ViolationCaseResponse)
def get_case(case_id: int, db: Session = Depends(get_db)):
    case = db.query(ViolationCase).filter(ViolationCase.id == case_id).first()
    if not case:
        raise HTTPException(status_code=404, detail="案件不存在")
    return case


@router.post("/", response_model=ViolationCaseResponse)
def create_case(case: ViolationCaseCreate, db: Session = Depends(get_db)):
    fisherman = db.query(Fisherman).filter(
        Fisherman.id == case.fisherman_id
    ).first()
    if not fisherman:
        raise HTTPException(status_code=404, detail="渔民信息不存在")
    
    if case.license_id:
        license_ = db.query(FishingLicense).filter(
            FishingLicense.id == case.license_id
        ).first()
        if not license_:
            raise HTTPException(status_code=404, detail="许可证不存在")
    
    existing = db.query(ViolationCase).filter(
        ViolationCase.case_number == case.case_number
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="案件编号已存在")
    
    db_case = ViolationCase(
        **case.model_dump(),
        status=CaseStatus.FILED,
        is_closed=False
    )
    db.add(db_case)
    db.commit()
    db.refresh(db_case)
    return db_case


@router.post("/{case_id}/start-investigation", response_model=ViolationCaseResponse)
def start_investigation(case_id: int, db: Session = Depends(get_db)):
    case = db.query(ViolationCase).filter(ViolationCase.id == case_id).first()
    if not case:
        raise HTTPException(status_code=404, detail="案件不存在")
    
    if case.status != CaseStatus.FILED:
        raise HTTPException(status_code=400, detail="案件状态不是已立案")
    
    case.status = CaseStatus.INVESTIGATING
    db.commit()
    db.refresh(case)
    return case


@router.post("/{case_id}/evidence", response_model=EvidenceResponse)
def add_evidence(evidence: EvidenceCreate, db: Session = Depends(get_db)):
    case = db.query(ViolationCase).filter(
        ViolationCase.id == evidence.case_id
    ).first()
    if not case:
        raise HTTPException(status_code=404, detail="案件不存在")
    
    if case.status not in [CaseStatus.FILED, CaseStatus.INVESTIGATING]:
        raise HTTPException(
            status_code=400,
            detail="只能在立案或调查阶段补充证据"
        )
    
    db_evidence = Evidence(**evidence.model_dump())
    db.add(db_evidence)
    db.commit()
    db.refresh(db_evidence)
    return db_evidence


@router.get("/{case_id}/evidence", response_model=List[EvidenceResponse])
def list_evidence(case_id: int, db: Session = Depends(get_db)):
    return db.query(Evidence).filter(Evidence.case_id == case_id).all()


@router.post("/{case_id}/decide", response_model=ViolationCaseResponse)
def make_decision(
    case_id: int,
    decision: ViolationCaseDecision,
    db: Session = Depends(get_db)
):
    case = db.query(ViolationCase).filter(ViolationCase.id == case_id).first()
    if not case:
        raise HTTPException(status_code=404, detail="案件不存在")
    
    if case.status != CaseStatus.INVESTIGATING:
        raise HTTPException(status_code=400, detail="案件状态不是调查中")
    
    if decision.penalty_type == PenaltyType.FINE:
        if decision.fine_amount is None:
            fine = calculate_fine(case.violation_type, case.severity)
        else:
            fine = decision.fine_amount
        case.fine_amount = fine
    
    if decision.penalty_type == PenaltyType.SUSPEND_LICENSE:
        if decision.suspension_days is None or decision.suspension_days <= 0:
            raise HTTPException(status_code=400, detail="暂扣许可证需指定天数")
        case.suspension_days = decision.suspension_days
    
    case.penalty_type = decision.penalty_type
    case.decision_reason = decision.decision_reason
    case.decision_date = date.today()
    case.appeal_deadline = date.today() + timedelta(days=APPEAL_DAYS)
    case.status = CaseStatus.DECIDED
    
    db.commit()
    db.refresh(case)
    
    if case.license_id:
        license_ = db.query(FishingLicense).filter(
            FishingLicense.id == case.license_id
        ).first()
        if license_:
            if decision.penalty_type == PenaltyType.SUSPEND_LICENSE:
                license_.status = LicenseStatus.SUSPENDED
            elif decision.penalty_type == PenaltyType.REVOKE_LICENSE:
                license_.status = LicenseStatus.REVOKED
            db.commit()
    
    return case


@router.post("/{case_id}/appeal", response_model=AppealResponse)
def file_appeal(appeal: AppealCreate, db: Session = Depends(get_db)):
    case = db.query(ViolationCase).filter(
        ViolationCase.id == appeal.case_id
    ).first()
    if not case:
        raise HTTPException(status_code=404, detail="案件不存在")
    
    if case.status != CaseStatus.DECIDED:
        raise HTTPException(status_code=400, detail="只能对已作出处罚决定的案件申请复议")
    
    if case.appeal_deadline and date.today() > case.appeal_deadline:
        raise HTTPException(status_code=400, detail="复议申请已超期")
    
    db_appeal = Appeal(**appeal.model_dump())
    db.add(db_appeal)
    case.status = CaseStatus.UNDER_APPEAL
    db.commit()
    db.refresh(db_appeal)
    return db_appeal


@router.post("/{case_id}/appeal-decision", response_model=AppealResponse)
def decide_appeal(
    case_id: int,
    decision: str = "维持原决定",
    decision_reason: str = "",
    db: Session = Depends(get_db)
):
    case = db.query(ViolationCase).filter(ViolationCase.id == case_id).first()
    if not case:
        raise HTTPException(status_code=404, detail="案件不存在")
    
    if case.status != CaseStatus.UNDER_APPEAL:
        raise HTTPException(status_code=400, detail="案件状态不是复议中")
    
    appeal = db.query(Appeal).filter(
        Appeal.case_id == case_id
    ).order_by(Appeal.id.desc()).first()
    if not appeal:
        raise HTTPException(status_code=404, detail="复议记录不存在")
    
    appeal.decision = decision
    appeal.decision_date = date.today()
    appeal.decision_reason = decision_reason
    
    case.status = CaseStatus.EXECUTING
    db.commit()
    db.refresh(appeal)
    return appeal


@router.post("/{case_id}/execute", response_model=ViolationCaseResponse)
def execute_penalty(case_id: int, db: Session = Depends(get_db)):
    case = db.query(ViolationCase).filter(ViolationCase.id == case_id).first()
    if not case:
        raise HTTPException(status_code=404, detail="案件不存在")
    
    if case.status not in [CaseStatus.DECIDED, CaseStatus.UNDER_APPEAL]:
        raise HTTPException(status_code=400, detail="案件状态不适合执行")
    
    case.status = CaseStatus.EXECUTING
    db.commit()
    db.refresh(case)
    return case


@router.post("/{case_id}/close", response_model=ViolationCaseResponse)
def close_case(case_id: int, db: Session = Depends(get_db)):
    case = db.query(ViolationCase).filter(ViolationCase.id == case_id).first()
    if not case:
        raise HTTPException(status_code=404, detail="案件不存在")
    
    if case.status != CaseStatus.EXECUTING:
        raise HTTPException(status_code=400, detail="案件状态不是执行中")
    
    case.status = CaseStatus.CLOSED
    case.is_closed = True
    case.closed_date = date.today()
    db.commit()
    db.refresh(case)
    return case


@router.delete("/{case_id}", response_model=MessageResponse)
def delete_case(case_id: int, db: Session = Depends(get_db)):
    case = db.query(ViolationCase).filter(ViolationCase.id == case_id).first()
    if not case:
        raise HTTPException(status_code=404, detail="案件不存在")
    
    db.delete(case)
    db.commit()
    return {"message": "案件已删除"}


@router.get("/fisherman/{fisherman_id}/open", response_model=List[ViolationCaseResponse])
def list_open_violations(fisherman_id: int, db: Session = Depends(get_db)):
    return db.query(ViolationCase).filter(
        ViolationCase.fisherman_id == fisherman_id,
        ViolationCase.is_closed == False
    ).all()
