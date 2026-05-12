from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from .. import schemas, crud
from ..database import get_db

router = APIRouter(prefix="/lines", tags=["lines"])


@router.get("/", response_model=List[schemas.LineResponse])
def read_lines(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return crud.get_lines(db, skip=skip, limit=limit)


@router.get("/{line_id}", response_model=schemas.LineResponse)
def read_line(line_id: int, db: Session = Depends(get_db)):
    db_line = crud.get_line(db, line_id=line_id)
    if db_line is None:
        raise HTTPException(status_code=404, detail="Line not found")
    return db_line


@router.post("/", response_model=schemas.LineResponse, status_code=status.HTTP_201_CREATED)
def create_line(line: schemas.LineCreate, db: Session = Depends(get_db)):
    db_line = crud.get_line_by_name(db, name=line.name)
    if db_line:
        raise HTTPException(status_code=400, detail="Line with this name already exists")
    return crud.create_line(db=db, line=line)


@router.put("/{line_id}", response_model=schemas.LineResponse)
def update_line(line_id: int, line: schemas.LineUpdate, db: Session = Depends(get_db)):
    db_line = crud.get_line(db, line_id=line_id)
    if db_line is None:
        raise HTTPException(status_code=404, detail="Line not found")
    return crud.update_line(db=db, line_id=line_id, line=line)


@router.get("/{line_id}/stations", response_model=List[schemas.StationResponse])
def read_line_stations(line_id: int, db: Session = Depends(get_db)):
    db_line = crud.get_line(db, line_id=line_id)
    if db_line is None:
        raise HTTPException(status_code=404, detail="Line not found")
    return crud.get_stations_by_line(db, line_id=line_id)


@router.get("/{line_id}/power-sections", response_model=List[schemas.PowerSectionResponse])
def read_line_power_sections(line_id: int, db: Session = Depends(get_db)):
    db_line = crud.get_line(db, line_id=line_id)
    if db_line is None:
        raise HTTPException(status_code=404, detail="Line not found")
    return crud.get_power_sections_by_line(db, line_id=line_id)
