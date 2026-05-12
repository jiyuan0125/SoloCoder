from datetime import datetime, timedelta
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Station, StationType, RedTideStatus
from ..schemas import (
    Station as StationSchema,
    StationCreate,
    StationData,
    TideResponse,
    WaveResponse,
    RedTideResponse
)
from ..services.tide_service import TideService
from ..services.wave_service import WaveService
from ..services.red_tide_service import RedTideService
from ..services.data_service import DataService

router = APIRouter(prefix="/stations", tags=["stations"])


@router.post("/", response_model=StationSchema, status_code=201)
def create_station(station: StationCreate, db: Session = Depends(get_db)):
    db_station = Station(**station.dict())
    db.add(db_station)
    db.commit()
    db.refresh(db_station)
    return db_station


@router.get("/", response_model=List[StationSchema])
def list_stations(
    skip: int = 0, 
    limit: int = 100, 
    station_type: Optional[StationType] = None,
    db: Session = Depends(get_db)
):
    query = db.query(Station)
    if station_type:
        query = query.filter(Station.station_type == station_type)
    return query.offset(skip).limit(limit).all()


@router.get("/{station_id}", response_model=StationSchema)
def get_station(station_id: int, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    return station


@router.get("/{station_id}/status")
def get_station_status(station_id: int, db: Session = Depends(get_db)):
    status = DataService.get_station_status(db, station_id)
    if not status:
        raise HTTPException(status_code=404, detail="监测站不存在")
    return status


@router.get("/{station_id}/tide", response_model=TideResponse)
def get_station_tide(
    station_id: int,
    date: Optional[str] = Query(None, description="日期格式: YYYY-MM-DD"),
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    target_date = None
    if date:
        try:
            target_date = datetime.strptime(date, "%Y-%m-%d")
        except ValueError:
            raise HTTPException(status_code=400, detail="日期格式无效，请使用 YYYY-MM-DD")
    
    return TideService.get_tide_response(db, station_id, target_date)


@router.get("/{station_id}/wave", response_model=WaveResponse)
def get_station_wave(
    station_id: int,
    limit: int = Query(24, ge=1, le=1000),
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    return WaveService.get_wave_response(db, station_id, limit)


@router.get("/{station_id}/redtide", response_model=RedTideResponse)
def get_station_redtide(
    station_id: int,
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    if station.station_type != StationType.BUOY:
        raise HTTPException(status_code=400, detail="岸基站不支持赤潮监测")
    
    return RedTideService.get_red_tide_response(db, station_id)
