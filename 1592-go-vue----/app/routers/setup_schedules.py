from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date
from app import crud, schemas
from app.database import get_db

router = APIRouter()


@router.post("/setup-schedules/", response_model=schemas.SetupScheduleResponse)
def create_setup_schedule(schedule: schemas.SetupScheduleCreate,
                          db: Session = Depends(get_db)):
    try:
        return crud.create_setup_schedule(db=db, schedule=schedule)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/setup-schedules/", response_model=List[schemas.SetupScheduleResponse])
def read_setup_schedules(hall_id: Optional[int] = None,
                         setup_date: Optional[date] = None,
                         db: Session = Depends(get_db)):
    if hall_id:
        return crud.get_setup_schedules_by_hall(db, hall_id=hall_id, setup_date=setup_date)
    return []


@router.get("/setup-schedules/{schedule_id}", response_model=schemas.SetupScheduleResponse)
def read_setup_schedule(schedule_id: int, db: Session = Depends(get_db)):
    db_schedule = crud.get_setup_schedule(db, schedule_id=schedule_id)
    if db_schedule is None:
        raise HTTPException(status_code=404, detail="Setup schedule not found")
    return db_schedule
