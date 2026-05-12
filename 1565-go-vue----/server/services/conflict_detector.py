from datetime import datetime, timedelta
from typing import List, Tuple, Optional
from sqlalchemy.orm import Session
from ..models import Flight, Route, Conflict, RouteWaypoint, Waypoint
from .haversine import haversine_distance
from ..schemas import ConflictResponse


def get_time_difference(time1: datetime, time2: datetime) -> float:
    diff = abs((time1 - time2).total_seconds()) / 60.0
    return diff


def detect_horizontal_conflict(
    flight1: Flight, 
    flight2: Flight, 
    min_interval_minutes: int
) -> Optional[Conflict]:
    if flight1.direction != flight2.direction:
        return None
    
    time_diff = get_time_difference(flight1.estimated_entry_time, flight2.estimated_entry_time)
    
    if time_diff < min_interval_minutes:
        return Conflict(
            route_id=flight1.route_id,
            flight1_id=flight1.id,
            flight2_id=flight2.id,
            conflict_type="horizontal",
            description=f"航班 {flight1.flight_number} 和 {flight2.flight_number} 预计进入航路时间差 {time_diff:.1f} 分钟，小于最小间隔 {min_interval_minutes} 分钟",
            resolution_suggestion=f"建议将其中一个航班调整进入时间至少 {min_interval_minutes - time_diff:.1f} 分钟，或安排不同高度层"
        )
    return None


def detect_vertical_conflict(
    flight1: Flight, 
    flight2: Flight
) -> Optional[Conflict]:
    if flight1.altitude == flight2.altitude:
        return Conflict(
            route_id=flight1.route_id,
            flight1_id=flight1.id,
            flight2_id=flight2.id,
            conflict_type="vertical",
            description=f"航班 {flight1.flight_number} 和 {flight2.flight_number} 使用相同高度层 {flight1.altitude}m",
            resolution_suggestion=f"建议调整其中一个航班高度层，按照东单西双规则分配：向东飞行使用奇数高度层，向西使用偶数高度层"
        )
    return None


def get_available_altitude(direction: str, used_altitudes: List[float]) -> float:
    is_eastbound = direction.lower() in ["east", "e", "东", "东向"]
    base = 9000.0
    step = 300.0
    
    if is_eastbound:
        available = [base + i * step for i in range(0, 20) if (base + i * step) % (2 * step) != 0]
    else:
        available = [base + i * step for i in range(0, 20) if (base + i * step) % (2 * step) == 0]
    
    for alt in available:
        if alt not in used_altitudes:
            return alt
    return available[0]


def detect_conflicts(db: Session, route: Route, flights: List[Flight]) -> List[Conflict]:
    conflicts = []
    min_interval = route.min_interval_minutes
    
    for i in range(len(flights)):
        for j in range(i + 1, len(flights)):
            f1 = flights[i]
            f2 = flights[j]
            
            h_conflict = detect_horizontal_conflict(f1, f2, min_interval)
            if h_conflict:
                existing = db.query(Conflict).filter(
                    Conflict.route_id == route.id,
                    Conflict.resolved == False,
                    ((Conflict.flight1_id == f1.id) & (Conflict.flight2_id == f2.id)) |
                    ((Conflict.flight1_id == f2.id) & (Conflict.flight2_id == f1.id)),
                    Conflict.conflict_type == "horizontal"
                ).first()
                
                if not existing:
                    conflicts.append(h_conflict)
            
            v_conflict = detect_vertical_conflict(f1, f2)
            if v_conflict:
                existing = db.query(Conflict).filter(
                    Conflict.route_id == route.id,
                    Conflict.resolved == False,
                    ((Conflict.flight1_id == f1.id) & (Conflict.flight2_id == f2.id)) |
                    ((Conflict.flight1_id == f2.id) & (Conflict.flight2_id == f1.id)),
                    Conflict.conflict_type == "vertical"
                ).first()
                
                if not existing:
                    conflicts.append(v_conflict)
    
    return conflicts


def reevaluate_conflicts(db: Session, route: Route) -> List[Conflict]:
    today = datetime.now().date()
    
    db.query(Conflict).filter(
        Conflict.route_id == route.id,
        Conflict.resolved == False
    ).update({"resolved": True})
    
    active_flights = db.query(Flight).filter(
        Flight.route_id == route.id,
        Flight.is_completed == False
    ).all()
    
    new_conflicts = detect_conflicts(db, route, active_flights)
    
    for conflict in new_conflicts:
        db.add(conflict)
    
    db.commit()
    
    return new_conflicts
