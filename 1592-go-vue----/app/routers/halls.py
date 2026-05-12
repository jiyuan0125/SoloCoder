from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from app import crud, schemas
from app.database import get_db

router = APIRouter()


@router.post("/halls/", response_model=schemas.HallResponse)
def create_hall(hall: schemas.HallCreate, db: Session = Depends(get_db)):
    db_venue = crud.get_venue(db, venue_id=hall.venue_id)
    if db_venue is None:
        raise HTTPException(status_code=404, detail="Venue not found")
    return crud.create_hall(db=db, hall=hall)


@router.get("/halls/", response_model=List[schemas.HallResponse])
def read_halls(venue_id: Optional[int] = None, db: Session = Depends(get_db)):
    if venue_id:
        return crud.get_halls_by_venue(db, venue_id=venue_id)
    return []


@router.get("/halls/{hall_id}", response_model=schemas.HallResponse)
def read_hall(hall_id: int, db: Session = Depends(get_db)):
    db_hall = crud.get_hall(db, hall_id=hall_id)
    if db_hall is None:
        raise HTTPException(status_code=404, detail="Hall not found")
    return db_hall


@router.put("/halls/{hall_id}", response_model=schemas.HallResponse)
def update_hall(hall_id: int, hall: schemas.HallUpdate, db: Session = Depends(get_db)):
    db_hall = crud.update_hall(db, hall_id=hall_id, hall=hall)
    if db_hall is None:
        raise HTTPException(status_code=404, detail="Hall not found")
    return db_hall


@router.delete("/halls/{hall_id}", response_model=schemas.HallResponse)
def delete_hall(hall_id: int, db: Session = Depends(get_db)):
    db_hall = crud.delete_hall(db, hall_id=hall_id)
    if db_hall is None:
        raise HTTPException(status_code=404, detail="Hall not found")
    return db_hall
