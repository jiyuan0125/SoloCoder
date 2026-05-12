from datetime import datetime
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from ..database import get_db
from ..models import Park, Activity, ActivityStatus
from .. import schemas

router = APIRouter(prefix="/parks")


def get_park_by_code(db: Session, park_code: str):
    park = db.query(Park).filter(Park.park_code == park_code).first()
    if not park:
        raise HTTPException(status_code=404, detail=f"Park with code {park_code} not found")
    return park


def get_activity_by_code(db: Session, park_id: int, activity_code: str):
    activity = db.query(Activity).filter(
        Activity.park_id == park_id,
        Activity.activity_code == activity_code
    ).first()
    if not activity:
        raise HTTPException(
            status_code=404, 
            detail=f"Activity {activity_code} not found in park"
        )
    return activity


@router.post("/{park_code}/activities/apply", response_model=schemas.Activity, status_code=201)
def apply_activity(
    park_code: str,
    activity: schemas.ActivityCreate,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    
    existing = db.query(Activity).filter(Activity.activity_code == activity.activity_code).first()
    if existing:
        raise HTTPException(status_code=400, detail=f"Activity code {activity.activity_code} already exists")
    
    if activity.start_time >= activity.end_time:
        raise HTTPException(status_code=400, detail="End time must be after start time")
    
    requires_security = activity.area > 500
    
    activity_data = activity.model_dump(exclude={"park_code"})
    db_activity = Activity(
        park_id=park.id,
        requires_security=requires_security,
        status=ActivityStatus.PENDING,
        **activity_data
    )
    db.add(db_activity)
    db.commit()
    db.refresh(db_activity)
    return db_activity


@router.get("/{park_code}/activities/", response_model=List[schemas.Activity])
def list_activities(
    park_code: str,
    status: Optional[ActivityStatus] = Query(None),
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    
    query = db.query(Activity).filter(Activity.park_id == park.id)
    if status:
        query = query.filter(Activity.status == status)
    
    return query.order_by(Activity.start_time.desc()).all()


@router.get("/{park_code}/activities/{activity_code}", response_model=schemas.Activity)
def get_activity(
    park_code: str,
    activity_code: str,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    return get_activity_by_code(db, park.id, activity_code)


@router.post("/{park_code}/activities/{activity_code}/approve", response_model=schemas.Activity)
def approve_activity(
    park_code: str,
    activity_code: str,
    approval: schemas.ActivityApprove,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    activity = get_activity_by_code(db, park.id, activity_code)
    
    if activity.status != ActivityStatus.PENDING:
        raise HTTPException(
            status_code=400, 
            detail=f"Cannot approve activity with status {activity.status}"
        )
    
    if approval.approved:
        activity.status = ActivityStatus.APPROVED
        activity.approved_by = approval.approved_by
        activity.approved_at = datetime.utcnow()
        activity.rejection_reason = None
    else:
        activity.status = ActivityStatus.REJECTED
        activity.approved_by = approval.approved_by
        activity.rejection_reason = approval.rejection_reason or "活动申请被拒绝"
    
    db.commit()
    db.refresh(activity)
    return activity


@router.post("/{park_code}/activities/{activity_code}/record", response_model=schemas.Activity)
def record_activity_completion(
    park_code: str,
    activity_code: str,
    record: schemas.ActivityRecord,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    activity = get_activity_by_code(db, park.id, activity_code)
    
    if activity.status not in [ActivityStatus.APPROVED, ActivityStatus.COMPLETED]:
        raise HTTPException(
            status_code=400, 
            detail=f"Cannot record completion for activity with status {activity.status}"
        )
    
    activity.status = ActivityStatus.COMPLETED
    activity.completion_notes = record.completion_notes
    
    db.commit()
    db.refresh(activity)
    return activity


@router.put("/{park_code}/activities/{activity_code}", response_model=schemas.Activity)
def update_activity(
    park_code: str,
    activity_code: str,
    activity_update: schemas.ActivityCreate,
    db: Session = Depends(get_db)
):
    park = get_park_by_code(db, park_code)
    activity = get_activity_by_code(db, park.id, activity_code)
    
    if activity.status not in [ActivityStatus.PENDING, ActivityStatus.REJECTED]:
        raise HTTPException(
            status_code=400, 
            detail=f"Cannot update activity with status {activity.status}"
        )
    
    if activity_update.start_time >= activity_update.end_time:
        raise HTTPException(status_code=400, detail="End time must be after start time")
    
    activity.name = activity_update.name
    activity.organizer = activity_update.organizer
    activity.organizer_contact = activity_update.organizer_contact
    activity.area = activity_update.area
    activity.start_time = activity_update.start_time
    activity.end_time = activity_update.end_time
    activity.expected_participants = activity_update.expected_participants
    activity.requires_security = activity_update.area > 500
    
    if activity.status == ActivityStatus.REJECTED:
        activity.status = ActivityStatus.PENDING
        activity.rejection_reason = None
    
    db.commit()
    db.refresh(activity)
    return activity



