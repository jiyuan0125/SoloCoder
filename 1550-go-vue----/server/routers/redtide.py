from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Station, StationType, RedTideStatus, RedTideEvent
from ..schemas import (
    RedTideEvent as RedTideSchema,
    RedTideEventConfirm,
    RedTideReading as RedTideReadingSchema
)
from ..services.red_tide_service import RedTideService
from ..services.notification_service import NotificationService

router = APIRouter(prefix="/redtide", tags=["redtide"])


@router.get("/events", response_model=List[RedTideSchema])
def list_events(
    skip: int = 0,
    limit: int = 100,
    status: Optional[RedTideStatus] = None,
    station_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(RedTideEvent)
    if status:
        query = query.filter(RedTideEvent.status == status)
    if station_id:
        query = query.filter(RedTideEvent.station_id == station_id)
    return query.order_by(RedTideEvent.suspected_at.desc()).offset(skip).limit(limit).all()


@router.get("/events/{event_id}", response_model=RedTideSchema)
def get_event(event_id: int, db: Session = Depends(get_db)):
    event = db.query(RedTideEvent).filter(RedTideEvent.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="赤潮事件不存在")
    return event


@router.post("/events/{event_id}/confirm", response_model=RedTideSchema)
def confirm_event(
    event_id: int,
    confirm_data: RedTideEventConfirm,
    db: Session = Depends(get_db)
):
    event = RedTideService.confirm_red_tide(db, event_id, confirm_data)
    if not event:
        raise HTTPException(status_code=404, detail="赤潮事件不存在或无法确认")
    return event


@router.post("/events/{event_id}/publish", response_model=RedTideSchema)
def publish_event(event_id: int, db: Session = Depends(get_db)):
    event = RedTideService.publish_red_tide(db, event_id)
    if not event:
        raise HTTPException(status_code=404, detail="赤潮事件不存在或无法发布")
    return event


@router.post("/events/{event_id}/resolve", response_model=RedTideSchema)
def resolve_event(event_id: int, db: Session = Depends(get_db)):
    event = db.query(RedTideEvent).filter(RedTideEvent.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="赤潮事件不存在")
    
    if event.status == RedTideStatus.RESOLVED:
        return event
    
    event.status = RedTideStatus.RESOLVED
    event.resolved_at = datetime.utcnow()
    db.commit()
    db.refresh(event)
    return event


@router.get("/events/{event_id}/readings", response_model=List[RedTideReadingSchema])
def get_event_readings(
    event_id: int,
    limit: int = Query(50, ge=1, le=500),
    db: Session = Depends(get_db)
):
    from ..models import RedTideReading
    return db.query(RedTideReading).filter(
        RedTideReading.event_id == event_id
    ).order_by(RedTideReading.reading_time.desc()).limit(limit).all()


@router.post("/baseline/{station_id}/update")
def update_baseline(station_id: int, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    if station.station_type != StationType.BUOY:
        raise HTTPException(status_code=400, detail="岸基站不支持叶绿素基线")
    
    baseline = RedTideService.update_baseline(db, station_id)
    if not baseline:
        return {
            "status": "insufficient_data",
            "message": "数据不足，至少需要10个正常数据样本"
        }
    return {
        "status": "updated",
        "baseline": baseline
    }


@router.get("/baseline/{station_id}")
def get_baseline(station_id: int, db: Session = Depends(get_db)):
    baseline = RedTideService.get_baseline(db, station_id)
    if not baseline:
        return {
            "status": "not_set",
            "message": "尚未建立基线，请先录入足够数据"
        }
    return {
        "status": "active",
        "mean": baseline.mean,
        "std_dev": baseline.std_dev,
        "threshold": baseline.threshold,
        "sample_count": baseline.sample_count,
        "last_updated": baseline.last_updated
    }
