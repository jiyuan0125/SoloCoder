from typing import List, Optional

from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from server.config import get_db
from server.models.models import AuditLog
from server.schemas.schemas import AuditLogResponse

router = APIRouter(prefix="/audit", tags=["audit"])


@router.get("", response_model=List[AuditLogResponse])
def list_audit_logs(
    declaration_id: Optional[int] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(AuditLog)
    if declaration_id is not None:
        query = query.filter(AuditLog.declaration_id == declaration_id)
    logs = query.order_by(AuditLog.id.desc()).offset(skip).limit(limit).all()
    return logs
