from sqlalchemy.orm import Session

from .models import AuditLog


def log_operation(
    db: Session,
    operation_type: str,
    operator: str,
    target_type: str = None,
    target_id: int = None,
    description: str = None,
) -> AuditLog:
    audit = AuditLog(
        operation_type=operation_type,
        operator=operator,
        target_type=target_type,
        target_id=target_id,
        description=description,
    )
    db.add(audit)
    db.commit()
    db.refresh(audit)
    return audit
