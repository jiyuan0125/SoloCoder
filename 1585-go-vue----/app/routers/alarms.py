from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional

from app.database import get_db
from app.schemas import AlarmResponse, AlarmCreate
from app.services.alarm_service import AlarmService

router = APIRouter(prefix="/alarms", tags=["alarms"])


@router.get("/", response_model=List[AlarmResponse])
def list_alarms(
    status: Optional[str] = None,
    subsystem: Optional[str] = None,
    level: Optional[str] = None,
    db: Session = Depends(get_db)
):
    return AlarmService.get_all_alarms(db, status, subsystem, level)


@router.get("/active", response_model=List[AlarmResponse])
def list_active_alarms(db: Session = Depends(get_db)):
    return AlarmService.get_active_alarms(db)


@router.get("/{alarm_id}", response_model=AlarmResponse)
def get_alarm(alarm_id: int, db: Session = Depends(get_db)):
    alarm = AlarmService.get_alarm(db, alarm_id)
    if not alarm:
        raise HTTPException(status_code=404, detail="告警不存在")
    return alarm


@router.post("/", response_model=AlarmResponse)
def create_alarm(alarm: AlarmCreate, db: Session = Depends(get_db)):
    return AlarmService.create_alarm(db, alarm.model_dump())


@router.post("/{alarm_id}/acknowledge", response_model=AlarmResponse)
def acknowledge_alarm(alarm_id: int, db: Session = Depends(get_db)):
    alarm = AlarmService.acknowledge_alarm(db, alarm_id)
    if not alarm:
        raise HTTPException(status_code=404, detail="告警不存在")
    return alarm


@router.post("/{alarm_id}/resolve", response_model=AlarmResponse)
def resolve_alarm(alarm_id: int, db: Session = Depends(get_db)):
    alarm = AlarmService.resolve_alarm(db, alarm_id)
    if not alarm:
        raise HTTPException(status_code=404, detail="告警不存在")
    return alarm
