from datetime import datetime
from typing import List, Tuple
from sqlalchemy.orm import Session
from ..models import Flight, Route, CapacityStatus, FlowRecord, RouteStatusHistory, FlightType
from ..schemas import RouteStatus


def get_active_flights(db: Session, route_id: int) -> List[Flight]:
    return db.query(Flight).filter(
        Flight.route_id == route_id,
        Flight.is_completed == False,
        Flight.is_waiting == False
    ).all()


def get_waiting_flights(db: Session, route_id: int) -> List[Flight]:
    return db.query(Flight).filter(
        Flight.route_id == route_id,
        Flight.is_completed == False,
        Flight.is_waiting == True
    ).all()


def calculate_utilization(active_count: int, capacity: int) -> float:
    if capacity <= 0:
        return 1.0
    return active_count / capacity


def determine_new_status(
    current_status: CapacityStatus,
    utilization: float,
    busy_threshold: float,
    saturated_threshold: float
) -> Tuple[CapacityStatus, str]:
    if current_status == CapacityStatus.RESTRICTED:
        if utilization < busy_threshold:
            return CapacityStatus.BUSY, "限流解除，恢复繁忙状态"
        return current_status, ""
    
    if current_status == CapacityStatus.SATURATED:
        if utilization < busy_threshold:
            return CapacityStatus.BUSY, "流量下降，从饱和恢复到繁忙"
        if utilization > saturated_threshold:
            return CapacityStatus.RESTRICTED, "流量超过饱和阈值，启动紧急管制限流"
        return current_status, ""
    
    if current_status == CapacityStatus.BUSY:
        if utilization > saturated_threshold:
            return CapacityStatus.RESTRICTED, "流量超过饱和阈值，启动紧急管制限流"
        if utilization < busy_threshold:
            return current_status, "流量正常，但繁忙状态需保持一段时间"
        return current_status, ""
    
    if current_status == CapacityStatus.IDLE:
        if utilization >= saturated_threshold:
            return CapacityStatus.RESTRICTED, "流量突发超过饱和阈值，启动紧急管制限流"
        if utilization >= busy_threshold:
            return CapacityStatus.BUSY, f"流量达到 {utilization*100:.1f}%，进入繁忙状态"
        return current_status, ""
    
    if utilization >= saturated_threshold:
        return CapacityStatus.RESTRICTED, "流量超过饱和阈值，启动紧急管制限流"
    if utilization >= busy_threshold:
        return CapacityStatus.BUSY, f"流量达到 {utilization*100:.1f}%，进入繁忙状态"
    
    return current_status, ""


def update_route_status(db: Session, route: Route, reason: str = "") -> RouteStatus:
    active = get_active_flights(db, route.id)
    waiting = get_waiting_flights(db, route.id)
    active_count = len(active)
    waiting_count = len(waiting)
    utilization = calculate_utilization(active_count, route.capacity)
    
    from sqlalchemy import func
    unresolved_conflicts = db.query(func.count(Conflict.id)).filter(
        Conflict.route_id == route.id,
        Conflict.resolved == False
    ).scalar() or 0
    
    new_status, change_reason = determine_new_status(
        route.status,
        utilization,
        route.busy_threshold,
        route.saturated_threshold
    )
    
    if new_status != route.status:
        history = RouteStatusHistory(
            route_id=route.id,
            old_status=route.status,
            new_status=new_status,
            reason=change_reason or reason
        )
        db.add(history)
        route.status = new_status
    
    flow_record = FlowRecord(
        route_id=route.id,
        active_flights=active_count,
        waiting_flights=waiting_count,
        capacity=route.capacity,
        utilization=utilization,
        status=route.status
    )
    db.add(flow_record)
    
    db.commit()
    db.refresh(route)
    
    return RouteStatus(
        route_id=route.id,
        route_code=route.code,
        current_status=route.status,
        active_flights=active_count,
        waiting_flights=waiting_count,
        capacity=route.capacity,
        utilization=utilization,
        unresolved_conflicts=unresolved_conflicts
    )


def prioritize_waiting_queue(db: Session, route_id: int) -> List[Flight]:
    waiting = get_waiting_flights(db, route_id)
    
    def sort_key(flight: Flight):
        priority = 0 if flight.flight_type == FlightType.INTERNATIONAL else 1
        return (priority, flight.estimated_entry_time)
    
    return sorted(waiting, key=sort_key)


def can_accept_flight(db: Session, route: Route) -> bool:
    if route.status == CapacityStatus.RESTRICTED:
        return False
    
    active = get_active_flights(db, route.id)
    if route.status == CapacityStatus.SATURATED:
        return False
    
    return len(active) < route.capacity


from ..models import Conflict
