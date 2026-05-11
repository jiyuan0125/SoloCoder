from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from server.database import get_db
from core.models import AgingAlertResponse
from core import services

router = APIRouter(prefix="/alerts", tags=["alerts"])


@router.get("", response_model=List[AgingAlertResponse])
def list_all(
    facility_id: Optional[int] = None,
    resolved: Optional[bool] = None,
    db: Session = Depends(get_db)
):
    return services.list_aging_alerts(db, facility_id, resolved)


@router.post("/{alert_id}/resolve", response_model=AgingAlertResponse)
def resolve(alert_id: int, db: Session = Depends(get_db)):
    alert = services.resolve_aging_alert(db, alert_id)
    if not alert:
        raise HTTPException(status_code=404, detail="Alert not found")
    return alert
