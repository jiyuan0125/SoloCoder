from datetime import datetime, timedelta
from typing import List, Optional
from sqlalchemy.orm import Session
from .models import (
    Alert, AlertStatus, AlertPriority,
    RescueForce, ForceStatus,
    Event, EventStatus, EventLevel,
    EventForceAssignment, TimelineEntry, RescueResult
)
from .config import settings
from .schemas import AlertCreate, ForceCreate, RescueResultCreate


def generate_event_code(db: Session) -> str:
    prefix = "EV"
    today = datetime.now().strftime("%Y%m%d")
    count = db.query(Event).filter(
        Event.event_code.like(f"{prefix}{today}%")
    ).count()
    return f"{prefix}{today}{count + 1:04d}"


def check_urgent_alert(db: Session, alert: Alert) -> bool:
    window_start = datetime.utcnow() - timedelta(
        minutes=settings.URGENT_ALERT_WINDOW_MINUTES
    )
    
    same_location_alerts = db.query(Alert).filter(
        Alert.location == alert.location,
        Alert.id != alert.id,
        Alert.alert_time >= window_start,
        Alert.status != AlertStatus.REJECTED
    ).all()
    
    if len(same_location_alerts) + 1 >= settings.URGENT_ALERT_COUNT:
        return True
    return False


def create_timeline_entry(
    db: Session,
    event_id: int,
    entry_type: str,
    description: str,
    operator: Optional[str] = None
) -> TimelineEntry:
    entry = TimelineEntry(
        event_id=event_id,
        entry_type=entry_type,
        entry_time=datetime.utcnow(),
        description=description,
        operator=operator
    )
    db.add(entry)
    db.commit()
    db.refresh(entry)
    return entry


def create_event_for_alert(
    db: Session,
    alert: Alert,
    verified_by: str,
    verification_notes: Optional[str] = None
) -> Event:
    event_code = generate_event_code(db)
    
    event = Event(
        event_code=event_code,
        status=EventStatus.VERIFIED,
        level=EventLevel.NORMAL,
        location=alert.location,
        location_lat=alert.location_lat,
        location_lon=alert.location_lon,
        description=alert.description,
        received_at=alert.alert_time,
        verified_at=datetime.utcnow()
    )
    
    db.add(event)
    db.flush()
    
    alert.event_id = event.id
    alert.status = AlertStatus.VERIFIED
    alert.verified_at = datetime.utcnow()
    alert.verified_by = verified_by
    alert.verification_notes = verification_notes
    
    create_timeline_entry(
        db,
        event.id,
        "alert_verified",
        f"报警信息已核实，核实人：{verified_by}",
        verified_by
    )
    
    db.commit()
    db.refresh(event)
    
    return event


def get_available_forces(db: Session, event_id: int) -> List[RescueForce]:
    return db.query(RescueForce).filter(
        RescueForce.status == ForceStatus.AVAILABLE
    ).order_by(
        RescueForce.estimated_arrival_minutes.asc()
    ).all()


def dispatch_force(
    db: Session,
    event: Event,
    force: RescueForce,
    operator: Optional[str] = None
) -> EventForceAssignment:
    if event.status not in [EventStatus.VERIFIED, EventStatus.DISPATCHED]:
        raise ValueError("Only verified or dispatched events can dispatch forces")
    
    if force.status != ForceStatus.AVAILABLE:
        raise ValueError("This force is not available")
    
    assignment = EventForceAssignment(
        event_id=event.id,
        force_id=force.id,
        assigned_at=datetime.utcnow()
    )
    
    db.add(assignment)
    
    force.status = ForceStatus.DISPATCHED
    force.current_event_id = event.id
    
    if event.status == EventStatus.VERIFIED:
        event.status = EventStatus.DISPATCHED
        event.dispatched_at = datetime.utcnow()
    
    create_timeline_entry(
        db,
        event.id,
        "force_dispatched",
        f"力量已调度：{force.name}",
        operator
    )
    
    db.commit()
    db.refresh(assignment)
    db.refresh(force)
    db.refresh(event)
    
    return assignment


def mark_force_arrived(
    db: Session,
    assignment: EventForceAssignment,
    operator: Optional[str] = None
) -> EventForceAssignment:
    event = db.query(Event).filter(Event.id == assignment.event_id).first()
    force = db.query(RescueForce).filter(RescueForce.id == assignment.force_id).first()
    
    if event.status not in [EventStatus.DISPATCHED, EventStatus.RESCUING]:
        raise ValueError("Event must be dispatched or rescuing")
    
    assignment.arrived_at = datetime.utcnow()
    force.status = ForceStatus.RESCUING
    
    if event.status == EventStatus.DISPATCHED:
        event.status = EventStatus.RESCUING
        event.rescuing_at = datetime.utcnow()
    
    create_timeline_entry(
        db,
        event.id,
        "force_arrived",
        f"力量已到达现场：{force.name}",
        operator
    )
    
    db.commit()
    db.refresh(assignment)
    db.refresh(force)
    db.refresh(event)
    
    return assignment


def calculate_response_time(event: Event) -> Optional[int]:
    if event.rescuing_at and event.received_at:
        delta = event.rescuing_at - event.received_at
        return int(delta.total_seconds() / 60)
    return None


def determine_event_level(response_time: Optional[int]) -> EventLevel:
    if response_time is None:
        return EventLevel.NORMAL
    if response_time >= settings.RED_RESPONSE_THRESHOLD:
        return EventLevel.RED
    if response_time >= settings.YELLOW_RESPONSE_THRESHOLD:
        return EventLevel.YELLOW
    return EventLevel.NORMAL


def complete_event(
    db: Session,
    event: Event,
    result_data: RescueResultCreate,
    operator: Optional[str] = None
) -> Event:
    if event.status != EventStatus.RESCUING:
        raise ValueError("Event must be rescuing to complete")
    
    response_time = calculate_response_time(event)
    event_level = determine_event_level(response_time)
    
    event.status = EventStatus.COMPLETED
    event.completed_at = datetime.utcnow()
    event.response_time_minutes = response_time
    event.level = event_level
    
    result = RescueResult(
        event_id=event.id,
        people_rescued=result_data.people_rescued,
        people_injured=result_data.people_injured,
        people_deceased=result_data.people_deceased,
        is_successful=result_data.is_successful,
        summary=result_data.summary,
        notes=result_data.notes
    )
    db.add(result)
    
    assignments = db.query(EventForceAssignment).filter(
        EventForceAssignment.event_id == event.id
    ).all()
    
    for assignment in assignments:
        assignment.departed_at = datetime.utcnow()
        force = db.query(RescueForce).filter(
            RescueForce.id == assignment.force_id
        ).first()
        if force:
            force.status = ForceStatus.AVAILABLE
            force.current_event_id = None
    
    level_text = "正常"
    if event_level == EventLevel.YELLOW:
        level_text = "黄色关注"
    elif event_level == EventLevel.RED:
        level_text = "红色预警"
    
    create_timeline_entry(
        db,
        event.id,
        "event_completed",
        f"救援任务完成。响应时间：{response_time}分钟，级别：{level_text}，成功救助：{result_data.people_rescued}人",
        operator
    )
    
    db.commit()
    db.refresh(event)
    db.refresh(result)
    
    return event


def get_monthly_stats(db: Session, year: int, month: int):
    start_date = datetime(year, month, 1)
    if month == 12:
        end_date = datetime(year + 1, 1, 1)
    else:
        end_date = datetime(year, month + 1, 1)
    
    total_alerts = db.query(Alert).filter(
        Alert.alert_time >= start_date,
        Alert.alert_time < end_date
    ).count()
    
    total_events = db.query(Event).filter(
        Event.received_at >= start_date,
        Event.received_at < end_date
    ).count()
    
    results = db.query(RescueResult).filter(
        RescueResult.created_at >= start_date,
        RescueResult.created_at < end_date
    ).all()
    
    total_people_rescued = sum(r.people_rescued for r in results)
    
    return {
        "year": year,
        "month": month,
        "total_alerts": total_alerts,
        "total_events": total_events,
        "total_people_rescued": total_people_rescued
    }


def link_alerts_to_existing_event(
    db: Session,
    new_alert: Alert,
    existing_event: Event
) -> None:
    new_alert.event_id = existing_event.id
    existing_event.description = (
        existing_event.description or ""
    ) + f"\n\n[新报警 #{new_alert.id}] {new_alert.description or ''}"
    
    if new_alert.priority == AlertPriority.URGENT:
        existing_event.level = EventLevel.RED
    
    create_timeline_entry(
        db,
        existing_event.id,
        "alert_linked",
        f"关联新报警 #{new_alert.id} - {new_alert.location}"
    )
    
    db.commit()
    db.refresh(existing_event)
    db.refresh(new_alert)


def find_existing_unverified_event(
    db: Session,
    location: str,
    window_minutes: int
) -> Optional[Event]:
    window_start = datetime.utcnow() - timedelta(minutes=window_minutes)
    
    return db.query(Event).filter(
        Event.location == location,
        Event.received_at >= window_start,
        Event.status.in_([EventStatus.RECEIVED, EventStatus.VERIFIED])
    ).first()


def reject_alert(
    db: Session,
    alert: Alert,
    verified_by: str,
    verification_notes: Optional[str] = None
) -> Alert:
    alert.status = AlertStatus.REJECTED
    alert.verified_at = datetime.utcnow()
    alert.verified_by = verified_by
    alert.verification_notes = verification_notes
    db.commit()
    db.refresh(alert)
    return alert
