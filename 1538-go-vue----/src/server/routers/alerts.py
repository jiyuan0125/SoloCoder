from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from src.core.database import get_db
from src.core.models import Alert
from src.core.schemas import AlertResponse
from src.core.services import check_storage_overdue

router = APIRouter()


@router.post("/check-overdue", response_model=List[AlertResponse])
def trigger_overdue_check(db: Session = Depends(get_db)):
    return check_storage_overdue(db)


@router.get("", response_model=List[AlertResponse])
def list_alerts(
    unit_id: int = None,
    alert_type: str = None,
    is_resolved: bool = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Alert)
    if unit_id:
        query = query.filter(Alert.unit_id == unit_id)
    if alert_type:
        query = query.filter(Alert.alert_type == alert_type)
    if is_resolved is not None:
        query = query.filter(Alert.is_resolved == is_resolved)
    return query.offset(skip).limit(limit).all()


@router.get("/{alert_id}", response_model=AlertResponse)
def get_alert(alert_id: int, db: Session = Depends(get_db)):
    alert = db.query(Alert).filter(Alert.id == alert_id).first()
    if not alert:
        raise HTTPException(status_code=404, detail="预警不存在")
    return alert


@router.post("/{alert_id}/resolve", response_model=AlertResponse)
def resolve_alert(alert_id: int, db: Session = Depends(get_db)):
    alert = db.query(Alert).filter(Alert.id == alert_id).first()
    if not alert:
        raise HTTPException(status_code=404, detail="预警不存在")
    alert.is_resolved = True
    db.commit()
    db.refresh(alert)
    return alert
