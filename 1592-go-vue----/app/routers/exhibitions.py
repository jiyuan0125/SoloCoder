from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app import crud, schemas
from app.database import get_db

router = APIRouter()


@router.post("/exhibitions/", response_model=schemas.ExhibitionResponse)
def create_exhibition(exhibition: schemas.ExhibitionCreate, db: Session = Depends(get_db)):
    try:
        return crud.create_exhibition(db=db, exhibition=exhibition)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("/exhibitions/", response_model=List[schemas.ExhibitionResponse])
def read_exhibitions(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    exhibitions = crud.get_exhibitions(db, skip=skip, limit=limit)
    return exhibitions


@router.get("/exhibitions/{exhibition_id}", response_model=schemas.ExhibitionResponse)
def read_exhibition(exhibition_id: int, db: Session = Depends(get_db)):
    db_exhibition = crud.get_exhibition(db, exhibition_id=exhibition_id)
    if db_exhibition is None:
        raise HTTPException(status_code=404, detail="Exhibition not found")
    return db_exhibition


@router.put("/exhibitions/{exhibition_id}", response_model=schemas.ExhibitionResponse)
def update_exhibition(exhibition_id: int, exhibition: schemas.ExhibitionUpdate,
                      db: Session = Depends(get_db)):
    db_exhibition = crud.update_exhibition(db, exhibition_id=exhibition_id, exhibition=exhibition)
    if db_exhibition is None:
        raise HTTPException(status_code=404, detail="Exhibition not found")
    return db_exhibition


@router.delete("/exhibitions/{exhibition_id}", response_model=schemas.ExhibitionResponse)
def delete_exhibition(exhibition_id: int, db: Session = Depends(get_db)):
    db_exhibition = crud.delete_exhibition(db, exhibition_id=exhibition_id)
    if db_exhibition is None:
        raise HTTPException(status_code=404, detail="Exhibition not found")
    return db_exhibition
