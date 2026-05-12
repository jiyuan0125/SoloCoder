from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.core.database import get_db
from server.models import Alarm, EnergyRecord, Lighthouse, LightRecord, SpareUsage, WorkOrder
from server.schemas import (
    AlarmResponse,
    EnergyRecordCreate,
    EnergyRecordResponse,
    LightRecordCreate,
    LightRecordResponse,
)
from server.services import MonitoringService

router = APIRouter(tags=["监控告警"])


@router.post("/monitoring/light", response_model=LightRecordResponse)
def add_light_record(data: LightRecordCreate, db: Session = Depends(get_db)):
    try:
        service = MonitoringService(db)
        return service.add_light_record(data)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.post("/monitoring/energy", response_model=EnergyRecordResponse)
def add_energy_record(data: EnergyRecordCreate, db: Session = Depends(get_db)):
    try:
        service = MonitoringService(db)
        return service.add_energy_record(data)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/monitoring/light/{lighthouse_id}", response_model=List[LightRecordResponse])
def get_light_records(lighthouse_id: int, limit: int = 100, db: Session = Depends(get_db)):
    service = MonitoringService(db)
    return service.get_light_records(lighthouse_id, limit)


@router.get("/monitoring/energy/{lighthouse_id}", response_model=List[EnergyRecordResponse])
def get_energy_records(lighthouse_id: int, limit: int = 100, db: Session = Depends(get_db)):
    service = MonitoringService(db)
    return service.get_energy_records(lighthouse_id, limit)


@router.get("/alarms", response_model=List[AlarmResponse])
def list_alarms(status: Optional[str] = None, db: Session = Depends(get_db)):
    if status:
        return db.query(Alarm).filter(Alarm.status == status).order_by(Alarm.created_at.desc()).all()
    return db.query(Alarm).order_by(Alarm.created_at.desc()).all()


@router.post("/alarms/{alarm_id}/resolve", response_model=AlarmResponse)
def resolve_alarm(alarm_id: int, resolved_by: str, db: Session = Depends(get_db)):
    service = MonitoringService(db)
    alarm = service.resolve_alarm(alarm_id, resolved_by)
    if not alarm:
        raise HTTPException(status_code=404, detail="告警不存在")
    return alarm


@router.get("/lighthouses/{lighthouse_id}/dashboard")
def get_lighthouse_dashboard(lighthouse_id: int, db: Session = Depends(get_db)):
    lighthouse = db.query(Lighthouse).filter(Lighthouse.id == lighthouse_id).first()
    if not lighthouse:
        raise HTTPException(status_code=404, detail="灯塔不存在")

    light_records = (
        db.query(LightRecord)
        .filter(LightRecord.lighthouse_id == lighthouse_id)
        .order_by(LightRecord.timestamp.desc())
        .limit(20)
        .all()
    )

    energy_records = (
        db.query(EnergyRecord)
        .filter(EnergyRecord.lighthouse_id == lighthouse_id)
        .order_by(EnergyRecord.timestamp.desc())
        .limit(20)
        .all()
    )

    work_orders = (
        db.query(WorkOrder)
        .filter(WorkOrder.lighthouse_id == lighthouse_id)
        .order_by(WorkOrder.created_at.desc())
        .limit(20)
        .all()
    )

    alarms = (
        db.query(Alarm)
        .filter(Alarm.lighthouse_id == lighthouse_id)
        .order_by(Alarm.created_at.desc())
        .limit(20)
        .all()
    )

    spare_usages = (
        db.query(SpareUsage)
        .filter(SpareUsage.lighthouse_id == lighthouse_id)
        .order_by(SpareUsage.created_at.desc())
        .limit(20)
        .all()
    )

    return {
        "lighthouse": lighthouse,
        "light_records": light_records,
        "energy_records": energy_records,
        "work_orders": work_orders,
        "alarms": alarms,
        "spare_usages": spare_usages,
    }
