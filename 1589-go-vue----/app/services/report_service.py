from datetime import datetime, date, timedelta
from typing import Optional
from sqlalchemy.orm import Session
from app.models import (
    DailyReport, StatusRecord, RopewayStatus, 
    PassengerRecord, WeatherData, Inspection,
    InspectionStatus
)
from app.config import MIN_OPERATING_HOURS


class ReportService:
    def __init__(self, db: Session):
        self.db = db

    def calculate_operating_hours(self, report_date: date) -> tuple:
        start_time = datetime.combine(report_date, datetime.min.time())
        end_time = datetime.combine(report_date, datetime.max.time())
        
        status_records = self.db.query(StatusRecord).filter(
            StatusRecord.timestamp >= start_time,
            StatusRecord.timestamp <= end_time
        ).order_by(StatusRecord.timestamp).all()
        
        if not status_records:
            return 0.0, 0.0
        
        total_hours = 0.0
        effective_hours = 0.0
        
        current_time = start_time
        current_status = RopewayStatus.NORMAL
        
        for record in status_records:
            time_diff = (record.timestamp - current_time).total_seconds() / 3600.0
            
            if time_diff > 0:
                if current_status in [RopewayStatus.NORMAL, RopewayStatus.DECELERATE, RopewayStatus.PAUSE]:
                    total_hours += time_diff
                
                if current_status == RopewayStatus.NORMAL:
                    effective_hours += time_diff
            
            current_time = record.timestamp
            current_status = record.to_status
        
        time_diff = (end_time - current_time).total_seconds() / 3600.0
        if time_diff > 0:
            if current_status in [RopewayStatus.NORMAL, RopewayStatus.DECELERATE, RopewayStatus.PAUSE]:
                total_hours += time_diff
            
            if current_status == RopewayStatus.NORMAL:
                effective_hours += time_diff
        
        return round(total_hours, 2), round(effective_hours, 2)

    def get_passenger_stats(self, report_date: date) -> tuple:
        start_time = datetime.combine(report_date, datetime.min.time())
        end_time = datetime.combine(report_date, datetime.max.time())
        
        records = self.db.query(PassengerRecord).filter(
            PassengerRecord.timestamp >= start_time,
            PassengerRecord.timestamp <= end_time
        ).all()
        
        total_actual = sum(r.adult_count + r.child_count for r in records)
        total_counted = sum(r.total_counted for r in records)
        
        return total_actual, total_counted

    def get_weather_events(self, report_date: date) -> str:
        start_time = datetime.combine(report_date, datetime.min.time())
        end_time = datetime.combine(report_date, datetime.max.time())
        
        weather_records = self.db.query(WeatherData).filter(
            WeatherData.timestamp >= start_time,
            WeatherData.timestamp <= end_time,
            (WeatherData.wind_speed >= 8.0) | (WeatherData.has_lightning == True)
        ).order_by(WeatherData.timestamp).all()
        
        events = []
        for record in weather_records:
            if record.has_lightning:
                events.append(f"{record.timestamp.strftime('%H:%M')}: 雷电")
            elif record.wind_speed >= 20.0:
                events.append(f"{record.timestamp.strftime('%H:%M')}: 强风 {record.wind_speed}m/s")
            elif record.wind_speed >= 14.0:
                events.append(f"{record.timestamp.strftime('%H:%M')}: 大风 {record.wind_speed}m/s")
            elif record.wind_speed >= 8.0:
                events.append(f"{record.timestamp.strftime('%H:%M')}: 阵风 {record.wind_speed}m/s")
        
        return "; ".join(events) if events else None

    def get_urgent_inspections(self, report_date: date) -> str:
        start_time = datetime.combine(report_date, datetime.min.time())
        end_time = datetime.combine(report_date, datetime.max.time())
        
        inspections = self.db.query(Inspection).filter(
            Inspection.scheduled_date <= end_time,
            Inspection.is_urgent == True,
            Inspection.resolved == False
        ).all()
        
        issues = []
        for insp in inspections:
            status = "未处理" if not insp.resolved else "已解决"
            issues.append(f"{insp.inspection_type.value}: {insp.issues_found or insp.notes or '问题'} ({status})")
        
        return "; ".join(issues) if issues else None

    def generate_daily_report(self, report_date: Optional[date] = None) -> DailyReport:
        if report_date is None:
            report_date = date.today()
        
        existing = self.db.query(DailyReport).filter(
            DailyReport.report_date == datetime.combine(report_date, datetime.min.time())
        ).first()
        
        total_operating_hours, effective_operating_hours = self.calculate_operating_hours(report_date)
        total_passengers, total_passengers_counted = self.get_passenger_stats(report_date)
        weather_events = self.get_weather_events(report_date)
        urgent_inspections = self.get_urgent_inspections(report_date)
        below_min_hours = effective_operating_hours < MIN_OPERATING_HOURS
        
        if existing:
            existing.total_operating_hours = total_operating_hours
            existing.effective_operating_hours = effective_operating_hours
            existing.total_passengers = total_passengers
            existing.total_passengers_counted = total_passengers_counted
            existing.below_min_hours = below_min_hours
            existing.weather_events = weather_events
            existing.urgent_inspections = urgent_inspections
            report = existing
        else:
            report = DailyReport(
                report_date=datetime.combine(report_date, datetime.min.time()),
                total_operating_hours=total_operating_hours,
                effective_operating_hours=effective_operating_hours,
                total_passengers=total_passengers,
                total_passengers_counted=total_passengers_counted,
                below_min_hours=below_min_hours,
                weather_events=weather_events,
                urgent_inspections=urgent_inspections
            )
            self.db.add(report)
        
        self.db.commit()
        self.db.refresh(report)
        return report

    def get_reports(self, start_date: date, end_date: date) -> list:
        start_dt = datetime.combine(start_date, datetime.min.time())
        end_dt = datetime.combine(end_date, datetime.max.time())
        
        return self.db.query(DailyReport).filter(
            DailyReport.report_date >= start_dt,
            DailyReport.report_date <= end_dt
        ).order_by(DailyReport.report_date).all()

    def get_report_by_date(self, report_date: date) -> Optional[DailyReport]:
        report_dt = datetime.combine(report_date, datetime.min.time())
        return self.db.query(DailyReport).filter(
            DailyReport.report_date == report_dt
        ).first()
