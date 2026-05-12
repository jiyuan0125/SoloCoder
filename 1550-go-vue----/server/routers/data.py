from datetime import datetime
from typing import Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Station, StationType
from ..schemas import (
    TideReading as TideReadingSchema,
    TideReadingCreate,
    WaveReading as WaveReadingSchema,
    WaveReadingCreate,
    WaterTempReading as WaterTempSchema,
    WaterTempReadingCreate,
    SalinityReading as SalinitySchema,
    SalinityReadingCreate,
    DissolvedOxygenReading as DOSchema,
    DissolvedOxygenReadingCreate,
    ChlorophyllReading as ChloroSchema,
    ChlorophyllReadingCreate,
)
from ..services.tide_service import TideService
from ..services.wave_service import WaveService
from ..services.red_tide_service import RedTideService
from ..services.data_service import DataService

router = APIRouter(prefix="/data", tags=["data"])


@router.post("/tide", response_model=TideReadingSchema, status_code=201)
def create_tide_reading(
    reading: TideReadingCreate,
    station_id: int,
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    return TideService.create_reading(
        db=db,
        station_id=station_id,
        value=reading.value,
        reading_time=reading.reading_time
    )


@router.post("/wave", response_model=WaveReadingSchema, status_code=201)
def create_wave_reading(
    reading: WaveReadingCreate,
    station_id: int,
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    return WaveService.create_reading(
        db=db,
        station_id=station_id,
        significant_wave_height=reading.significant_wave_height,
        reading_time=reading.reading_time
    )


@router.post("/water-temp", response_model=WaterTempSchema, status_code=201)
def create_water_temp_reading(
    reading: WaterTempReadingCreate,
    station_id: int,
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    return DataService.create_water_temp_reading(
        db=db,
        station_id=station_id,
        value=reading.value,
        reading_time=reading.reading_time
    )


@router.post("/salinity", response_model=SalinitySchema, status_code=201)
def create_salinity_reading(
    reading: SalinityReadingCreate,
    station_id: int,
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    if station.station_type != StationType.BUOY:
        raise HTTPException(status_code=400, detail="岸基站不支持盐度监测")
    
    return DataService.create_salinity_reading(
        db=db,
        station_id=station_id,
        value=reading.value,
        reading_time=reading.reading_time
    )


@router.post("/dissolved-oxygen", response_model=DOSchema, status_code=201)
def create_do_reading(
    reading: DissolvedOxygenReadingCreate,
    station_id: int,
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    if station.station_type != StationType.BUOY:
        raise HTTPException(status_code=400, detail="岸基站不支持溶解氧监测")
    
    return DataService.create_do_reading(
        db=db,
        station_id=station_id,
        value=reading.value,
        reading_time=reading.reading_time
    )


@router.post("/chlorophyll", response_model=ChloroSchema, status_code=201)
def create_chlorophyll_reading(
    reading: ChlorophyllReadingCreate,
    station_id: int,
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.id == station_id).first()
    if not station:
        raise HTTPException(status_code=404, detail="监测站不存在")
    
    if station.station_type != StationType.BUOY:
        raise HTTPException(status_code=400, detail="岸基站不支持叶绿素监测")
    
    return RedTideService.create_chlorophyll_reading(
        db=db,
        station_id=station_id,
        value=reading.value,
        reading_time=reading.reading_time
    )
