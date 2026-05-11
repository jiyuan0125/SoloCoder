from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from src.core.database import get_db
from src.core.services import AlarmService
from src.server.schemas import (
    AlarmCreate,
    AlarmUpdateStatus,
    AlarmResponse
)


router = APIRouter(prefix="/alarms", tags=["alarms"])


@router.post("/", response_model=AlarmResponse)
def create_alarm(alarm: AlarmCreate, db: Session = Depends(get_db)):
    service = AlarmService(db)
    return service.create_alarm(alarm.model_dump())


@router.get("/", response_model=List[AlarmResponse])
def list_alarms(
    status: Optional[str] = None,
    severity: Optional[str] = None,
    db: Session = Depends(get_db)
):
    service = AlarmService(db)
    return service.list_alarms(status=status, severity=severity)


@router.get("/{alarm_id}", response_model=AlarmResponse)
def get_alarm(alarm_id: int, db: Session = Depends(get_db)):
    service = AlarmService(db)
    alarm = service.get_alarm(alarm_id)
    if not alarm:
        raise HTTPException(status_code=404, detail="Alarm not found")
    return alarm


@router.put("/{alarm_id}/status", response_model=AlarmResponse)
def update_alarm_status(
    alarm_id: int,
    data: AlarmUpdateStatus,
    db: Session = Depends(get_db)
):
    service = AlarmService(db)
    try:
        alarm = service.update_alarm_status(
            alarm_id,
            data.status,
            handled_by=data.handled_by,
            false_alarm_reason=data.false_alarm_reason
        )
        if not alarm:
            raise HTTPException(status_code=404, detail="Alarm not found")
        return alarm
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/escalate")
def trigger_escalation(db: Session = Depends(get_db)):
    service = AlarmService(db)
    count = service.escalate_alarms()
    return {"message": f"Escalated {count} alarms"}


@router.post("/reminders")
def trigger_reminders(db: Session = Depends(get_db)):
    service = AlarmService(db)
    count = service.send_reminders()
    return {"message": f"Sent reminders for {count} alarms"}
