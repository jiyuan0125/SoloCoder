from sqlalchemy.orm import Session
from sqlalchemy import func, and_, or_
from datetime import datetime, timedelta
from typing import List, Optional
from . import models, schemas
from .models import (
    AlertStatus, AlertUrgency, EventStatus, ResponseLevel,
    TimelineEventType, ForceStatus
)


def check_alert_aggregation(db: Session, alert: models.Alert) -> models.Alert:
    thirty_minutes_ago = datetime.utcnow() - timedelta(minutes=30)
    recent_alerts = db.query(models.Alert).filter(
        and_(
            models.Alert.location == alert.location,
            models.Alert.created_at >= thirty_minutes_ago,
            models.Alert.id != alert.id,
            or_(
                models.Alert.status == AlertStatus.PENDING,
                models.Alert.status == AlertStatus.VERIFIED
            ),
            models.Alert.source != alert.source
        )
    ).all()

    if len(recent_alerts) >= 1:
        alert.urgency = AlertUrgency.EMERGENCY
        for other_alert in recent_alerts:
            other_alert.urgency = AlertUrgency.EMERGENCY

        db.commit()
        db.refresh(alert)

    return alert


def create_alert(db: Session, alert: schemas.AlertCreate) -> models.Alert:
    db_alert = models.Alert(
        location=alert.location,
        latitude=alert.latitude,
        longitude=alert.longitude,
        description=alert.description,
        source=alert.source
    )
    db.add(db_alert)
    db.commit()
    db.refresh(db_alert)

    check_alert_aggregation(db, db_alert)

    add_timeline_event(
        db,
        event_type=TimelineEventType.ALERT_RECEIVED,
        description=f"收到新报警: {alert.location}, 来源: {alert.source}",
        alert_id=db_alert.id
    )

    return db_alert


def get_alert(db: Session, alert_id: int) -> Optional[models.Alert]:
    return db.query(models.Alert).filter(models.Alert.id == alert_id).first()


def get_alerts(db: Session, skip: int = 0, limit: int = 100) -> List[models.Alert]:
    return db.query(models.Alert).order_by(models.Alert.created_at.desc()).offset(skip).limit(limit).all()


def verify_alert(db: Session, alert_id: int, verify_data: schemas.AlertVerify) -> Optional[models.Alert]:
    alert = get_alert(db, alert_id)
    if not alert:
        return None

    if verify_data.verify_as_valid:
        alert.status = AlertStatus.VERIFIED
    else:
        alert.status = AlertStatus.DISMISSED

    alert.verified_at = datetime.utcnow()
    alert.verified_by = verify_data.verified_by
    alert.verification_notes = verify_data.verification_notes

    db.commit()
    db.refresh(alert)

    add_timeline_event(
        db,
        event_type=TimelineEventType.ALERT_VERIFIED,
        description=f"报警已{'核实' if verify_data.verify_as_valid else '驳回'}: {verify_data.verification_notes or '无备注'}",
        alert_id=alert.id,
        actor=verify_data.verified_by
    )

    return alert


def create_event_from_alert(db: Session, alert: models.Alert) -> models.Event:
    event = models.Event(
        title=f"搜救事件: {alert.location}",
        location=alert.location,
        latitude=alert.latitude,
        longitude=alert.longitude,
        description=alert.description,
        urgency=alert.urgency
    )
    db.add(event)
    db.commit()
    db.refresh(event)

    alert.event_id = event.id
    db.commit()

    add_timeline_event(
        db,
        event_type=TimelineEventType.EVENT_CREATED,
        description=f"创建搜救事件: {event.title}",
        event_id=event.id
    )

    return event


def get_event(db: Session, event_id: int) -> Optional[models.Event]:
    return db.query(models.Event).filter(models.Event.id == event_id).first()


def get_events(db: Session, skip: int = 0, limit: int = 100) -> List[models.Event]:
    return db.query(models.Event).order_by(models.Event.created_at.desc()).offset(skip).limit(limit).all()


def create_event(db: Session, event_data: schemas.EventCreate) -> models.Event:
    event = models.Event(
        title=event_data.title,
        location=event_data.location,
        latitude=event_data.latitude,
        longitude=event_data.longitude,
        description=event_data.description
    )
    db.add(event)
    db.commit()
    db.refresh(event)

    if event_data.alert_ids:
        for alert_id in event_data.alert_ids:
            alert = get_alert(db, alert_id)
            if alert and not alert.event_id:
                alert.event_id = event.id
                if alert.urgency == AlertUrgency.EMERGENCY:
                    event.urgency = AlertUrgency.EMERGENCY

        db.commit()
        db.refresh(event)

    add_timeline_event(
        db,
        event_type=TimelineEventType.EVENT_CREATED,
        description=f"创建搜救事件: {event.title}",
        event_id=event.id
    )

    return event


def verify_event(db: Session, event_id: int, verify_data: schemas.EventVerify) -> Optional[models.Event]:
    event = get_event(db, event_id)
    if not event or event.status != EventStatus.RECEIVED:
        return None

    event.status = EventStatus.VERIFIED
    event.verified_at = datetime.utcnow()

    db.commit()
    db.refresh(event)

    add_timeline_event(
        db,
        event_type=TimelineEventType.EVENT_VERIFIED,
        description=f"事件已核实: {verify_data.notes or '无备注'}",
        event_id=event.id,
        actor=verify_data.verified_by
    )

    return event


def get_available_forces(db: Session) -> List[models.SearchRescueForce]:
    return db.query(models.SearchRescueForce).filter(
        models.SearchRescueForce.status == ForceStatus.AVAILABLE
    ).order_by(models.SearchRescueForce.estimated_arrival_minutes.asc()).all()


def get_force_recommendations(db: Session, event_id: int) -> List[models.SearchRescueForce]:
    available_forces = get_available_forces(db)
    return sorted(available_forces, key=lambda f: f.estimated_arrival_minutes)


def is_force_available(db: Session, force_id: int) -> bool:
    force = db.query(models.SearchRescueForce).filter(models.SearchRescueForce.id == force_id).first()
    if not force:
        return False
    return force.status == ForceStatus.AVAILABLE


def assign_force_to_event(db: Session, event_id: int, force_id: int, assignment_notes: str = None) -> Optional[models.ForceAssignment]:
    event = get_event(db, event_id)
    if not event:
        return None

    if event.status not in [EventStatus.VERIFIED, EventStatus.DISPATCHED]:
        return None

    if not is_force_available(db, force_id):
        return None

    force = db.query(models.SearchRescueForce).filter(models.SearchRescueForce.id == force_id).first()

    assignment = models.ForceAssignment(
        event_id=event_id,
        force_id=force_id,
        assignment_notes=assignment_notes
    )
    db.add(assignment)

    force.status = ForceStatus.DISPATCHED
    event.status = EventStatus.DISPATCHED

    db.commit()
    db.refresh(assignment)

    add_timeline_event(
        db,
        event_type=TimelineEventType.FORCE_DISPATCHED,
        description=f"出动搜救力量: {force.name}",
        event_id=event_id
    )

    return assignment


def mark_force_arrived(db: Session, assignment_id: int) -> Optional[models.ForceAssignment]:
    assignment = db.query(models.ForceAssignment).filter(models.ForceAssignment.id == assignment_id).first()
    if not assignment or assignment.arrived_at:
        return None

    assignment.arrived_at = datetime.utcnow()
    assignment.force.status = ForceStatus.WORKING

    event = assignment.event
    if event.status == EventStatus.DISPATCHED:
        event.status = EventStatus.RESCUING

    db.commit()
    db.refresh(assignment)

    add_timeline_event(
        db,
        event_type=TimelineEventType.FORCE_ARRIVED,
        description=f"搜救力量到达现场: {assignment.force.name}",
        event_id=assignment.event_id
    )

    add_timeline_event(
        db,
        event_type=TimelineEventType.RESCUE_STARTED,
        description="开始现场施救",
        event_id=assignment.event_id
    )

    return assignment


def complete_event(db: Session, event_id: int, result_data: schemas.EventResultCreate) -> Optional[models.EventResult]:
    event = get_event(db, event_id)
    if not event or event.status == EventStatus.COMPLETED:
        return None

    response_time = None
    if event.created_at:
        time_diff = datetime.utcnow() - event.created_at
        response_time = int(time_diff.total_seconds() / 60)

    response_level = ResponseLevel.NORMAL
    if response_time:
        if response_time > 60:
            response_level = ResponseLevel.RED
        elif response_time > 30:
            response_level = ResponseLevel.YELLOW

    event.response_level = response_level
    event.status = EventStatus.COMPLETED

    result = models.EventResult(
        event_id=event_id,
        successful=result_data.successful,
        injured=result_data.injured,
        fatalities=result_data.fatalities,
        total_people=result_data.total_people,
        notes=result_data.notes,
        response_time_minutes=response_time
    )
    db.add(result)

    for assignment in event.force_assignments:
        if not assignment.completed_at:
            assignment.completed_at = datetime.utcnow()
            assignment.force.status = ForceStatus.AVAILABLE

    db.commit()
    db.refresh(event)
    db.refresh(result)

    add_timeline_event(
        db,
        event_type=TimelineEventType.RESCUE_ENDED,
        description="现场施救结束",
        event_id=event_id
    )

    add_timeline_event(
        db,
        event_type=TimelineEventType.EVENT_COMPLETED,
        description=f"事件结束, 响应等级: {response_level.value}, 成功救助: {result_data.successful}人",
        event_id=event_id
    )

    return result


def add_timeline_event(
    db: Session,
    event_type: TimelineEventType,
    description: str,
    event_id: int = None,
    alert_id: int = None,
    actor: str = None
) -> models.TimelineEvent:
    timeline_event = models.TimelineEvent(
        event_id=event_id,
        alert_id=alert_id,
        event_type=event_type,
        description=description,
        actor=actor
    )
    db.add(timeline_event)
    db.commit()
    db.refresh(timeline_event)
    return timeline_event


def get_event_timeline(db: Session, event_id: int) -> List[models.TimelineEvent]:
    return db.query(models.TimelineEvent).filter(
        models.TimelineEvent.event_id == event_id
    ).order_by(models.TimelineEvent.timestamp.asc()).all()


def get_event_alerts(db: Session, event_id: int) -> List[models.Alert]:
    return db.query(models.Alert).filter(models.Alert.event_id == event_id).all()


def get_event_forces(db: Session, event_id: int) -> List[models.ForceAssignment]:
    return db.query(models.ForceAssignment).filter(
        models.ForceAssignment.event_id == event_id
    ).all()


def get_event_result(db: Session, event_id: int) -> Optional[models.EventResult]:
    return db.query(models.EventResult).filter(models.EventResult.event_id == event_id).first()


def get_monthly_stats(db: Session, year: int, month: int) -> schemas.MonthlyStats:
    start_date = datetime(year, month, 1)
    if month == 12:
        end_date = datetime(year + 1, 1, 1)
    else:
        end_date = datetime(year, month + 1, 1)

    alert_count = db.query(models.Alert).filter(
        and_(
            models.Alert.created_at >= start_date,
            models.Alert.created_at < end_date
        )
    ).count()

    event_count = db.query(models.Event).filter(
        and_(
            models.Event.created_at >= start_date,
            models.Event.created_at < end_date
        )
    ).count()

    successful_count = db.query(func.sum(models.EventResult.successful)).filter(
        and_(
            models.EventResult.completed_at >= start_date,
            models.EventResult.completed_at < end_date
        )
    ).scalar() or 0

    return schemas.MonthlyStats(
        year=year,
        month=month,
        alert_count=alert_count,
        event_count=event_count,
        successful_rescues=int(successful_count)
    )


def create_rescue_force(db: Session, force_data: schemas.RescueForceCreate) -> models.SearchRescueForce:
    force = models.SearchRescueForce(
        name=force_data.name,
        type=force_data.type,
        capacity=force_data.capacity,
        current_location=force_data.current_location,
        latitude=force_data.latitude,
        longitude=force_data.longitude,
        estimated_arrival_minutes=force_data.estimated_arrival_minutes
    )
    db.add(force)
    db.commit()
    db.refresh(force)
    return force


def get_rescue_forces(db: Session, skip: int = 0, limit: int = 100) -> List[models.SearchRescueForce]:
    return db.query(models.SearchRescueForce).offset(skip).limit(limit).all()


def get_rescue_force(db: Session, force_id: int) -> Optional[models.SearchRescueForce]:
    return db.query(models.SearchRescueForce).filter(models.SearchRescueForce.id == force_id).first()


def get_pending_alerts(db: Session) -> List[models.Alert]:
    return db.query(models.Alert).filter(
        models.Alert.status == AlertStatus.PENDING
    ).order_by(models.Alert.created_at.asc()).all()


def get_overdue_alerts(db: Session) -> List[models.Alert]:
    fifteen_minutes_ago = datetime.utcnow() - timedelta(minutes=15)
    return db.query(models.Alert).filter(
        and_(
            models.Alert.status == AlertStatus.PENDING,
            models.Alert.created_at < fifteen_minutes_ago
        )
    ).all()
