from typing import List, Optional
from datetime import datetime
from fastapi import APIRouter, HTTPException, status, Query

from src.core.models.schemas import (
    SensorReading, SensorReadingCreate, AlarmRecord, AlarmLevel
)
from src.core.storage.memory_store import get_store
from src.core.services.alarm_service import get_alarm_service


router = APIRouter(prefix="/sensors", tags=["sensors"])
store = get_store()
alarm_service = get_alarm_service()


@router.post("/reading", response_model=SensorReading, status_code=status.HTTP_201_CREATED)
def report_sensor_reading(data: SensorReadingCreate):
    reactor = store.get_reactor(data.reactor_id)
    if not reactor:
        raise HTTPException(status_code=404, detail="Reactor not found")
    
    alarm_level, is_high_risk = alarm_service.determine_alarm_level(
        reactor, data.temperature, data.pressure
    )
    
    reading = SensorReading(
        reactor_id=data.reactor_id,
        temperature=data.temperature,
        pressure=data.pressure,
        alarm_level=alarm_level,
        is_high_risk=is_high_risk
    )
    
    reading = store.add_sensor_reading(reading)
    
    if alarm_level != AlarmLevel.NORMAL:
        alarm_record = alarm_service.create_alarm_record(reading, reactor)
        store.add_alarm_record(alarm_record)
    
    return reading


@router.get("/readings", response_model=List[SensorReading])
def list_readings(
    reactor_id: Optional[str] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None
):
    return store.list_sensor_readings(reactor_id, start_time, end_time)


@router.get("/alarms", response_model=List[AlarmRecord])
def list_alarms(
    reactor_id: Optional[str] = None,
    start_time: Optional[datetime] = None,
    end_time: Optional[datetime] = None,
    is_high_risk: Optional[bool] = None
):
    return store.list_alarm_records(reactor_id, start_time, end_time, is_high_risk)
