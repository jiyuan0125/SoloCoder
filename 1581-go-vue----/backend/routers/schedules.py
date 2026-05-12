from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from datetime import datetime
from .. import schemas, crud, services
from ..database import get_db
from ..config import settings

router = APIRouter(prefix="/schedules", tags=["schedules"])


@router.get("/", response_model=List[schemas.ScheduleResponse])
def read_schedules(line_id: Optional[int] = None, active_only: bool = False, db: Session = Depends(get_db)):
    if line_id:
        return crud.get_schedules_by_line(db, line_id=line_id, active_only=active_only)
    return []


@router.get("/{schedule_id}", response_model=schemas.ScheduleResponse)
def read_schedule(schedule_id: int, db: Session = Depends(get_db)):
    db_schedule = crud.get_schedule(db, schedule_id=schedule_id)
    if db_schedule is None:
        raise HTTPException(status_code=404, detail="Schedule not found")
    return db_schedule


@router.post("/", response_model=schemas.ScheduleResponse, status_code=status.HTTP_201_CREATED)
def create_schedule(schedule: schemas.ScheduleCreate, db: Session = Depends(get_db)):
    db_line = crud.get_line(db, line_id=schedule.line_id)
    if db_line is None:
        raise HTTPException(status_code=404, detail="Line not found")

    db_schedule = crud.create_schedule(db=db, schedule=schedule)

    if schedule.auto_generate_trains:
        dispatcher = services.get_dispatcher_service(db)
        dispatcher.generate_schedule_trains(db_schedule)
        db.refresh(db_schedule)

    return db_schedule


@router.put("/{schedule_id}", response_model=schemas.ScheduleResponse)
def update_schedule(
    schedule_id: int,
    schedule: schemas.ScheduleUpdate,
    db: Session = Depends(get_db),
):
    db_schedule = crud.get_schedule(db, schedule_id=schedule_id)
    if db_schedule is None:
        raise HTTPException(status_code=404, detail="Schedule not found")

    old_interval = db_schedule.interval_seconds
    updated = crud.update_schedule(db=db, schedule_id=schedule_id, schedule=schedule)

    if schedule.interval_seconds and schedule.interval_seconds != old_interval:
        dispatcher = services.get_dispatcher_service(db)
        dispatcher.reschedule_upcoming_trains(schedule_id, schedule.interval_seconds)

    return updated


@router.get("/{schedule_id}/trains", response_model=List[schemas.ScheduleTrainResponse])
def read_schedule_trains(schedule_id: int, db: Session = Depends(get_db)):
    db_schedule = crud.get_schedule(db, schedule_id=schedule_id)
    if db_schedule is None:
        raise HTTPException(status_code=404, detail="Schedule not found")
    return crud.get_schedule_trains_by_schedule(db, schedule_id=schedule_id)


@router.post("/{schedule_id}/optimal-interval")
def calculate_optimal_interval(schedule_id: int, db: Session = Depends(get_db)):
    db_schedule = crud.get_schedule(db, schedule_id=schedule_id)
    if db_schedule is None:
        raise HTTPException(status_code=404, detail="Schedule not found")

    dispatcher = services.get_dispatcher_service(db)
    optimal = dispatcher.calculate_optimal_interval(db_schedule.line_id)

    return {
        "schedule_id": schedule_id,
        "optimal_interval_seconds": optimal,
        "current_interval_seconds": db_schedule.interval_seconds,
        "min_safe_interval_seconds": settings.MIN_SAFE_INTERVAL,
    }


@router.get("/trains/{st_id}", response_model=schemas.ScheduleTrainResponse)
def read_schedule_train(st_id: int, db: Session = Depends(get_db)):
    db_st = crud.get_schedule_train(db, st_id=st_id)
    if db_st is None:
        raise HTTPException(status_code=404, detail="Schedule train not found")
    return db_st


@router.put("/trains/{st_id}/departure")
def update_train_departure(
    st_id: int,
    new_departure: datetime = Query(..., description="New scheduled departure time"),
    db: Session = Depends(get_db),
):
    dispatcher = services.get_dispatcher_service(db)
    try:
        st, conflicts = dispatcher.update_schedule_train_departure(st_id, new_departure)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))

    return {
        "schedule_train": schemas.ScheduleTrainResponse.model_validate(st),
        "conflicts": conflicts,
    }


@router.post("/trains/{st_id}/complete", response_model=schemas.ScheduleTrainResponse)
def complete_schedule_train(
    st_id: int,
    actual_departure: datetime = Query(...),
    actual_arrival: datetime = Query(...),
    db: Session = Depends(get_db),
):
    dispatcher = services.get_dispatcher_service(db)
    try:
        st = dispatcher.complete_schedule_train(st_id, actual_departure, actual_arrival)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    return st
