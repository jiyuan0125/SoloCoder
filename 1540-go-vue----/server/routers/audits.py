from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session
from sqlalchemy import select, and_

from ..database import get_db
from ..models import AuditLog
from ..schemas import AuditLog as AuditLogSchema

router = APIRouter(prefix="/audit-logs", tags=["审计日志"])


@router.get("", response_model=List[AuditLogSchema])
def list_audit_logs(
    page: int = Query(1, ge=1),
    size: int = Query(10, ge=1, le=100),
    operation_type: Optional[str] = None,
    operator: Optional[str] = None,
    target_type: Optional[str] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    db: Session = Depends(get_db),
):
    stmt = select(AuditLog)
    conditions = []
    
    if operation_type:
        conditions.append(AuditLog.operation_type == operation_type)
    if operator:
        conditions.append(AuditLog.operator.contains(operator))
    if target_type:
        conditions.append(AuditLog.target_type == target_type)
    if start_time:
        conditions.append(AuditLog.operation_time >= start_time)
    if end_time:
        conditions.append(AuditLog.operation_time <= end_time)
    
    if conditions:
        stmt = stmt.where(and_(*conditions))
    
    stmt = stmt.offset((page - 1) * size).limit(size).order_by(AuditLog.operation_time.desc())
    return db.execute(stmt).scalars().all()


@router.get("/types")
def get_operation_types(db: Session = Depends(get_db)):
    stmt = select(AuditLog.operation_type).distinct()
    types = [r[0] for r in db.execute(stmt).all()]
    return {"operation_types": types}


@router.get("/{log_id}", response_model=AuditLogSchema)
def get_audit_log(log_id: int, db: Session = Depends(get_db)):
    from fastapi import HTTPException
    
    stmt = select(AuditLog).where(AuditLog.id == log_id)
    log = db.execute(stmt).scalar_one_or_none()
    if not log:
        raise HTTPException(status_code=404, detail="日志不存在")
    return log
