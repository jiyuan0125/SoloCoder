from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from ..models import (
    WaterTempReading, SalinityReading, DissolvedOxygenReading,
    Station, StationType
)


class DataService:
    @staticmethod
    def create_water_temp_reading(
        db: Session, 
        station_id: int, 
        value: float, 
        reading_time: datetime
    ):
        reading = WaterTempReading(
            station_id=station_id,
            value=value,
            reading_time=reading_time
        )
        db.add(reading)
        db.commit()
        db.refresh(reading)
        return reading

    @staticmethod
    def create_salinity_reading(
        db: Session, 
        station_id: int, 
        value: float, 
        reading_time: datetime
    ):
        reading = SalinityReading(
            station_id=station_id,
            value=value,
            reading_time=reading_time
        )
        db.add(reading)
        db.commit()
        db.refresh(reading)
        return reading

    @staticmethod
    def create_do_reading(
        db: Session, 
        station_id: int, 
        value: float, 
        reading_time: datetime
    ):
        reading = DissolvedOxygenReading(
            station_id=station_id,
            value=value,
            reading_time=reading_time
        )
        db.add(reading)
        db.commit()
        db.refresh(reading)
        return reading

    @staticmethod
    def get_latest_water_temp(db: Session, station_id: int):
        return db.query(WaterTempReading).filter(
            WaterTempReading.station_id == station_id
        ).order_by(WaterTempReading.reading_time.desc()).first()

    @staticmethod
    def get_latest_salinity(db: Session, station_id: int):
        return db.query(SalinityReading).filter(
            SalinityReading.station_id == station_id
        ).order_by(SalinityReading.reading_time.desc()).first()

    @staticmethod
    def get_latest_do(db: Session, station_id: int):
        return db.query(DissolvedOxygenReading).filter(
            DissolvedOxygenReading.station_id == station_id
        ).order_by(DissolvedOxygenReading.reading_time.desc()).first()

    @staticmethod
    def check_salinity_missing(db: Session, station_id: int, hours: int = 24) -> dict:
        latest = DataService.get_latest_salinity(db, station_id)
        if latest is None:
            return {
                "status": "missing",
                "hours_missing": float('inf'),
                "last_reading_time": None,
                "message": "无盐度数据记录"
            }
        
        time_since_last = datetime.utcnow() - latest.reading_time
        hours_since = time_since_last.total_seconds() / 3600
        
        if hours_since >= hours:
            return {
                "status": "missing",
                "hours_missing": hours_since,
                "last_reading_time": latest.reading_time,
                "message": f"盐度数据缺失超过 {hours} 小时"
            }
        
        return {
            "status": "normal",
            "hours_missing": hours_since,
            "last_reading_time": latest.reading_time,
            "message": "盐度数据正常"
        }

    @staticmethod
    def get_station_status(db: Session, station_id: int):
        station = db.query(Station).filter(Station.id == station_id).first()
        if not station:
            return None
        
        salinity_status = None
        if station.station_type == StationType.BUOY:
            salinity_status = DataService.check_salinity_missing(db, station_id)
        
        return {
            "station": station,
            "latest_water_temp": DataService.get_latest_water_temp(db, station_id),
            "salinity_status": salinity_status["status"] if salinity_status else None,
            "latest_salinity": DataService.get_latest_salinity(db, station_id),
            "latest_do": DataService.get_latest_do(db, station_id),
            "latest_chlorophyll": None
        }
