from datetime import datetime, timedelta
from typing import List
from sqlalchemy.orm import Session
from ..models import TideReading, Station
from ..schemas import HourlyTideReading, DailyTideStats


class TideService:
    @staticmethod
    def create_reading(db: Session, station_id: int, value: float, reading_time: datetime):
        reading = TideReading(
            station_id=station_id,
            value=value,
            reading_time=reading_time
        )
        db.add(reading)
        db.commit()
        db.refresh(reading)
        return reading

    @staticmethod
    def get_readings_by_station(db: Session, station_id: int, limit: int = 100):
        return db.query(TideReading).filter(
            TideReading.station_id == station_id
        ).order_by(TideReading.reading_time.desc()).limit(limit).all()

    @staticmethod
    def get_readings_by_time_range(
        db: Session, 
        station_id: int, 
        start_time: datetime, 
        end_time: datetime
    ):
        return db.query(TideReading).filter(
            TideReading.station_id == station_id,
            TideReading.reading_time >= start_time,
            TideReading.reading_time <= end_time
        ).order_by(TideReading.reading_time.asc()).all()

    @staticmethod
    def get_hourly_readings(
        db: Session, 
        station_id: int, 
        target_date: datetime
    ) -> List[HourlyTideReading]:
        start_of_day = target_date.replace(hour=0, minute=0, second=0, microsecond=0)
        end_of_day = start_of_day + timedelta(days=1) - timedelta(microseconds=1)
        
        readings = db.query(TideReading).filter(
            TideReading.station_id == station_id,
            TideReading.reading_time >= start_of_day,
            TideReading.reading_time <= end_of_day
        ).order_by(TideReading.reading_time.asc()).all()
        
        hourly_readings = []
        for reading in readings:
            hour = reading.reading_time.hour
            minute = reading.reading_time.minute
            second = reading.reading_time.second
            if minute == 0 and second == 0:
                hourly_readings.append(HourlyTideReading(
                    hour=hour,
                    value=reading.value,
                    reading_time=reading.reading_time
                ))
        return hourly_readings

    @staticmethod
    def get_daily_stats(
        db: Session, 
        station_id: int, 
        start_date: datetime, 
        end_date: datetime
    ) -> List[DailyTideStats]:
        results = []
        current_date = start_date.replace(hour=0, minute=0, second=0, microsecond=0)
        
        while current_date <= end_date:
            day_start = current_date
            day_end = current_date + timedelta(days=1) - timedelta(microseconds=1)
            
            readings = db.query(TideReading).filter(
                TideReading.station_id == station_id,
                TideReading.reading_time >= day_start,
                TideReading.reading_time <= day_end
            ).all()
            
            if readings:
                values = [r.value for r in readings]
                results.append(DailyTideStats(
                    date=current_date.strftime("%Y-%m-%d"),
                    max_value=max(values),
                    min_value=min(values),
                    avg_value=sum(values) / len(values)
                ))
            
            current_date += timedelta(days=1)
        
        return results

    @staticmethod
    def get_tide_response(
        db: Session, 
        station_id: int, 
        target_date: datetime = None
    ):
        if target_date is None:
            target_date = datetime.utcnow()
        
        hourly = TideService.get_hourly_readings(db, station_id, target_date)
        
        start_date = target_date - timedelta(days=6)
        daily = TideService.get_daily_stats(db, station_id, start_date, target_date)
        
        return {
            "hourly_readings": hourly,
            "daily_stats": daily
        }
