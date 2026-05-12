from datetime import datetime
from typing import List, Optional
from sqlalchemy.orm import Session

from app.models.education import EducationEvent, EducationParticipant
from app.schemas.education import EducationEventCreate, EducationEventUpdate, EducationParticipantCreate
from app.utils.age_calculator import validate_age_range


def get_education_event(db: Session, event_id: int) -> Optional[EducationEvent]:
    return db.query(EducationEvent).filter(EducationEvent.id == event_id).first()


def get_education_events(db: Session, skip: int = 0, limit: int = 100) -> List[EducationEvent]:
    return db.query(EducationEvent).offset(skip).limit(limit).all()


def create_education_event(db: Session, event: EducationEventCreate) -> EducationEvent:
    db_event = EducationEvent(**event.model_dump())
    db.add(db_event)
    db.commit()
    db.refresh(db_event)
    return db_event


def update_education_event(db: Session, event_id: int, event_update: EducationEventUpdate) -> Optional[EducationEvent]:
    db_event = get_education_event(db, event_id)
    if not db_event:
        return None
    
    for key, value in event_update.model_dump(exclude_unset=True).items():
        setattr(db_event, key, value)
    
    db.commit()
    db.refresh(db_event)
    return db_event


def delete_education_event(db: Session, event_id: int) -> bool:
    db_event = get_education_event(db, event_id)
    if not db_event:
        return False
    
    db.delete(db_event)
    db.commit()
    return True


def get_participants_count(db: Session, event_id: int) -> int:
    return db.query(EducationParticipant).filter(EducationParticipant.event_id == event_id).count()


def check_participant_age(
    db: Session,
    participant_id: int
) -> Optional[dict]:
    participant = db.query(EducationParticipant).filter(
        EducationParticipant.id == participant_id
    ).first()
    
    if not participant:
        return None
    
    event = get_education_event(db, participant.event_id)
    if not event:
        return None
    
    age_check = validate_age_range(
        participant.birth_date,
        event.min_age_months,
        event.max_age_months,
        event.start_time
    )
    
    return {
        "participant_id": participant_id,
        "participant_name": participant.participant_name,
        "age_months": age_check["age_months"],
        "is_eligible": age_check["is_eligible"],
        "min_age_months": event.min_age_months,
        "max_age_months": event.max_age_months,
        "message": age_check["message"]
    }


def create_participant(db: Session, participant: EducationParticipantCreate) -> Optional[dict]:
    event = get_education_event(db, participant.event_id)
    if not event:
        return {"error": "活动不存在", "created": None}
    
    current_count = get_participants_count(db, participant.event_id)
    if current_count >= event.max_participants:
        return {"error": "名额已满", "created": None}
    
    age_check = validate_age_range(
        participant.birth_date,
        event.min_age_months,
        event.max_age_months,
        event.start_time
    )
    
    if not age_check["is_eligible"]:
        return {"error": age_check["message"], "created": None}
    
    db_participant = EducationParticipant(**participant.model_dump())
    db.add(db_participant)
    db.commit()
    db.refresh(db_participant)
    
    return {
        "error": None,
        "created": db_participant,
        "age_check": age_check
    }


def get_participant(db: Session, participant_id: int) -> Optional[EducationParticipant]:
    return db.query(EducationParticipant).filter(EducationParticipant.id == participant_id).first()


def get_participants_by_event(db: Session, event_id: int) -> List[EducationParticipant]:
    return db.query(EducationParticipant).filter(EducationParticipant.event_id == event_id).all()


def delete_participant(db: Session, participant_id: int) -> bool:
    db_participant = get_participant(db, participant_id)
    if not db_participant:
        return False
    
    db.delete(db_participant)
    db.commit()
    return True
