from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from .. import schemas, crud
from ..database import get_db

router = APIRouter(prefix="/stations", tags=["stations"])


@router.get("/", response_model=List[schemas.StationResponse])
def read_stations(db: Session = Depends(get_db)):
    return crud.get_stations_by_line(db, line_id=1)


@router.get("/{station_id}", response_model=schemas.StationResponse)
def read_station(station_id: int, db: Session = Depends(get_db)):
    db_station = crud.get_station(db, station_id=station_id)
    if db_station is None:
        raise HTTPException(status_code=404, detail="Station not found")
    return db_station


@router.post("/", response_model=schemas.StationResponse, status_code=status.HTTP_201_CREATED)
def create_station(station: schemas.StationCreate, db: Session = Depends(get_db)):
    db_line = crud.get_line(db, line_id=station.line_id)
    if db_line is None:
        raise HTTPException(status_code=404, detail="Line not found")
    return crud.create_station(db=db, station=station)


@router.put("/{station_id}", response_model=schemas.StationResponse)
def update_station(station_id: int, station: schemas.StationUpdate, db: Session = Depends(get_db)):
    db_station = crud.get_station(db, station_id=station_id)
    if db_station is None:
        raise HTTPException(status_code=404, detail="Station not found")
    return crud.update_station(db=db, station_id=station_id, station=station)
