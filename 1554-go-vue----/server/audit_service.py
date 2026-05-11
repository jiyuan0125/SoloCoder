from datetime import datetime
from sqlalchemy.orm import Session
from . import models, schemas


def log_action(db: Session, action: str, model_name: str, record_id: int = None, detail: str = None):
    audit_log = models.AuditLog(
        action=action,
        model_name=model_name,
        record_id=record_id,
        detail=detail,
        created_at=datetime.utcnow()
    )
    db.add(audit_log)
    db.commit()
    return audit_log


def get_audit_logs(db: Session, skip: int = 0, limit: int = 100):
    return db.query(models.AuditLog).order_by(models.AuditLog.id.desc()).offset(skip).limit(limit).all()


def get_audit_logs_by_model(db: Session, model_name: str, skip: int = 0, limit: int = 100):
    return db.query(models.AuditLog).filter(
        models.AuditLog.model_name == model_name
    ).order_by(models.AuditLog.id.desc()).offset(skip).limit(limit).all()


def get_audit_logs_by_record(db: Session, model_name: str, record_id: int, skip: int = 0, limit: int = 100):
    return db.query(models.AuditLog).filter(
        models.AuditLog.model_name == model_name,
        models.AuditLog.record_id == record_id
    ).order_by(models.AuditLog.id.desc()).offset(skip).limit(limit).all()
