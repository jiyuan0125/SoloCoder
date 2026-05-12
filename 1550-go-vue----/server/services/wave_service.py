from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from ..models import WaveReading, Station, WaveAlertLevel
from ..schemas import WaveResponse


class WaveService:
    @staticmethod
    def determine_alert_level(significant_wave_height: float) -> WaveAlertLevel:
        if significant_wave_height >= 9.0:
            return WaveAlertLevel.RED
        elif significant_wave_height >= 6.0:
            return WaveAlertLevel.ORANGE
        elif significant_wave_height >= 4.0:
            return WaveAlertLevel.YELLOW
        elif significant_wave_height >= 2.5:
            return WaveAlertLevel.BLUE
        return WaveAlertLevel.NONE

    @staticmethod
    def create_reading(
        db: Session, 
        station_id: int, 
        significant_wave_height: float, 
        reading_time: datetime
    ):
        alert_level = WaveService.determine_alert_level(significant_wave_height)
        reading = WaveReading(
            station_id=station_id,
            significant_wave_height=significant_wave_height,
            reading_time=reading_time,
            alert_level=alert_level
        )
        db.add(reading)
        db.commit()
        db.refresh(reading)
        return reading

    @staticmethod
    def get_readings_by_station(db: Session, station_id: int, limit: int = 100):
        return db.query(WaveReading).filter(
            WaveReading.station_id == station_id
        ).order_by(WaveReading.reading_time.desc()).limit(limit).all()

    @staticmethod
    def get_readings_by_time_range(
        db: Session, 
        station_id: int, 
        start_time: datetime, 
        end_time: datetime
    ):
        return db.query(WaveReading).filter(
            WaveReading.station_id == station_id,
            WaveReading.reading_time >= start_time,
            WaveReading.reading_time <= end_time
        ).order_by(WaveReading.reading_time.asc()).all()

    @staticmethod
    def get_latest_reading(db: Session, station_id: int):
        return db.query(WaveReading).filter(
            WaveReading.station_id == station_id
        ).order_by(WaveReading.reading_time.desc()).first()

    @staticmethod
    def get_wave_response(db: Session, station_id: int, limit: int = 24):
        readings = WaveService.get_readings_by_station(db, station_id, limit)
        latest = WaveService.get_latest_reading(db, station_id)
        
        return {
            "readings": readings,
            "latest_alert": latest.alert_level if latest else None
        }

    @staticmethod
    def get_alerts_by_time_range(
        db: Session, 
        station_id: int, 
        start_time: datetime, 
        end_time: datetime
    ):
        return db.query(WaveReading).filter(
            WaveReading.station_id == station_id,
            WaveReading.reading_time >= start_time,
            WaveReading.reading_time <= end_time,
            WaveReading.alert_level != WaveAlertLevel.NONE
        ).order_by(WaveReading.reading_time.asc()).all()
