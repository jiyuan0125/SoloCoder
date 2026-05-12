from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import date
from app import crud, schemas
from app.database import get_db

router = APIRouter()


@router.post("/selections/", response_model=schemas.BoothSelectionResponse)
def create_booth_selection(selection: schemas.BoothSelectionCreate,
                           db: Session = Depends(get_db)):
    try:
        return crud.create_booth_selection(db=db, selection=selection)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/selections/", response_model=List[schemas.BoothSelectionResponse])
def read_selections(exhibition_id: Optional[int] = None, db: Session = Depends(get_db)):
    if exhibition_id:
        return crud.get_booth_selections_by_exhibition(db, exhibition_id=exhibition_id)
    return []


@router.get("/selections/{selection_id}", response_model=schemas.BoothSelectionResponse)
def read_selection(selection_id: int, db: Session = Depends(get_db)):
    db_selection = crud.get_booth_selection(db, selection_id=selection_id)
    if db_selection is None:
        raise HTTPException(status_code=404, detail="Selection not found")
    return db_selection


@router.post("/selections/{selection_id}/confirm", response_model=schemas.BoothSelectionResponse)
def confirm_selection(selection_id: int, confirm: schemas.BoothSelectionConfirm,
                      db: Session = Depends(get_db)):
    try:
        db_selection = crud.confirm_booth_selection(db, selection_id=selection_id)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    if db_selection is None:
        raise HTTPException(status_code=404, detail="Selection not found")
    return db_selection
