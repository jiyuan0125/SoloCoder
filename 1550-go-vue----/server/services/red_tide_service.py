from datetime import datetime, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session
from ..models import (
    ChlorophyllReading, ChlorophyllBaseline, RedTideEvent, RedTideReading,
    Station, RedTideStatus, DissolvedOxygenReading
)
from ..schemas import RedTideEventCreate, RedTideEventConfirm
from .notification_service import NotificationService


MIN_BASELINE_SAMPLES = 10
ANOMALY_THRESHOLD_MULTIPLIER = 3


class RedTideService:
    @staticmethod
    def calculate_baseline(values: List[float]) -> Optional[dict]:
        if len(values) < MIN_BASELINE_SAMPLES:
            return None
        
        n = len(values)
        mean = sum(values) / n
        variance = sum((x - mean) ** 2 for x in values) / n
        std_dev = variance ** 0.5
        
        return {
            "mean": mean,
            "std_dev": std_dev,
            "threshold": mean + (ANOMALY_THRESHOLD_MULTIPLIER * std_dev),
            "sample_count": n
        }

    @staticmethod
    def update_baseline(db: Session, station_id: int):
        readings = db.query(ChlorophyllReading).filter(
            ChlorophyllReading.station_id == station_id,
            ChlorophyllReading.is_anomaly == False
        ).order_by(ChlorophyllReading.reading_time.desc()).limit(100).all()
        
        values = [r.value for r in readings]
        baseline = RedTideService.calculate_baseline(values)
        
        if baseline is None:
            return None
        
        existing = db.query(ChlorophyllBaseline).filter(
            ChlorophyllBaseline.station_id == station_id
        ).first()
        
        if existing:
            existing.mean = baseline["mean"]
            existing.std_dev = baseline["std_dev"]
            existing.threshold = baseline["threshold"]
            existing.sample_count = baseline["sample_count"]
        else:
            baseline_record = ChlorophyllBaseline(
                station_id=station_id,
                mean=baseline["mean"],
                std_dev=baseline["std_dev"],
                threshold=baseline["threshold"],
                sample_count=baseline["sample_count"]
            )
            db.add(baseline_record)
        
        db.commit()
        return baseline

    @staticmethod
    def get_baseline(db: Session, station_id: int) -> Optional[ChlorophyllBaseline]:
        return db.query(ChlorophyllBaseline).filter(
            ChlorophyllBaseline.station_id == station_id
        ).first()

    @staticmethod
    def create_chlorophyll_reading(
        db: Session, 
        station_id: int, 
        value: float, 
        reading_time: datetime
    ):
        baseline = RedTideService.get_baseline(db, station_id)
        is_anomaly = False
        
        if baseline and value > baseline.threshold:
            is_anomaly = True
        
        reading = ChlorophyllReading(
            station_id=station_id,
            value=value,
            reading_time=reading_time,
            is_anomaly=is_anomaly
        )
        db.add(reading)
        db.commit()
        db.refresh(reading)
        
        if is_anomaly:
            RedTideService.check_suspected_red_tide(db, station_id, reading)
        else:
            RedTideService.check_resolve_red_tide(db, station_id, reading)
        
        return reading

    @staticmethod
    def check_suspected_red_tide(
        db: Session, 
        station_id: int, 
        current_reading: ChlorophyllReading
    ):
        active_event = db.query(RedTideEvent).filter(
            RedTideEvent.station_id == station_id,
            RedTideEvent.status != RedTideStatus.RESOLVED
        ).first()
        
        if active_event:
            return
        
        recent_anomalies = db.query(ChlorophyllReading).filter(
            ChlorophyllReading.station_id == station_id,
            ChlorophyllReading.is_anomaly == True
        ).order_by(ChlorophyllReading.reading_time.desc()).limit(3).all()
        
        if len(recent_anomalies) >= 3:
            reading_times = [r.reading_time for r in recent_anomalies]
            if len(reading_times) >= 3:
                baseline = RedTideService.get_baseline(db, station_id)
                event = RedTideEvent(
                    station_id=station_id,
                    status=RedTideStatus.SUSPECTED,
                    notes=f"连续3次叶绿素浓度异常检测。阈值: {baseline.threshold:.2f} μg/L" if baseline else "连续3次叶绿素浓度异常检测"
                )
                db.add(event)
                db.commit()
                db.refresh(event)
                
                for reading in recent_anomalies:
                    rt_reading = RedTideReading(
                        event_id=event.id,
                        chlorophyll_value=reading.value,
                        reading_time=reading.reading_time,
                        is_over_threshold=True
                    )
                    db.add(rt_reading)
                db.commit()

    @staticmethod
    def check_resolve_red_tide(
        db: Session, 
        station_id: int, 
        current_reading: ChlorophyllReading
    ):
        active_event = db.query(RedTideEvent).filter(
            RedTideEvent.station_id == station_id,
            RedTideEvent.status.notin_([RedTideStatus.RESOLVED, RedTideStatus.SUSPECTED, RedTideStatus.CONFIRMED])
        ).first()
        
        if not active_event:
            return
        
        baseline = RedTideService.get_baseline(db, station_id)
        if not baseline:
            return
        
        recent_readings = db.query(ChlorophyllReading).filter(
            ChlorophyllReading.station_id == station_id
        ).order_by(ChlorophyllReading.reading_time.desc()).limit(3).all()
        
        all_normal = all(r.value <= baseline.threshold for r in recent_readings)
        
        if all_normal:
            active_event.status = RedTideStatus.RESOLVED
            active_event.resolved_at = datetime.utcnow()
            db.commit()

    @staticmethod
    def confirm_red_tide(
        db: Session, 
        event_id: int, 
        confirm_data: RedTideEventConfirm
    ):
        event = db.query(RedTideEvent).filter(RedTideEvent.id == event_id).first()
        if not event:
            return None
        
        event.status = RedTideStatus.CONFIRMED
        event.confirmed_at = datetime.utcnow()
        event.confirmed_by = confirm_data.confirmed_by
        if confirm_data.notes:
            event.notes = confirm_data.notes
        db.commit()
        db.refresh(event)
        return event

    @staticmethod
    def publish_red_tide(db: Session, event_id: int):
        event = db.query(RedTideEvent).filter(RedTideEvent.id == event_id).first()
        if not event or event.status != RedTideStatus.CONFIRMED:
            return None
        
        event.status = RedTideStatus.PUBLISHED
        event.published_at = datetime.utcnow()
        db.commit()
        db.refresh(event)
        
        NotificationService.notify_farmers(db, event)
        
        event.status = RedTideStatus.MONITORING
        db.commit()
        db.refresh(event)
        
        return event

    @staticmethod
    def get_active_event(db: Session, station_id: int) -> Optional[RedTideEvent]:
        return db.query(RedTideEvent).filter(
            RedTideEvent.station_id == station_id,
            RedTideEvent.status != RedTideStatus.RESOLVED
        ).order_by(RedTideEvent.suspected_at.desc()).first()

    @staticmethod
    def get_events_by_station(db: Session, station_id: int, limit: int = 20):
        return db.query(RedTideEvent).filter(
            RedTideEvent.station_id == station_id
        ).order_by(RedTideEvent.suspected_at.desc()).limit(limit).all()

    @staticmethod
    def get_recent_readings(db: Session, station_id: int, limit: int = 20):
        return db.query(ChlorophyllReading).filter(
            ChlorophyllReading.station_id == station_id
        ).order_by(ChlorophyllReading.reading_time.desc()).limit(limit).all()

    @staticmethod
    def get_red_tide_response(db: Session, station_id: int):
        active_event = RedTideService.get_active_event(db, station_id)
        recent_events = RedTideService.get_events_by_station(db, station_id, 5)
        recent_readings = RedTideService.get_recent_readings(db, station_id, 10)
        
        current_status = RedTideStatus.NORMAL
        if active_event:
            current_status = active_event.status
        
        return {
            "current_status": current_status,
            "current_event": active_event,
            "recent_events": recent_events,
            "recent_readings": recent_readings
        }
