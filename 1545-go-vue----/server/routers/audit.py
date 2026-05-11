from fastapi import APIRouter, Depends
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload

from ..auth import get_current_admin
from ..database import get_db
from ..models import User, AuditLog
from ..schemas import AuditLogResponse

router = APIRouter(prefix="/api/audit", tags=["审计日志"])


@router.get("", response_model=list[AuditLogResponse])
async def list_audit_logs(
    user_id: int | None = None,
    target_type: str | None = None,
    action: str | None = None,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_admin)
):
    query = select(AuditLog).options(selectinload(AuditLog.operator))
    if user_id:
        query = query.where(AuditLog.user_id == user_id)
    if target_type:
        query = query.where(AuditLog.target_type == target_type)
    if action:
        query = query.where(AuditLog.action == action)
    query = query.order_by(AuditLog.created_at.desc())

    result = await db.execute(query)
    logs = result.scalars().all()
    return [AuditLogResponse.model_validate(log) for log in logs]


@router.get("/{log_id}", response_model=AuditLogResponse)
async def get_audit_log(
    log_id: int,
    db: AsyncSession = Depends(get_db),
    current_user: User = Depends(get_current_admin)
):
    from sqlalchemy.orm import selectinload
    result = await db.execute(
        select(AuditLog)
        .options(selectinload(AuditLog.operator))
        .where(AuditLog.id == log_id)
    )
    log = result.scalar_one_or_none()
    if not log:
        return {"error": "日志不存在"}
    return AuditLogResponse.model_validate(log)
