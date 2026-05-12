from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models import Station
from app.schemas import StationCreate, StationResponse

router = APIRouter(prefix="/stations", tags=["stations"])


@router.get("", response_model=List[StationResponse])
def list_stations(db: Session = Depends(get_db)):
    return db.query(Station).all()


@router.post("", response_model=StationResponse)
def create_station(station: StationCreate, db: Session = Depends(get_db)):
    existing = db.query(Station).filter(Station.code == station.code).first()
    if existing:
        raise HTTPException(status_code=400, detail="Station code already exists")
    
    db_station = Station(**station.model_dump())
    db.add(db_station)
    db.commit()
    db.refresh(db_station)
    return db_station


@router.get("/{station_code}", response_model=StationResponse)
def get_station(station_code: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="Station not found")
    return station
