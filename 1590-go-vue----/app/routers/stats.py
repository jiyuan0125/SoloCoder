from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session
from sqlalchemy import func
from datetime import datetime, date
from typing import Optional
from app.database import get_db
from app.models import TripRecord, ShiftRecord, MaintenanceTask, MaintenanceStatus
from app.schemas import DailyStatsResponse, ShiftStatsResponse, ShiftStatsItem

router = APIRouter(prefix="/stats", tags=["运营统计"])


@router.get("/daily", response_model=DailyStatsResponse)
def get_daily_stats(
    target_date: Optional[date] = Query(None, description="查询日期，默认为今天"),
    db: Session = Depends(get_db)
):
    if not target_date:
        target_date = date.today()
    
    start_dt = datetime.combine(target_date, datetime.min.time())
    end_dt = datetime.combine(target_date, datetime.max.time())
    
    trips = db.query(TripRecord).filter(
        TripRecord.dispatch_time >= start_dt,
        TripRecord.dispatch_time <= end_dt
    ).all()
    
    total_trips = len(trips)
    total_passengers = sum(t.passenger_count for t in trips if t.passenger_count is not None) or None
    total_weight = sum(t.weight for t in trips) or None
    avg_weight = total_weight / total_trips if total_trips > 0 and total_weight else None
    
    hour_counts = {}
    for trip in trips:
        if trip.passenger_count:
            hour = trip.dispatch_time.hour
            hour_counts[hour] = hour_counts.get(hour, 0) + trip.passenger_count
    
    peak_passengers = max(hour_counts.values()) if hour_counts else None
    
    maintenance_completed = db.query(MaintenanceTask).filter(
        MaintenanceTask.scheduled_date == target_date,
        MaintenanceTask.status == MaintenanceStatus.COMPLETED
    ).count()
    
    maintenance_overdue = db.query(MaintenanceTask).filter(
        MaintenanceTask.scheduled_date <= target_date,
        MaintenanceTask.status == MaintenanceStatus.OVERDUE
    ).count()
    
    return DailyStatsResponse(
        date=target_date,
        total_trips=total_trips,
        total_passengers=total_passengers,
        total_weight=total_weight,
        average_weight_per_trip=round(avg_weight, 1) if avg_weight else None,
        peak_hour_passengers=peak_passengers,
        maintenance_completed=maintenance_completed,
        maintenance_overdue=maintenance_overdue
    )


@router.get("/shifts", response_model=ShiftStatsResponse)
def get_shift_stats(
    target_date: Optional[date] = Query(None, description="查询日期，默认为今天"),
    db: Session = Depends(get_db)
):
    if not target_date:
        target_date = date.today()
    
    start_dt = datetime.combine(target_date, datetime.min.time())
    end_dt = datetime.combine(target_date, datetime.max.time())
    
    shifts = db.query(ShiftRecord).filter(
        ShiftRecord.shift_date == target_date
    ).order_by(ShiftRecord.shift_number).all()
    
    shift_items = []
    for shift in shifts:
        shift_items.append(ShiftStatsItem(
            shift_number=shift.shift_number,
            start_time=shift.start_time,
            end_time=shift.end_time,
            total_passengers=shift.total_passengers,
            total_weight=shift.total_weight,
            total_trips=shift.total_trips,
            total_revenue=shift.total_revenue,
            average_wait_time=shift.average_wait_time
        ))
    
    return ShiftStatsResponse(
        date=target_date,
        shifts=shift_items
    )
