from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy import func, and_
from typing import List, Optional
from datetime import datetime, timedelta

from database import get_db
from models import Complaint, ComplaintStatus, ComplaintSeverity
from schemas import (
    ComplaintCreate,
    ComplaintUpdate,
    ComplaintResponse
)
from config import settings

router = APIRouter(prefix="/complaints", tags=["投诉管理"])


@router.post("/", response_model=ComplaintResponse, summary="登记投诉")
def create_complaint(
    complaint: ComplaintCreate,
    db: Session = Depends(get_db)
):
    db_complaint = Complaint(
        complainant_name=complaint.complainant_name,
        complainant_phone=complaint.complainant_phone,
        content=complaint.content,
        severity=complaint.severity,
        status=ComplaintStatus.REGISTERED
    )
    
    db.add(db_complaint)
    db.commit()
    db.refresh(db_complaint)
    return db_complaint


@router.get("/", response_model=List[ComplaintResponse], summary="查询投诉列表")
def list_complaints(
    status: Optional[str] = Query(None, description="状态筛选: registered/processing/pending_confirmation/closed"),
    severity: Optional[str] = Query(None, description="严重程度筛选: normal/serious"),
    is_overdue: Optional[bool] = Query(None),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Complaint)
    
    if status:
        status_map = {
            "registered": ComplaintStatus.REGISTERED,
            "processing": ComplaintStatus.PROCESSING,
            "pending_confirmation": ComplaintStatus.PENDING_CONFIRMATION,
            "closed": ComplaintStatus.CLOSED
        }
        if status in status_map:
            query = query.filter(Complaint.status == status_map[status])
    
    if severity:
        severity_map = {
            "normal": ComplaintSeverity.NORMAL,
            "serious": ComplaintSeverity.SERIOUS
        }
        if severity in severity_map:
            query = query.filter(Complaint.severity == severity_map[severity])
    
    if is_overdue is not None:
        query = query.filter(Complaint.is_overdue == (1 if is_overdue else 0))
    
    return query.order_by(Complaint.created_at.desc()).offset(skip).limit(limit).all()


@router.get("/{complaint_id}", response_model=ComplaintResponse, summary="查询单条投诉")
def get_complaint(complaint_id: int, db: Session = Depends(get_db)):
    complaint = db.query(Complaint).filter(Complaint.id == complaint_id).first()
    if not complaint:
        raise HTTPException(status_code=404, detail="投诉记录不存在")
    return complaint


@router.post("/{complaint_id}/start-processing", response_model=ComplaintResponse, summary="开始处理投诉")
def start_processing(
    complaint_id: int,
    handler: str = "",
    db: Session = Depends(get_db)
):
    complaint = db.query(Complaint).filter(Complaint.id == complaint_id).first()
    if not complaint:
        raise HTTPException(status_code=404, detail="投诉记录不存在")
    
    if complaint.status != ComplaintStatus.REGISTERED:
        raise HTTPException(
            status_code=400, 
            detail=f"当前状态为{complaint.status.value}，无法开始处理"
        )
    
    complaint.status = ComplaintStatus.PROCESSING
    complaint.processing_at = datetime.utcnow()
    if handler:
        complaint.current_handler = handler
    
    db.commit()
    db.refresh(complaint)
    return complaint


@router.post("/{complaint_id}/send-response", response_model=ComplaintResponse, summary="发送回复并等待旅客确认")
def send_response(
    complaint_id: int,
    response_content: str,
    db: Session = Depends(get_db)
):
    complaint = db.query(Complaint).filter(Complaint.id == complaint_id).first()
    if not complaint:
        raise HTTPException(status_code=404, detail="投诉记录不存在")
    
    if complaint.status != ComplaintStatus.PROCESSING:
        raise HTTPException(
            status_code=400,
            detail=f"当前状态为{complaint.status.value}，无法发送回复"
        )
    
    complaint.response_content = response_content
    complaint.status = ComplaintStatus.PENDING_CONFIRMATION
    complaint.pending_confirmation_at = datetime.utcnow()
    
    db.commit()
    db.refresh(complaint)
    return complaint


@router.post("/{complaint_id}/confirm-close", response_model=ComplaintResponse, summary="旅客确认关闭")
def confirm_close(
    complaint_id: int,
    db: Session = Depends(get_db)
):
    complaint = db.query(Complaint).filter(Complaint.id == complaint_id).first()
    if not complaint:
        raise HTTPException(status_code=404, detail="投诉记录不存在")
    
    if complaint.status != ComplaintStatus.PENDING_CONFIRMATION:
        raise HTTPException(
            status_code=400,
            detail=f"当前状态为{complaint.status.value}，无法确认关闭"
        )
    
    complaint.status = ComplaintStatus.CLOSED
    complaint.closed_at = datetime.utcnow()
    
    db.commit()
    db.refresh(complaint)
    return complaint


@router.post("/{complaint_id}/reopen", response_model=ComplaintResponse, summary="旅客不满意重新处理")
def reopen_complaint(
    complaint_id: int,
    db: Session = Depends(get_db)
):
    complaint = db.query(Complaint).filter(Complaint.id == complaint_id).first()
    if not complaint:
        raise HTTPException(status_code=404, detail="投诉记录不存在")
    
    if complaint.status != ComplaintStatus.PENDING_CONFIRMATION:
        raise HTTPException(
            status_code=400,
            detail=f"当前状态为{complaint.status.value}，无法重新处理"
        )
    
    complaint.status = ComplaintStatus.PROCESSING
    complaint.pending_confirmation_at = None
    complaint.response_content = None
    
    db.commit()
    db.refresh(complaint)
    return complaint


@router.put("/{complaint_id}", response_model=ComplaintResponse, summary="更新投诉信息")
def update_complaint(
    complaint_id: int,
    complaint: ComplaintUpdate,
    db: Session = Depends(get_db)
):
    db_complaint = db.query(Complaint).filter(Complaint.id == complaint_id).first()
    if not db_complaint:
        raise HTTPException(status_code=404, detail="投诉记录不存在")
    
    update_data = complaint.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_complaint, key, value)
    
    db.commit()
    db.refresh(db_complaint)
    return db_complaint


def check_overdue_complaints(db: Session):
    now = datetime.utcnow()
    
    processing = db.query(Complaint).filter(
        and_(
            Complaint.status == ComplaintStatus.PROCESSING,
            Complaint.is_overdue == 0
        )
    ).all()
    
    overdue_count = 0
    supervisor_notifications = 0
    
    for complaint in processing:
        if complaint.processing_at:
            hours_passed = (now - complaint.processing_at).total_seconds() / 3600
            limit_hours = (
                settings.SERIOUS_COMPLAINT_REPLY_HOURS 
                if complaint.severity == ComplaintSeverity.SERIOUS 
                else settings.NORMAL_COMPLAINT_REPLY_HOURS
            )
            
            if hours_passed > limit_hours:
                complaint.is_overdue = 1
                if complaint.supervisor_notified == 0:
                    complaint.supervisor_notified = 1
                    supervisor_notifications += 1
                overdue_count += 1
    
    db.commit()
    return {"overdue_count": overdue_count, "supervisor_notifications": supervisor_notifications}


def auto_close_pending_confirmation(db: Session):
    now = datetime.utcnow()
    cutoff = now - timedelta(days=settings.COMPLAINT_CONFIRMATION_DAYS)
    
    to_close = db.query(Complaint).filter(
        and_(
            Complaint.status == ComplaintStatus.PENDING_CONFIRMATION,
            Complaint.pending_confirmation_at < cutoff
        )
    ).all()
    
    for complaint in to_close:
        complaint.status = ComplaintStatus.CLOSED
        complaint.closed_at = now
    
    db.commit()
    return len(to_close)


def get_monthly_statistics(db: Session):
    now = datetime.utcnow()
    start_of_month = datetime(now.year, now.month, 1)
    
    monthly_count = db.query(func.count(Complaint.id)).filter(
        Complaint.registered_at >= start_of_month
    ).scalar() or 0
    
    closed_complaints = db.query(Complaint).filter(
        and_(
            Complaint.status == ComplaintStatus.CLOSED,
            Complaint.closed_at >= start_of_month
        )
    ).all()
    
    total_hours = 0.0
    count = 0
    for c in closed_complaints:
        if c.registered_at and c.closed_at:
            hours = (c.closed_at - c.registered_at).total_seconds() / 3600
            total_hours += hours
            count += 1
    
    avg_hours = total_hours / count if count > 0 else 0.0
    
    overdue_count = db.query(func.count(Complaint.id)).filter(
        and_(
            Complaint.registered_at >= start_of_month,
            Complaint.is_overdue == 1
        )
    ).scalar() or 0
    
    return {
        "monthly_complaint_count": monthly_count,
        "avg_processing_hours": avg_hours,
        "overdue_count": overdue_count
    }
