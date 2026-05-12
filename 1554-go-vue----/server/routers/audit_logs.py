from typing import List, Optional
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import AuditLogResponse
from server import services

router = APIRouter(prefix="/audit-logs", tags=["audit logs"])


@router.get("", response_model=List[AuditLogResponse])
def list_audit_logs(
    skip: int = 0,
    limit: int = 100,
    entity_type: Optional[str] = None,
    db: Session = Depends(get_db)
):
    return services.get_audit_logs(db, skip=skip, limit=limit, entity_type=entity_type)
