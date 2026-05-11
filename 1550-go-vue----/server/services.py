from datetime import datetime, timedelta
from typing import List, Optional, Tuple
from sqlalchemy.orm import Session
from sqlalchemy import func, and_
import math
from .models import (
    Station, StationType, ChlorophyllMeasurement, SalinityMeasurement,
    RedtideEvent, RedtideStatus, WaveAlertLevel, Farmer, Notification
)
from .config import settings


def calculate_wave_alert_level(wave_height: float) -> WaveAlertLevel:
    if wave_height >= 9.0:
        return WaveAlertLevel.RED
    elif wave_height >= 6.0:
        return WaveAlertLevel.ORANGE
    elif wave_height >= 4.0:
        return WaveAlertLevel.YELLOW
    elif wave_height >= 2.5:
        return WaveAlertLevel.BLUE
    return WaveAlertLevel.NONE


def calculate_normal_range(values: List[float], std_multiple: float = 3.0) -> Tuple[float, float]:
    if len(values) < 2:
        return 0.0, float('inf')
    
    mean = sum(values) / len(values)
    variance = sum((x - mean) ** 2 for x in values) / (len(values) - 1)
    std_dev = math.sqrt(variance)
    
    lower = mean - std_multiple * std_dev
    upper = mean + std_multiple * std_dev
    
    return max(lower, 0.0), upper


def is_salinity_sensor_ok(db: Session, station_id: str, check_time: Optional[datetime] = None) -> bool:
    if check_time is None:
        check_time = datetime.utcnow()
    
    threshold_time = check_time - timedelta(hours=settings.salt_missing_hours)
    
    latest = db.query(SalinityMeasurement).filter(
        SalinityMeasurement.station_id == station_id,
        SalinityMeasurement.measured_at >= threshold_time
    ).order_by(SalinityMeasurement.measured_at.desc()).first()
    
    return latest is not None


def check_redtide_suspected(db: Session, station_id: str, measured_at: datetime) -> bool:
    threshold_time = measured_at - timedelta(days=30)
    
    historical = db.query(ChlorophyllMeasurement).filter(
        ChlorophyllMeasurement.station_id == station_id,
        ChlorophyllMeasurement.measured_at >= threshold_time,
        ChlorophyllMeasurement.measured_at < measured_at
    ).order_by(ChlorophyllMeasurement.measured_at.desc()).limit(settings.redtide_normal_samples).all()
    
    if len(historical) < 10:
        return False
    
    historical_values = [m.chlorophyll for m in historical]
    _, upper_bound = calculate_normal_range(historical_values, settings.redtide_std_multiple)
    
    recent = db.query(ChlorophyllMeasurement).filter(
        ChlorophyllMeasurement.station_id == station_id,
        ChlorophyllMeasurement.measured_at <= measured_at
    ).order_by(ChlorophyllMeasurement.measured_at.desc()).limit(settings.redtide_consecutive_samples).all()
    
    if len(recent) < settings.redtide_consecutive_samples:
        return False
    
    return all(m.chlorophyll > upper_bound for m in recent)


def check_redtide_resolved(db: Session, station_id: str, measured_at: datetime) -> bool:
    active_event = db.query(RedtideEvent).filter(
        RedtideEvent.station_id == station_id,
        RedtideEvent.status.notin_([RedtideStatus.RESOLVED])
    ).first()
    
    if not active_event:
        return False
    
    threshold_time = measured_at - timedelta(days=30)
    
    historical = db.query(ChlorophyllMeasurement).filter(
        ChlorophyllMeasurement.station_id == station_id,
        ChlorophyllMeasurement.measured_at >= threshold_time,
        ChlorophyllMeasurement.measured_at < measured_at
    ).order_by(ChlorophyllMeasurement.measured_at.desc()).limit(settings.redtide_normal_samples).all()
    
    if len(historical) < 10:
        return False
    
    historical_values = [m.chlorophyll for m in historical]
    _, upper_bound = calculate_normal_range(historical_values, settings.redtide_std_multiple)
    
    recent = db.query(ChlorophyllMeasurement).filter(
        ChlorophyllMeasurement.station_id == station_id,
        ChlorophyllMeasurement.measured_at <= measured_at
    ).order_by(ChlorophyllMeasurement.measured_at.desc()).limit(settings.redtide_consecutive_samples).all()
    
    if len(recent) < settings.redtide_consecutive_samples:
        return False
    
    return all(m.chlorophyll <= upper_bound for m in recent)


def create_redtide_event(db: Session, station_id: str, description: Optional[str] = None) -> RedtideEvent:
    event = RedtideEvent(
        station_id=station_id,
        status=RedtideStatus.SUSPECTED,
        description=description or "疑似赤潮事件"
    )
    db.add(event)
    db.commit()
    db.refresh(event)
    return event


def update_redtide_event_status(db: Session, event_id: int, new_status: RedtideStatus) -> Optional[RedtideEvent]:
    event = db.query(RedtideEvent).filter(RedtideEvent.id == event_id).first()
    if not event:
        return None
    
    now = datetime.utcnow()
    
    if new_status == RedtideStatus.CONFIRMED and event.status == RedtideStatus.SUSPECTED:
        event.confirmed_at = now
        send_farmer_notifications(db, event)
    
    if new_status == RedtideStatus.PUBLISHED and event.status in [RedtideStatus.SUSPECTED, RedtideStatus.CONFIRMED]:
        event.published_at = now
    
    if new_status == RedtideStatus.RESOLVED and event.status != RedtideStatus.RESOLVED:
        event.resolved_at = now
    
    event.status = new_status
    db.commit()
    db.refresh(event)
    return event


def send_farmer_notifications(db: Session, event: RedtideEvent) -> None:
    station = db.query(Station).filter(Station.id == event.station_id).first()
    if not station or not station.sea_area:
        return
    
    farmers = db.query(Farmer).filter(
        Farmer.sea_area == station.sea_area,
        Farmer.is_active == True
    ).all()
    
    message = (
        f"【赤潮预警通知】\n"
        f"监测站：{station.name} (ID: {station.id})\n"
        f"海域：{station.sea_area}\n"
        f"赤潮事件已确认，请密切关注海域情况，及时采取防护措施。"
    )
    
    for farmer in farmers:
        notification = Notification(
            redtide_event_id=event.id,
            farmer_id=farmer.id,
            message=message
        )
        db.add(notification)
    
    db.commit()
