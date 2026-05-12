from typing import List, Optional
from datetime import date
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import (
    Event,
    EventRegistration,
    EventRegistrationStatus,
)
from app.schemas import (
    EventCreate,
    EventUpdate,
    EventRegistrationCreate,
    EventRegistrationCancel,
    Event as EventSchema,
    EventRegistration as EventRegistrationSchema,
)
from app.services import (
    register_for_event,
    cancel_event_registration,
)


router = APIRouter(prefix="/events", tags=["events"])


@router.post("", response_model=EventSchema)
def create_event(event_data: EventCreate, db: Session = Depends(get_db)):
    event = Event(**event_data.model_dump())
    db.add(event)
    db.commit()
    db.refresh(event)
    return event


@router.get("", response_model=List[EventSchema])
def list_events(
    from_date: Optional[date] = None,
    to_date: Optional[date] = None,
    is_active: Optional[bool] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(Event)
    if from_date:
        query = query.filter(Event.event_date >= from_date)
    if to_date:
        query = query.filter(Event.event_date <= to_date)
    if is_active is not None:
        query = query.filter(Event.is_active == is_active)
    return query.order_by(Event.event_date.asc(), Event.start_time.asc()).offset(skip).limit(limit).all()


@router.get("/{event_id}", response_model=EventSchema)
def get_event(event_id: int, db: Session = Depends(get_db)):
    event = db.query(Event).filter(Event.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="活动不存在")
    return event


@router.put("/{event_id}", response_model=EventSchema)
def update_event(event_id: int, update_data: EventUpdate, db: Session = Depends(get_db)):
    event = db.query(Event).filter(Event.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="活动不存在")
    
    for key, value in update_data.model_dump(exclude_unset=True).items():
        setattr(event, key, value)
    
    db.commit()
    db.refresh(event)
    return event


@router.delete("/{event_id}")
def delete_event(event_id: int, db: Session = Depends(get_db)):
    event = db.query(Event).filter(Event.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="活动不存在")
    
    db.delete(event)
    db.commit()
    return {"message": "删除成功"}


@router.post("/register", response_model=EventRegistrationSchema)
def register_event(data: EventRegistrationCreate, db: Session = Depends(get_db)):
    registration, msg = register_for_event(db, data)
    if not registration:
        raise HTTPException(status_code=400, detail=msg)
    return registration


@router.post("/cancel", response_model=EventRegistrationSchema)
def cancel_registration(data: EventRegistrationCancel, db: Session = Depends(get_db)):
    registration, msg = cancel_event_registration(db, data)
    if not registration:
        raise HTTPException(status_code=400, detail=msg)
    return registration


@router.get("/{event_id}/registrations", response_model=List[EventRegistrationSchema])
def list_event_registrations(
    event_id: int,
    status: Optional[EventRegistrationStatus] = None,
    db: Session = Depends(get_db),
):
    event = db.query(Event).filter(Event.id == event_id).first()
    if not event:
        raise HTTPException(status_code=404, detail="活动不存在")
    
    query = db.query(EventRegistration).filter(EventRegistration.event_id == event_id)
    if status:
        query = query.filter(EventRegistration.status == status.value)
    
    return query.order_by(
        EventRegistration.status.asc(),
        EventRegistration.waitlist_position.asc().nullslast(),
        EventRegistration.registered_at.asc(),
    ).all()


@router.get("/reader/{reader_id}/registrations", response_model=List[EventRegistrationSchema])
def list_reader_registrations(
    reader_id: int,
    status: Optional[EventRegistrationStatus] = None,
    db: Session = Depends(get_db),
):
    query = db.query(EventRegistration).filter(EventRegistration.reader_id == reader_id)
    if status:
        query = query.filter(EventRegistration.status == status.value)
    return query.order_by(EventRegistration.registered_at.desc()).all()
