from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.database import get_db
from app.crud import education as crud_education
from app.schemas.education import (
    EducationEventCreate,
    EducationEventUpdate,
    EducationEventOut,
    EducationParticipantCreate,
    EducationParticipantOut,
    AgeCheckResponse
)

router = APIRouter(prefix="/api/education", tags=["education"])


@router.post("/events", response_model=EducationEventOut, status_code=status.HTTP_201_CREATED)
def create_education_event(event: EducationEventCreate, db: Session = Depends(get_db)):
    return crud_education.create_education_event(db=db, event=event)


@router.get("/events", response_model=List[EducationEventOut])
def read_education_events(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return crud_education.get_education_events(db=db, skip=skip, limit=limit)


@router.get("/events/{event_id}", response_model=EducationEventOut)
def read_education_event(event_id: int, db: Session = Depends(get_db)):
    db_event = crud_education.get_education_event(db=db, event_id=event_id)
    if db_event is None:
        raise HTTPException(status_code=404, detail="教育活动不存在")
    return db_event


@router.put("/events/{event_id}", response_model=EducationEventOut)
def update_education_event(
    event_id: int,
    event_update: EducationEventUpdate,
    db: Session = Depends(get_db)
):
    db_event = crud_education.update_education_event(
        db=db,
        event_id=event_id,
        event_update=event_update
    )
    if db_event is None:
        raise HTTPException(status_code=404, detail="教育活动不存在")
    return db_event


@router.delete("/events/{event_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_education_event(event_id: int, db: Session = Depends(get_db)):
    if not crud_education.delete_education_event(db=db, event_id=event_id):
        raise HTTPException(status_code=404, detail="教育活动不存在")
    return None


@router.post("/participants", response_model=EducationParticipantOut, status_code=status.HTTP_201_CREATED)
def create_participant(participant: EducationParticipantCreate, db: Session = Depends(get_db)):
    result = crud_education.create_participant(db=db, participant=participant)
    if result["error"] or result["created"] is None:
        raise HTTPException(status_code=400, detail=result["error"] or "报名失败")
    return result["created"]


@router.get("/events/{event_id}/participants", response_model=List[EducationParticipantOut])
def read_participants_by_event(event_id: int, db: Session = Depends(get_db)):
    return crud_education.get_participants_by_event(db=db, event_id=event_id)


@router.get("/participants/{participant_id}", response_model=EducationParticipantOut)
def read_participant(participant_id: int, db: Session = Depends(get_db)):
    db_participant = crud_education.get_participant(db=db, participant_id=participant_id)
    if db_participant is None:
        raise HTTPException(status_code=404, detail="参与者不存在")
    return db_participant


@router.get("/participants/{participant_id}/age-check", response_model=AgeCheckResponse)
def check_participant_age(participant_id: int, db: Session = Depends(get_db)):
    result = crud_education.check_participant_age(db=db, participant_id=participant_id)
    if result is None:
        raise HTTPException(status_code=404, detail="参与者或活动不存在")
    return result


@router.delete("/participants/{participant_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_participant(participant_id: int, db: Session = Depends(get_db)):
    if not crud_education.delete_participant(db=db, participant_id=participant_id):
        raise HTTPException(status_code=404, detail="参与者不存在")
    return None
