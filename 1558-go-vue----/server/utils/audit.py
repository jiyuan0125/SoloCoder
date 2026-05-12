import json
from typing import Optional

from sqlalchemy.orm import Session

from server.models.models import AuditLog


def create_audit_log(
    db: Session,
    action: str,
    actor: str,
    details: Optional[dict] = None,
    declaration_id: Optional[int] = None,
) -> AuditLog:
    log = AuditLog(
        declaration_id=declaration_id,
        action=action,
        actor=actor,
        details=json.dumps(details) if details else None,
    )
    db.add(log)
    db.flush()
    return log
