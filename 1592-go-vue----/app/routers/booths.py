from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from app import crud, schemas
from app.database import get_db

router = APIRouter()


@router.post("/booths/", response_model=schemas.BoothResponse)
def create_booth(booth: schemas.BoothCreate, db: Session = Depends(get_db)):
    db_hall = crud.get_hall(db, hall_id=booth.hall_id)
    if db_hall is None:
        raise HTTPException(status_code=404, detail="Hall not found")
    return crud.create_booth(db=db, booth=booth)


@router.get("/booths/", response_model=List[schemas.BoothResponse])
def read_booths(hall_id: Optional[int] = None,
                exhibition_id: Optional[int] = None,
                db: Session = Depends(get_db)):
    if exhibition_id:
        return crud.get_available_booths(db, exhibition_id=exhibition_id, hall_id=hall_id)
    elif hall_id:
        return crud.get_booths_by_hall(db, hall_id=hall_id)
    return []


@router.get("/booths/{booth_id}", response_model=schemas.BoothResponse)
def read_booth(booth_id: int, db: Session = Depends(get_db)):
    db_booth = crud.get_booth(db, booth_id=booth_id)
    if db_booth is None:
        raise HTTPException(status_code=404, detail="Booth not found")
    return db_booth


@router.put("/booths/{booth_id}", response_model=schemas.BoothResponse)
def update_booth(booth_id: int, booth: schemas.BoothUpdate, db: Session = Depends(get_db)):
    db_booth = crud.update_booth(db, booth_id=booth_id, booth=booth)
    if db_booth is None:
        raise HTTPException(status_code=404, detail="Booth not found")
    return db_booth


@router.post("/booths/{booth_id}/update-price", response_model=schemas.BoothResponse)
def update_booth_price(booth_id: int, price_update: schemas.BoothPriceUpdate,
                       db: Session = Depends(get_db)):
    db_booth = crud.update_booth_price(db, booth_id=booth_id, new_price=price_update.new_price)
    if db_booth is None:
        raise HTTPException(status_code=404, detail="Booth not found")
    return db_booth


@router.delete("/booths/{booth_id}", response_model=schemas.BoothResponse)
def delete_booth(booth_id: int, db: Session = Depends(get_db)):
    db_booth = crud.delete_booth(db, booth_id=booth_id)
    if db_booth is None:
        raise HTTPException(status_code=404, detail="Booth not found")
    return db_booth
