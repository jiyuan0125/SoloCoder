from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from sqlalchemy import func, and_
from typing import List, Dict, Any
from datetime import datetime, timedelta
from app.database import get_db, Station, Zone, SecurityGate, PassengerCount, Alert, HolidayForecast
from app.schemas import PassengerCountResponse, AlertResponse, HourlyStats, DailyStats

router = APIRouter()

YELLOW_ALERT_THRESHOLD = 0.8
RED_ALERT_THRESHOLD = 0.9
HOLIDAY_PEAK_THRESHOLD = 1.5

@router.post("/stations/{station_code}/passenger-flow/update", response_model=Dict[str, Any])
def update_passenger_flow(
    station_code: str, 
    zone_id: int, 
    count: int, 
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    zone = db.query(Zone).filter(Zone.id == zone_id, Zone.station_id == station.id).first()
    if not zone:
        raise HTTPException(status_code=404, detail="区域不存在")
    
    zone.current_count = count
    
    passenger_count = PassengerCount(
        zone_id=zone_id,
        count=count,
        timestamp=datetime.utcnow(),
        count_type="realtime"
    )
    db.add(passenger_count)
    
    check_and_create_crowd_alerts(db, station, zone, count)
    
    db.commit()
    db.refresh(zone)
    
    return {
        "zone_id": zone_id,
        "zone_name": zone.name,
        "current_count": count,
        "max_capacity": zone.max_capacity,
        "occupancy_rate": round(count / zone.max_capacity * 100, 2) if zone.max_capacity > 0 else 0
    }

@router.get("/stations/{station_code}/passenger-flow/realtime", response_model=Dict[str, Any])
def get_realtime_flow(station_code: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    zones = db.query(Zone).filter(Zone.station_id == station.id).all()
    
    total_count = sum(z.current_count for z in zones)
    total_capacity = sum(z.max_capacity for z in zones)
    
    zone_details = []
    for zone in zones:
        occupancy_rate = zone.current_count / zone.max_capacity * 100 if zone.max_capacity > 0 else 0
        zone_details.append({
            "zone_id": zone.id,
            "zone_name": zone.name,
            "zone_type": zone.zone_type,
            "current_count": zone.current_count,
            "max_capacity": zone.max_capacity,
            "occupancy_rate": round(occupancy_rate, 2)
        })
    
    overall_occupancy = total_count / total_capacity * 100 if total_capacity > 0 else 0
    
    return {
        "station_code": station_code,
        "station_name": station.name,
        "total_count": total_count,
        "total_capacity": total_capacity,
        "overall_occupancy_rate": round(overall_occupancy, 2),
        "zones": zone_details,
        "timestamp": datetime.utcnow()
    }

@router.get("/stations/{station_code}/passenger-flow/hourly", response_model=Dict[str, Any])
def get_hourly_stats(station_code: str, date: str = None, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    if date:
        target_date = datetime.strptime(date, "%Y-%m-%d")
    else:
        target_date = datetime.utcnow()
    
    start_of_day = datetime(target_date.year, target_date.month, target_date.day)
    end_of_day = start_of_day + timedelta(days=1)
    
    zone_ids = [z.id for z in db.query(Zone).filter(Zone.station_id == station.id).all()]
    
    if not zone_ids:
        return {"station_code": station_code, "date": target_date.strftime("%Y-%m-%d"), "hourly_stats": []}
    
    hourly = db.query(
        func.strftime('%H', PassengerCount.timestamp).label('hour'),
        func.avg(PassengerCount.count).label('avg_count'),
        func.max(PassengerCount.count).label('max_count'),
        func.min(PassengerCount.count).label('min_count')
    ).filter(
        PassengerCount.zone_id.in_(zone_ids),
        PassengerCount.timestamp >= start_of_day,
        PassengerCount.timestamp < end_of_day
    ).group_by(
        func.strftime('%H', PassengerCount.timestamp)
    ).all()
    
    stats = []
    for row in hourly:
        stats.append({
            "hour": int(row.hour),
            "average_count": int(row.avg_count) if row.avg_count else 0,
            "max_count": int(row.max_count) if row.max_count else 0,
            "min_count": int(row.min_count) if row.min_count else 0
        })
    
    return {
        "station_code": station_code,
        "date": target_date.strftime("%Y-%m-%d"),
        "hourly_stats": stats
    }

@router.get("/stations/{station_code}/passenger-flow/daily", response_model=Dict[str, Any])
def get_daily_stats(station_code: str, start_date: str = None, end_date: str = None, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    if end_date:
        end = datetime.strptime(end_date, "%Y-%m-%d") + timedelta(days=1)
    else:
        end = datetime.utcnow() + timedelta(days=1)
    
    if start_date:
        start = datetime.strptime(start_date, "%Y-%m-%d")
    else:
        start = end - timedelta(days=7)
    
    zone_ids = [z.id for z in db.query(Zone).filter(Zone.station_id == station.id).all()]
    
    if not zone_ids:
        return {"station_code": station_code, "daily_stats": []}
    
    daily = db.query(
        func.strftime('%Y-%m-%d', PassengerCount.timestamp).label('date'),
        func.avg(PassengerCount.count).label('avg_count'),
        func.max(PassengerCount.count).label('max_count'),
        func.sum(PassengerCount.count).label('total_count')
    ).filter(
        PassengerCount.zone_id.in_(zone_ids),
        PassengerCount.timestamp >= start,
        PassengerCount.timestamp < end
    ).group_by(
        func.strftime('%Y-%m-%d', PassengerCount.timestamp)
    ).all()
    
    stats = []
    for row in daily:
        stats.append({
            "date": row.date,
            "average_count": int(row.avg_count) if row.avg_count else 0,
            "max_count": int(row.max_count) if row.max_count else 0,
            "total_count": int(row.total_count) if row.total_count else 0
        })
    
    return {
        "station_code": station_code,
        "start_date": start.strftime("%Y-%m-%d"),
        "end_date": (end - timedelta(days=1)).strftime("%Y-%m-%d"),
        "daily_stats": stats
    }

@router.get("/stations/{station_code}/alerts", response_model=List[AlertResponse])
def get_active_alerts(station_code: str, active_only: bool = True, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    query = db.query(Alert).filter(Alert.station_id == station.id)
    if active_only:
        query = query.filter(Alert.is_active == True)
    
    return query.order_by(Alert.created_at.desc()).all()

@router.post("/stations/{station_code}/holiday-forecast", response_model=Dict[str, Any])
def create_holiday_forecast(
    station_code: str, 
    forecast_date: str, 
    predicted_flow: float, 
    normal_flow: float, 
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    forecast_dt = datetime.strptime(forecast_date, "%Y-%m-%d")
    
    is_peak = predicted_flow >= normal_flow * HOLIDAY_PEAK_THRESHOLD
    
    forecast = HolidayForecast(
        station_id=station.id,
        forecast_date=forecast_dt,
        predicted_flow=predicted_flow,
        normal_flow=normal_flow,
        is_peak_alert=is_peak
    )
    db.add(forecast)
    
    if is_peak:
        alert = Alert(
            station_id=station.id,
            alert_type="holiday_peak",
            severity="warning",
            message=f"节假日客流高峰预警：{forecast_date} 预测客流 {predicted_flow}，超过平常水平的 {HOLIDAY_PEAK_THRESHOLD * 100}%",
            is_active=True
        )
        db.add(alert)
        
        adjust_security_gates_for_peak(db, station)
    
    db.commit()
    db.refresh(forecast)
    
    return {
        "forecast_id": forecast.id,
        "station_code": station_code,
        "forecast_date": forecast_date,
        "predicted_flow": predicted_flow,
        "normal_flow": normal_flow,
        "is_peak_alert": is_peak,
        "peak_ratio": round(predicted_flow / normal_flow, 2) if normal_flow > 0 else 0
    }

def check_and_create_crowd_alerts(db: Session, station: Station, zone: Zone, count: int):
    occupancy_rate = count / zone.max_capacity if zone.max_capacity > 0 else 0
    
    active_alerts = db.query(Alert).filter(
        Alert.zone_id == zone.id,
        Alert.alert_type.in_(["crowd_yellow", "crowd_red"]),
        Alert.is_active == True
    ).all()
    
    for alert in active_alerts:
        alert.is_active = False
        alert.resolved_at = datetime.utcnow()
    
    if occupancy_rate >= RED_ALERT_THRESHOLD:
        alert = Alert(
            station_id=station.id,
            zone_id=zone.id,
            alert_type="crowd_red",
            severity="critical",
            message=f"区域 {zone.name} 红色严重拥挤预警：当前 {count} 人，达到容量的 {occupancy_rate * 100:.1f}%",
            is_active=True
        )
        db.add(alert)
        adjust_security_gates(db, zone)
        
    elif occupancy_rate >= YELLOW_ALERT_THRESHOLD:
        alert = Alert(
            station_id=station.id,
            zone_id=zone.id,
            alert_type="crowd_yellow",
            severity="warning",
            message=f"区域 {zone.name} 黄色拥挤预警：当前 {count} 人，达到容量的 {occupancy_rate * 100:.1f}%",
            is_active=True
        )
        db.add(alert)
        adjust_security_gates(db, zone)

def adjust_security_gates(db: Session, zone: Zone):
    gates = db.query(SecurityGate).filter(
        SecurityGate.zone_id == zone.id,
        SecurityGate.is_faulty == False
    ).all()
    
    if not gates:
        return
    
    open_gates = [g for g in gates if g.is_open]
    closed_gates = [g for g in gates if not g.is_open]
    
    if closed_gates:
        gate_to_open = closed_gates[0]
        gate_to_open.is_open = True
        
        alert = Alert(
            station_id=zone.station_id,
            zone_id=zone.id,
            alert_type="gate_auto_open",
            severity="info",
            message=f"因客流拥挤，自动开启安检通道 {gate_to_open.gate_number}",
            is_active=True
        )
        db.add(alert)

def adjust_security_gates_for_peak(db: Session, station: Station):
    zones = db.query(Zone).filter(Zone.station_id == station.id).all()
    
    for zone in zones:
        gates = db.query(SecurityGate).filter(
            SecurityGate.zone_id == zone.id,
            SecurityGate.is_faulty == False
        ).all()
        
        if not gates:
            continue
        
        for gate in gates:
            if not gate.is_open:
                gate.is_open = True
        
        alert = Alert(
            station_id=station.id,
            zone_id=zone.id,
            alert_type="holiday_gate_adjust",
            severity="info",
            message=f"节假日高峰期间，区域 {zone.name} 已开启所有可用安检通道",
            is_active=True
        )
        db.add(alert)
