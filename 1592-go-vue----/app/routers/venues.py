from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app import crud, schemas
from app.database import get_db

router = APIRouter()


@router.post("/venues/", response_model=schemas.VenueResponse)
def create_venue(venue: schemas.VenueCreate, db: Session = Depends(get_db)):
    return crud.create_venue(db=db, venue=venue)


@router.get("/venues/", response_model=List[schemas.VenueResponse])
def read_venues(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    venues = crud.get_venues(db, skip=skip, limit=limit)
    return venues


@router.get("/venues/{venue_id}", response_model=schemas.VenueResponse)
def read_venue(venue_id: int, db: Session = Depends(get_db)):
    db_venue = crud.get_venue(db, venue_id=venue_id)
    if db_venue is None:
        raise HTTPException(status_code=404, detail="Venue not found")
    return db_venue


@router.put("/venues/{venue_id}", response_model=schemas.VenueResponse)
def update_venue(venue_id: int, venue: schemas.VenueUpdate, db: Session = Depends(get_db)):
    db_venue = crud.update_venue(db, venue_id=venue_id, venue=venue)
    if db_venue is None:
        raise HTTPException(status_code=404, detail="Venue not found")
    return db_venue


@router.delete("/venues/{venue_id}", response_model=schemas.VenueResponse)
def delete_venue(venue_id: int, db: Session = Depends(get_db)):
    db_venue = crud.delete_venue(db, venue_id=venue_id)
    if db_venue is None:
        raise HTTPException(status_code=404, detail="Venue not found")
    return db_venue
