from datetime import datetime, timedelta, time
from typing import List, Optional
from sqlalchemy.orm import Session
import json

from app.models import (
    Route, Vehicle, OperationParameter, RouteStation, 
    Station, Schedule, ScheduleStatus, Direction,
    ArrivalRecord
)
from app.utils import get_route_interval, calculate_route_timing, reschedule_pending_schedules


def parse_time(time_str: str) -> time:
    hh, mm = map(int, time_str.split(':'))
    return time(hh, mm)


def _create_single_trip(
    db: Session,
    route_id: int,
    vehicle_id: int,
    direction: Direction,
    start_time: datetime,
    route_timing: List[dict]
) -> Schedule:
    total_trip_time = route_timing[-1]["arrival_offset_minutes"] + route_timing[-1]["stop_time"]
    end_time_trip = start_time + timedelta(minutes=total_trip_time)
    
    stations = []
    for timing in route_timing:
        arrival_time = start_time + timedelta(minutes=timing["arrival_offset_minutes"])
        station = db.query(Station).filter(Station.id == timing["station_id"]).first()
        
        stations.append({
            "station_id": timing["station_id"],
            "station_name": station.name if station else "",
            "sequence": timing["sequence"],
            "planned_arrival_time": arrival_time.isoformat()
        })
    
    schedule_data = {
        "route_id": route_id,
        "vehicle_id": vehicle_id,
        "direction": direction.value,
        "start_time": start_time.isoformat(),
        "end_time": end_time_trip.isoformat(),
        "stations": stations
    }
    
    schedule = Schedule(
        route_id=route_id,
        vehicle_id=vehicle_id,
        direction=direction,
        start_time=start_time,
        end_time=end_time_trip,
        status=ScheduleStatus.PENDING,
        schedule_data=json.dumps(schedule_data, ensure_ascii=False)
    )
    
    db.add(schedule)
    return schedule


def _get_opposite_direction(direction: Direction) -> Direction:
    return Direction.DOWN if direction == Direction.UP else Direction.UP


def generate_schedule_for_route(
    db: Session,
    route_id: int,
    start_date: Optional[datetime] = None,
    vehicle_ids: Optional[List[int]] = None
) -> List[Schedule]:
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise ValueError(f"Route with id {route_id} not found")
    
    op_param = db.query(OperationParameter).filter(
        OperationParameter.route_id == route_id
    ).first()
    
    if not op_param:
        raise ValueError(f"Operation parameters not found for route {route_id}")
    
    if start_date is None:
        start_date = datetime.now().replace(hour=0, minute=0, second=0, microsecond=0)
    else:
        start_date = start_date.replace(hour=0, minute=0, second=0, microsecond=0)
    
    start_time = parse_time(op_param.start_time)
    end_time = parse_time(op_param.end_time)
    
    start_datetime = datetime.combine(start_date.date(), start_time)
    end_datetime = datetime.combine(start_date.date(), end_time)
    
    if vehicle_ids is None:
        available_vehicles = db.query(Vehicle).filter(Vehicle.status == "idle").all()
        vehicle_ids = [v.id for v in available_vehicles]
    
    if not vehicle_ids:
        raise ValueError("No available vehicles for scheduling")
    
    route_timing_up = calculate_route_timing(db, route_id, Direction.UP)
    route_timing_down = calculate_route_timing(db, route_id, Direction.DOWN)
    
    if not route_timing_up and not route_timing_down:
        raise ValueError("Route has no stations configured")
    
    interval = get_route_interval(db, route_id, Direction.UP)
    turnaround_time = op_param.turnaround_time
    
    schedules = []
    
    for vehicle_idx, vehicle_id in enumerate(vehicle_ids):
        vehicle_available_time = start_datetime + timedelta(minutes=vehicle_idx * interval)
        current_direction = Direction.UP
        
        while vehicle_available_time < end_datetime:
            route_timing = route_timing_up if current_direction == Direction.UP else route_timing_down
            
            if not route_timing:
                current_direction = _get_opposite_direction(current_direction)
                vehicle_available_time += timedelta(minutes=turnaround_time)
                continue
            
            total_trip_time = route_timing[-1]["arrival_offset_minutes"] + route_timing[-1]["stop_time"]
            trip_end_time = vehicle_available_time + timedelta(minutes=total_trip_time)
            
            if vehicle_available_time >= end_datetime:
                break
            
            schedule = _create_single_trip(
                db, route_id, vehicle_id, current_direction, vehicle_available_time, route_timing
            )
            schedules.append(schedule)
            
            vehicle_available_time = trip_end_time + timedelta(minutes=turnaround_time)
            current_direction = _get_opposite_direction(current_direction)
    
    db.commit()
    
    for schedule in schedules:
        schedule_data = json.loads(schedule.schedule_data)
        for station_data in schedule_data["stations"]:
            arrival_record = ArrivalRecord(
                schedule_id=schedule.id,
                station_id=station_data["station_id"],
                planned_arrival_time=datetime.fromisoformat(station_data["planned_arrival_time"]),
                actual_arrival_time=None,
                is_on_time=None
            )
            db.add(arrival_record)
    
    db.commit()
    
    return schedules


def update_route_stations_and_reschedule(
    db: Session,
    route_id: int,
    direction: Optional[Direction] = None
):
    reschedule_pending_schedules(db, route_id, direction)


def get_schedules_for_route(
    db: Session,
    route_id: int,
    date: Optional[datetime] = None
) -> List[Schedule]:
    query = db.query(Schedule).filter(Schedule.route_id == route_id)
    
    if date:
        date_start = date.replace(hour=0, minute=0, second=0, microsecond=0)
        date_end = date_start + timedelta(days=1)
        query = query.filter(
            Schedule.start_time >= date_start,
            Schedule.start_time < date_end
        )
    
    return query.order_by(Schedule.start_time).all()


def get_punctuality_report(
    db: Session,
    route_id: Optional[int] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None
) -> List[dict]:
    from app.utils import is_on_time
    
    query = db.query(ArrivalRecord).filter(
        ArrivalRecord.actual_arrival_time.isnot(None)
    )
    
    if route_id:
        query = query.join(Schedule).filter(Schedule.route_id == route_id)
    
    if start_date:
        query = query.filter(ArrivalRecord.planned_arrival_time >= start_date)
    
    if end_date:
        query = query.filter(ArrivalRecord.planned_arrival_time <= end_date)
    
    records = query.all()
    
    if route_id:
        route = db.query(Route).filter(Route.id == route_id).first()
        route_name = route.name if route else ""
        
        total = len(records)
        on_time = sum(1 for r in records if r.is_on_time)
        
        return [{
            "route_id": route_id,
            "route_name": route_name,
            "total_schedules": total,
            "on_time_count": on_time,
            "punctuality_rate": (on_time / total * 100) if total > 0 else 0.0
        }]
    
    route_stats = {}
    
    for record in records:
        schedule = db.query(Schedule).filter(Schedule.id == record.schedule_id).first()
        if not schedule:
            continue
        
        rid = schedule.route_id
        if rid not in route_stats:
            route = db.query(Route).filter(Route.id == rid).first()
            route_stats[rid] = {
                "route_id": rid,
                "route_name": route.name if route else "",
                "total": 0,
                "on_time": 0
            }
        
        route_stats[rid]["total"] += 1
        if record.is_on_time:
            route_stats[rid]["on_time"] += 1
    
    result = []
    for stat in route_stats.values():
        total = stat["total"]
        result.append({
            "route_id": stat["route_id"],
            "route_name": stat["route_name"],
            "total_schedules": total,
            "on_time_count": stat["on_time"],
            "punctuality_rate": (stat["on_time"] / total * 100) if total > 0 else 0.0
        })
    
    return result
