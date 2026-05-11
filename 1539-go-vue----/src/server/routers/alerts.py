from fastapi import APIRouter, HTTPException, Query
from typing import List, Optional
from datetime import datetime

from src.core import storage, check_calibration_status
from src.core.models import AlertEvent, AlertType, AlertLevel
from src.server.schemas import AlertEventResponse, MessageResponse

router = APIRouter(prefix="/alerts", tags=["alerts"])


@router.post("/check-calibration", response_model=List[AlertEventResponse])
def trigger_calibration_check():
    return check_calibration_status()


@router.get("/", response_model=List[AlertEventResponse])
def get_alerts(
    point_id: Optional[int] = Query(None, description="Filter by monitoring point ID"),
    alert_type: Optional[AlertType] = Query(None, description="Filter by alert type"),
    is_resolved: Optional[bool] = Query(None, description="Filter by resolved status"),
):
    results = storage.get_all(AlertEvent)
    
    if point_id is not None:
        results = [a for a in results if a.point_id == point_id]
    
    if alert_type:
        results = [a for a in results if a.alert_type == alert_type]
    
    if is_resolved is not None:
        results = [a for a in results if a.is_resolved == is_resolved]
    
    results.sort(key=lambda x: x.created_at, reverse=True)
    return results


@router.get("/{alert_id}", response_model=AlertEventResponse)
def get_alert(alert_id: int):
    alert = storage.get_by_id(AlertEvent, alert_id)
    if not alert:
        raise HTTPException(status_code=404, detail="Alert not found")
    return alert


@router.post("/{alert_id}/resolve", response_model=AlertEventResponse)
def resolve_alert(alert_id: int):
    alert = storage.get_by_id(AlertEvent, alert_id)
    if not alert:
        raise HTTPException(status_code=404, detail="Alert not found")
    
    if alert.is_resolved:
        return alert
    
    alert.is_resolved = True
    alert.resolved_at = datetime.now()
    
    return storage.update(alert)


@router.delete("/{alert_id}", status_code=204)
def delete_alert(alert_id: int):
    if not storage.delete(AlertEvent, alert_id):
        raise HTTPException(status_code=404, detail="Alert not found")
