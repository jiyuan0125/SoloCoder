from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List
from .database import get_db
from . import schemas, services
from .models import EventStatus


router = APIRouter()


@router.get("/alerts/pending", response_model=List[schemas.Alert])
def get_pending_alerts(db: Session = Depends(get_db)):
    return services.get_pending_alerts(db)


@router.get("/alerts/overdue", response_model=List[schemas.Alert])
def get_overdue_alerts(db: Session = Depends(get_db)):
    return services.get_overdue_alerts(db)


@router.post("/alerts", response_model=schemas.Alert)
def create_alert(alert: schemas.AlertCreate, db: Session = Depends(get_db)):
    return services.create_alert(db, alert)


@router.get("/alerts", response_model=List[schemas.Alert])
def get_alerts(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_alerts(db, skip=skip, limit=limit)


@router.get("/alerts/{alert_id}", response_model=schemas.Alert)
def get_alert(alert_id: int, db: Session = Depends(get_db)):
    alert = services.get_alert(db, alert_id)
    if not alert:
        raise HTTPException(status_code=404, detail="Alert not found")
    return alert


@router.post("/alerts/{alert_id}/verify", response_model=schemas.Alert)
def verify_alert(alert_id: int, verify_data: schemas.AlertVerify, db: Session = Depends(get_db)):
    alert = services.verify_alert(db, alert_id, verify_data)
    if not alert:
        raise HTTPException(status_code=404, detail="Alert not found")
    return alert


@router.post("/alerts/{alert_id}/create-event", response_model=schemas.Event)
def create_event_from_alert(alert_id: int, db: Session = Depends(get_db)):
    alert = services.get_alert(db, alert_id)
    if not alert:
        raise HTTPException(status_code=404, detail="Alert not found")
    if alert.event_id:
        raise HTTPException(status_code=400, detail="Alert already linked to an event")
    if alert.status != AlertStatus.VERIFIED:
        raise HTTPException(status_code=400, detail="Alert must be verified first")
    return services.create_event_from_alert(db, alert)


@router.post("/events", response_model=schemas.Event)
def create_event(event: schemas.EventCreate, db: Session = Depends(get_db)):
    return services.create_event(db, event)


@router.get("/events", response_model=List[schemas.Event])
def get_events(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_events(db, skip=skip, limit=limit)


@router.get("/events/{event_id}", response_model=schemas.EventDetail)
def get_event(event_id: int, db: Session = Depends(get_db)):
    event = services.get_event(db, event_id)
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")
    return event


@router.post("/events/{event_id}/verify", response_model=schemas.Event)
def verify_event(event_id: int, verify_data: schemas.EventVerify, db: Session = Depends(get_db)):
    event = services.verify_event(db, event_id, verify_data)
    if not event:
        raise HTTPException(status_code=404, detail="Event not found or already verified")
    return event


@router.get("/events/{event_id}/alerts", response_model=List[schemas.Alert])
def get_event_alerts(event_id: int, db: Session = Depends(get_db)):
    event = services.get_event(db, event_id)
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")
    return services.get_event_alerts(db, event_id)


@router.get("/events/{event_id}/forces", response_model=List[schemas.ForceAssignment])
def get_event_forces(event_id: int, db: Session = Depends(get_db)):
    event = services.get_event(db, event_id)
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")
    return services.get_event_forces(db, event_id)


@router.get("/events/{event_id}/timeline", response_model=List[schemas.TimelineEvent])
def get_event_timeline(event_id: int, db: Session = Depends(get_db)):
    event = services.get_event(db, event_id)
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")
    return services.get_event_timeline(db, event_id)


@router.get("/events/{event_id}/result", response_model=schemas.EventResult)
def get_event_result(event_id: int, db: Session = Depends(get_db)):
    event = services.get_event(db, event_id)
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")
    result = services.get_event_result(db, event_id)
    if not result:
        raise HTTPException(status_code=404, detail="Event result not found")
    return result


@router.get("/events/{event_id}/force-recommendations", response_model=List[schemas.AlertRecommendation])
def get_force_recommendations(event_id: int, db: Session = Depends(get_db)):
    event = services.get_event(db, event_id)
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")
    if event.status != EventStatus.VERIFIED:
        raise HTTPException(status_code=400, detail="Event must be verified first")
    return services.get_force_recommendations(db, event_id)


@router.post("/events/{event_id}/assign-force", response_model=schemas.ForceAssignment)
def assign_force(event_id: int, assignment: schemas.ForceAssignmentCreate, db: Session = Depends(get_db)):
    result = services.assign_force_to_event(db, event_id, assignment.force_id, assignment.assignment_notes)
    if not result:
        raise HTTPException(
            status_code=400,
            detail="Cannot assign force. Check event status (must be verified) and force availability."
        )
    return result


@router.post("/force-assignments/{assignment_id}/arrive", response_model=schemas.ForceAssignment)
def mark_force_arrived(assignment_id: int, db: Session = Depends(get_db)):
    assignment = services.mark_force_arrived(db, assignment_id)
    if not assignment:
        raise HTTPException(status_code=404, detail="Assignment not found or already arrived")
    return assignment


@router.post("/events/{event_id}/complete", response_model=schemas.EventResult)
def complete_event(event_id: int, result: schemas.EventResultCreate, db: Session = Depends(get_db)):
    event = services.get_event(db, event_id)
    if not event:
        raise HTTPException(status_code=404, detail="Event not found")
    if event.status != EventStatus.RESCUING:
        raise HTTPException(status_code=400, detail="Event must be in rescuing status to complete")

    event_result = services.complete_event(db, event_id, result)
    if not event_result:
        raise HTTPException(status_code=400, detail="Failed to complete event")
    return event_result


@router.post("/rescue-forces", response_model=schemas.RescueForce)
def create_rescue_force(force: schemas.RescueForceCreate, db: Session = Depends(get_db)):
    return services.create_rescue_force(db, force)


@router.get("/rescue-forces", response_model=List[schemas.RescueForce])
def get_rescue_forces(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return services.get_rescue_forces(db, skip=skip, limit=limit)


@router.get("/rescue-forces/{force_id}", response_model=schemas.RescueForce)
def get_rescue_force(force_id: int, db: Session = Depends(get_db)):
    force = services.get_rescue_force(db, force_id)
    if not force:
        raise HTTPException(status_code=404, detail="Rescue force not found")
    return force


@router.get("/stats/monthly", response_model=schemas.MonthlyStats)
def get_monthly_stats(
    year: int = Query(..., description="Year (e.g., 2024)"),
    month: int = Query(..., ge=1, le=12, description="Month (1-12)"),
    db: Session = Depends(get_db)
):
    return services.get_monthly_stats(db, year, month)
