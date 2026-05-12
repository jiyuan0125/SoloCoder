from datetime import datetime, timedelta
from typing import Optional, List
from sqlalchemy.orm import Session
from app.models import (
    Vehicle, SignalPriority, SignalPriorityRequest, 
    ArrivalRecord, OperationParameter, Direction,
    Station, RouteStation, Schedule, ScheduleStatus
)
import os
from dotenv import load_dotenv

load_dotenv()

SIGNAL_PRIORITY_RADIUS = float(os.getenv("SIGNAL_PRIORITY_RADIUS", "300"))
PUNCTUALITY_THRESHOLD = 2


def check_and_request_signal_priority(db: Session, vehicle_id: int, distance: float, intersection_name: str) -> dict:
    vehicle = db.query(Vehicle).filter(Vehicle.id == vehicle_id).first()
    if not vehicle:
        return {"granted": False, "message": "Vehicle not found"}
    
    if distance > SIGNAL_PRIORITY_RADIUS:
        return {"granted": False, "message": f"Distance {distance}m exceeds threshold {SIGNAL_PRIORITY_RADIUS}m"}
    
    today = datetime.now().date()
    daily_count = db.query(SignalPriorityRequest).filter(
        SignalPriorityRequest.vehicle_id == vehicle_id,
        SignalPriorityRequest.request_time >= datetime.combine(today, datetime.min.time())
    ).count()
    
    op_param = None
    if vehicle.current_route_id:
        op_param = db.query(OperationParameter).filter(
            OperationParameter.route_id == vehicle.current_route_id
        ).first()
    
    max_requests = op_param.max_signal_priority_daily if op_param else 20
    
    if daily_count >= max_requests:
        return {"granted": False, "message": f"Daily limit {max_requests} reached"}
    
    signal = db.query(SignalPriority).filter(
        SignalPriority.intersection_name == intersection_name,
        SignalPriority.is_active == True
    ).first()
    
    granted = signal is not None
    
    request = SignalPriorityRequest(
        vehicle_id=vehicle_id,
        intersection_name=intersection_name,
        granted=granted
    )
    db.add(request)
    db.commit()
    db.refresh(request)
    
    return {
        "granted": granted,
        "request_id": request.id,
        "remaining_requests": max_requests - (daily_count + 1),
        "message": "Signal priority request processed"
    }


def is_on_time(planned: datetime, actual: datetime) -> bool:
    diff = abs((actual - planned).total_seconds() / 60)
    return diff <= PUNCTUALITY_THRESHOLD


def check_vehicle_crowd(vehicle: Vehicle) -> bool:
    if vehicle.capacity <= 0:
        return False
    load_ratio = vehicle.current_load / vehicle.capacity
    return load_ratio > 0.8


def get_route_interval(db: Session, route_id: int, direction: Direction) -> int:
    op_param = db.query(OperationParameter).filter(
        OperationParameter.route_id == route_id
    ).first()
    
    if not op_param:
        return int(os.getenv("DEFAULT_INTERVAL", "10"))
    
    route_stations = db.query(RouteStation).filter(
        RouteStation.route_id == route_id,
        RouteStation.direction == direction
    ).order_by(RouteStation.sequence).all()
    
    if not route_stations:
        return op_param.default_interval
    
    station_ids = [rs.station_id for rs in route_stations]
    
    signals = db.query(SignalPriority).filter(
        SignalPriority.station_id.in_(station_ids),
        SignalPriority.is_active == True
    ).all()
    
    if signals:
        min_interval = min(s.priority_interval for s in signals)
        return min(min_interval, op_param.default_interval)
    
    return op_param.default_interval


def calculate_route_timing(db: Session, route_id: int, direction: Direction) -> List[dict]:
    route_stations = db.query(RouteStation).filter(
        RouteStation.route_id == route_id,
        RouteStation.direction == direction
    ).order_by(RouteStation.sequence).all()
    
    if not route_stations:
        return []
    
    result = []
    cumulative_time = 0
    
    for rs in route_stations:
        cumulative_time += rs.travel_time_from_prev
        result.append({
            "station_id": rs.station_id,
            "sequence": rs.sequence,
            "arrival_offset_minutes": cumulative_time,
            "stop_time": rs.stop_time
        })
        cumulative_time += rs.stop_time
    
    return result


def reschedule_pending_schedules(db: Session, route_id: int, direction: Optional[Direction] = None):
    from app.scheduler import generate_schedule_for_route
    
    now = datetime.now()
    
    query = db.query(Schedule).filter(
        Schedule.route_id == route_id,
        Schedule.status == ScheduleStatus.PENDING,
        Schedule.start_time > now
    )
    
    if direction:
        query = query.filter(Schedule.direction == direction)
    
    pending_schedules = query.all()
    
    for schedule in pending_schedules:
        new_schedule_data = recalculate_schedule_timing(db, schedule)
        if new_schedule_data:
            schedule.schedule_data = new_schedule_data
            schedule.updated_at = datetime.now()
    
    db.commit()


def recalculate_schedule_timing(db: Session, schedule: Schedule) -> str:
    import json
    route_timing = calculate_route_timing(db, schedule.route_id, schedule.direction)
    
    if not route_timing:
        return None
    
    stations = []
    current_time = schedule.start_time
    
    for timing in route_timing:
        arrival_time = current_time + timedelta(minutes=timing["arrival_offset_minutes"])
        station = db.query(Station).filter(Station.id == timing["station_id"]).first()
        
        stations.append({
            "station_id": timing["station_id"],
            "station_name": station.name if station else "",
            "sequence": timing["sequence"],
            "planned_arrival_time": arrival_time.isoformat()
        })
    
    schedule_data = {
        "route_id": schedule.route_id,
        "vehicle_id": schedule.vehicle_id,
        "direction": schedule.direction.value,
        "start_time": schedule.start_time.isoformat(),
        "end_time": schedule.end_time.isoformat(),
        "stations": stations
    }
    
    return json.dumps(schedule_data, ensure_ascii=False)
