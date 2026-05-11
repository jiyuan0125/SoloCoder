from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy import select

from ..database import get_db
from ..models import RadiationSource, RetirementApproval, ApprovalStatus, SourceStatus
from ..schemas import ApprovalCreate, ApprovalUpdate, ApprovalSubmit, ApprovalApprove, Approval as ApprovalSchema
from ..utils import log_operation

router = APIRouter(prefix="/approvals", tags=["退役审批管理"])


@router.get("", response_model=List[ApprovalSchema])
def list_approvals(
    page: int = Query(1, ge=1),
    size: int = Query(10, ge=1, le=100),
    status: Optional[ApprovalStatus] = None,
    source_id: Optional[int] = None,
    db: Session = Depends(get_db),
):
    stmt = select(RetirementApproval)
    if status:
        stmt = stmt.where(RetirementApproval.status == status)
    if source_id:
        stmt = stmt.where(RetirementApproval.source_id == source_id)
    stmt = stmt.offset((page - 1) * size).limit(size).order_by(RetirementApproval.id.desc())
    return db.execute(stmt).scalars().all()


@router.get("/{approval_id}", response_model=ApprovalSchema)
def get_approval(approval_id: int, db: Session = Depends(get_db)):
    stmt = select(RetirementApproval).where(RetirementApproval.id == approval_id)
    approval = db.execute(stmt).scalar_one_or_none()
    if not approval:
        raise HTTPException(status_code=404, detail="审批记录不存在")
    return approval


@router.post("", response_model=ApprovalSchema)
def create_approval(approval_in: ApprovalCreate, db: Session = Depends(get_db)):
    stmt = select(RadiationSource).where(RadiationSource.id == approval_in.source_id)
    source = db.execute(stmt).scalar_one_or_none()
    if not source:
        raise HTTPException(status_code=404, detail="放射源不存在")
    
    if source.status == SourceStatus.RETIRED:
        raise HTTPException(status_code=400, detail="该放射源已退役")
    
    approval = RetirementApproval(
        source_id=approval_in.source_id,
        source_code=source.source_code,
        plan_content=approval_in.plan_content,
        applicant=approval_in.applicant,
        status=ApprovalStatus.DRAFT,
    )
    db.add(approval)
    db.commit()
    db.refresh(approval)
    
    log_operation(db, "创建退役方案", approval_in.applicant, "RetirementApproval", approval.id,
                  f"创建退役方案: 放射源 {source.source_code}")
    
    return approval


@router.put("/{approval_id}", response_model=ApprovalSchema)
def update_approval(approval_id: int, approval_in: ApprovalUpdate, db: Session = Depends(get_db)):
    stmt = select(RetirementApproval).where(RetirementApproval.id == approval_id)
    approval = db.execute(stmt).scalar_one_or_none()
    if not approval:
        raise HTTPException(status_code=404, detail="审批记录不存在")
    
    if approval.status != ApprovalStatus.DRAFT:
        raise HTTPException(status_code=400, detail="只能修改草稿状态的方案")
    
    update_data = approval_in.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(approval, key, value)
    
    db.commit()
    db.refresh(approval)
    
    log_operation(db, "更新退役方案", approval.applicant, "RetirementApproval", approval.id,
                  f"更新退役方案内容")
    
    return approval


@router.post("/{approval_id}/submit", response_model=ApprovalSchema)
def submit_approval(approval_id: int, submit_in: ApprovalSubmit, db: Session = Depends(get_db)):
    stmt = select(RetirementApproval).where(RetirementApproval.id == approval_id)
    approval = db.execute(stmt).scalar_one_or_none()
    if not approval:
        raise HTTPException(status_code=404, detail="审批记录不存在")
    
    if approval.status != ApprovalStatus.DRAFT:
        raise HTTPException(status_code=400, detail="只能提交草稿状态的方案")
    
    approval.status = ApprovalStatus.SUBMITTED
    approval.applicant = submit_in.operator
    approval.apply_time = datetime.utcnow()
    
    db.commit()
    db.refresh(approval)
    
    log_operation(db, "提交退役审批", submit_in.operator, "RetirementApproval", approval.id,
                  f"提交退役方案审批: 放射源 {approval.source_code}")
    
    return approval


@router.post("/{approval_id}/approve", response_model=ApprovalSchema)
def approve_approval(approval_id: int, approve_in: ApprovalApprove, db: Session = Depends(get_db)):
    stmt = select(RetirementApproval).where(RetirementApproval.id == approval_id)
    approval = db.execute(stmt).scalar_one_or_none()
    if not approval:
        raise HTTPException(status_code=404, detail="审批记录不存在")
    
    if approval.status != ApprovalStatus.SUBMITTED:
        raise HTTPException(status_code=400, detail="只能审批待审批状态的方案")
    
    approval.status = ApprovalStatus.APPROVED if approve_in.approved else ApprovalStatus.REJECTED
    approval.approver = approve_in.operator
    approval.approval_time = datetime.utcnow()
    approval.approval_remarks = approve_in.remarks
    
    db.commit()
    db.refresh(approval)
    
    result_desc = "通过" if approve_in.approved else "驳回"
    log_operation(db, f"审批退役方案{result_desc}", approve_in.operator, "RetirementApproval", approval.id,
                  f"退役方案审批{result_desc}: 放射源 {approval.source_code}")
    
    return approval


@router.delete("/{approval_id}")
def delete_approval(approval_id: int, operator: str = Query(..., description="操作人"), db: Session = Depends(get_db)):
    stmt = select(RetirementApproval).where(RetirementApproval.id == approval_id)
    approval = db.execute(stmt).scalar_one_or_none()
    if not approval:
        raise HTTPException(status_code=404, detail="审批记录不存在")
    
    if approval.status == ApprovalStatus.APPROVED:
        raise HTTPException(status_code=400, detail="已通过的方案不能删除")
    
    source_code = approval.source_code
    db.delete(approval)
    db.commit()
    
    log_operation(db, "删除退役方案", operator, "RetirementApproval", approval_id,
                  f"删除退役方案: 放射源 {source_code}")
    
    return {"message": "删除成功"}
