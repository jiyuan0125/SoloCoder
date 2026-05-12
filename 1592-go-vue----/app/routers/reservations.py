from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date
from app import crud, schemas
from app.database import get_db

router = APIRouter()


@router.post("/daily-capacities/", response_model=schemas.DailyCapacityResponse)
def create_daily_capacity(capacity: schemas.DailyCapacityCreate,
                          db: Session = Depends(get_db)):
    db_exhibition = crud.get_exhibition(db, exhibition_id=capacity.exhibition_id)
    if db_exhibition is None:
        raise HTTPException(status_code=404, detail="Exhibition not found")
    return crud.create_daily_capacity(db=db, capacity=capacity)


@router.get("/daily-capacities/", response_model=Optional[schemas.DailyCapacityResponse])
def read_daily_capacity(exhibition_id: int, visit_date: date, db: Session = Depends(get_db)):
    return crud.get_daily_capacity(db, exhibition_id=exhibition_id, visit_date=visit_date)


@router.post("/reservations/", response_model=schemas.VisitReservationResponse)
def create_reservation(reservation: schemas.VisitReservationCreate,
                       db: Session = Depends(get_db)):
    try:
        return crud.create_visit_reservation(db=db, reservation=reservation)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/reservations/", response_model=List[schemas.VisitReservationResponse])
def read_reservations(exhibition_id: Optional[int] = None,
                      visit_date: Optional[date] = None,
                      db: Session = Depends(get_db)):
    if exhibition_id:
        return crud.get_reservations_by_exhibition(
            db, exhibition_id=exhibition_id, visit_date=visit_date
        )
    return []


@router.get("/reservations/{reservation_id}", response_model=schemas.VisitReservationResponse)
def read_reservation(reservation_id: int, db: Session = Depends(get_db)):
    db_reservation = crud.get_visit_reservation(db, reservation_id=reservation_id)
    if db_reservation is None:
        raise HTTPException(status_code=404, detail="Reservation not found")
    return db_reservation
