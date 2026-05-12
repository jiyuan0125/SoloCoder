from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from app.database import get_db
from app.crud import venue as crud_venue
from app.schemas.venue import (
    VenueCreate,
    VenueUpdate,
    VenueOut,
    RentalCreate,
    RentalOut
)

router = APIRouter(prefix="/api/venues", tags=["venues"])


@router.post("/", response_model=VenueOut, status_code=status.HTTP_201_CREATED)
def create_venue(venue: VenueCreate, db: Session = Depends(get_db)):
    return crud_venue.create_venue(db=db, venue=venue)


@router.get("/", response_model=List[VenueOut])
def read_venues(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return crud_venue.get_venues(db=db, skip=skip, limit=limit)


@router.get("/{venue_id}", response_model=VenueOut)
def read_venue(venue_id: int, db: Session = Depends(get_db)):
    db_venue = crud_venue.get_venue(db=db, venue_id=venue_id)
    if db_venue is None:
        raise HTTPException(status_code=404, detail="场地不存在")
    return db_venue


@router.put("/{venue_id}", response_model=VenueOut)
def update_venue(
    venue_id: int,
    venue_update: VenueUpdate,
    db: Session = Depends(get_db)
):
    db_venue = crud_venue.update_venue(
        db=db,
        venue_id=venue_id,
        venue_update=venue_update
    )
    if db_venue is None:
        raise HTTPException(status_code=404, detail="场地不存在")
    return db_venue


@router.delete("/{venue_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_venue(venue_id: int, db: Session = Depends(get_db)):
    if not crud_venue.delete_venue(db=db, venue_id=venue_id):
        raise HTTPException(status_code=404, detail="场地不存在")
    return None


@router.post("/rentals", response_model=RentalOut, status_code=status.HTTP_201_CREATED)
def create_rental(rental: RentalCreate, db: Session = Depends(get_db)):
    db_rental = crud_venue.create_rental(db=db, rental=rental)
    if db_rental is None:
        raise HTTPException(status_code=400, detail="租赁失败，场地不存在")
    return db_rental


@router.get("/rentals/", response_model=List[RentalOut])
def read_rentals(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return crud_venue.get_rentals(db=db, skip=skip, limit=limit)


@router.get("/rentals/{rental_id}", response_model=RentalOut)
def read_rental(rental_id: int, db: Session = Depends(get_db)):
    db_rental = crud_venue.get_rental(db=db, rental_id=rental_id)
    if db_rental is None:
        raise HTTPException(status_code=404, detail="租赁记录不存在")
    return db_rental


@router.get("/{venue_id}/rentals", response_model=List[RentalOut])
def read_rentals_by_venue(venue_id: int, db: Session = Depends(get_db)):
    return crud_venue.get_rentals_by_venue(db=db, venue_id=venue_id)
