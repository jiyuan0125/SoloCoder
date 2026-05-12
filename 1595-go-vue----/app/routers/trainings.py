from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.database import get_db
from app.crud import training as crud_training
from app.schemas.training import (
    TrainingCreate,
    TrainingUpdate,
    TrainingOut,
    RegistrationCreate,
    RegistrationOut,
    AttendanceCreate,
    AttendanceOut,
    NotificationResponse
)

router = APIRouter(prefix="/api/trainings", tags=["trainings"])


@router.post("/", response_model=TrainingOut, status_code=status.HTTP_201_CREATED)
def create_training(training: TrainingCreate, db: Session = Depends(get_db)):
    return crud_training.create_training(db=db, training=training)


@router.get("/", response_model=List[TrainingOut])
def read_trainings(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return crud_training.get_trainings(db=db, skip=skip, limit=limit)


@router.get("/{training_id}", response_model=TrainingOut)
def read_training(training_id: int, db: Session = Depends(get_db)):
    db_training = crud_training.get_training(db=db, training_id=training_id)
    if db_training is None:
        raise HTTPException(status_code=404, detail="培训不存在")
    return db_training


@router.put("/{training_id}", response_model=TrainingOut)
def update_training(
    training_id: int,
    training_update: TrainingUpdate,
    db: Session = Depends(get_db)
):
    db_training = crud_training.update_training(
        db=db,
        training_id=training_id,
        training_update=training_update
    )
    if db_training is None:
        raise HTTPException(status_code=404, detail="培训不存在")
    return db_training


@router.delete("/{training_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_training(training_id: int, db: Session = Depends(get_db)):
    if not crud_training.delete_training(db=db, training_id=training_id):
        raise HTTPException(status_code=404, detail="培训不存在")
    return None


@router.post("/registrations", response_model=RegistrationOut, status_code=status.HTTP_201_CREATED)
def create_registration(registration: RegistrationCreate, db: Session = Depends(get_db)):
    db_registration = crud_training.create_registration(db=db, registration=registration)
    if db_registration is None:
        raise HTTPException(status_code=400, detail="报名失败，培训不存在")
    return db_registration


@router.get("/{training_id}/registrations", response_model=List[RegistrationOut])
def read_registrations_by_training(training_id: int, db: Session = Depends(get_db)):
    return crud_training.get_registrations_by_training(db=db, training_id=training_id)


@router.get("/registrations/{registration_id}", response_model=RegistrationOut)
def read_registration(registration_id: int, db: Session = Depends(get_db)):
    db_registration = crud_training.get_registration(db=db, registration_id=registration_id)
    if db_registration is None:
        raise HTTPException(status_code=404, detail="报名记录不存在")
    return db_registration


@router.post("/registrations/{registration_id}/cancel", response_model=RegistrationOut)
def cancel_registration(registration_id: int, db: Session = Depends(get_db)):
    db_registration = crud_training.cancel_registration(db=db, registration_id=registration_id)
    if db_registration is None:
        raise HTTPException(status_code=400, detail="取消报名失败，可能记录不存在或已取消")
    return db_registration


@router.post("/{training_id}/waitlist/process", response_model=NotificationResponse)
def process_waitlist(training_id: int, db: Session = Depends(get_db)):
    result = crud_training.process_waitlist_notifications(db=db, training_id=training_id)
    return result


@router.post("/attendances", response_model=AttendanceOut, status_code=status.HTTP_201_CREATED)
def create_attendance(attendance: AttendanceCreate, db: Session = Depends(get_db)):
    db_attendance = crud_training.create_attendance(db=db, attendance=attendance)
    if db_attendance is None:
        raise HTTPException(
            status_code=400,
            detail="签到失败，可能报名不存在或状态不正确"
        )
    return db_attendance


@router.get("/registrations/{registration_id}/attendances", response_model=List[AttendanceOut])
def read_attendances(registration_id: int, db: Session = Depends(get_db)):
    return crud_training.get_attendance(db=db, registration_id=registration_id)
