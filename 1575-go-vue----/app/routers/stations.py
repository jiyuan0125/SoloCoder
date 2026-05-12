from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime
from app.database import get_db, Station, Zone, SecurityGate
from app.schemas import (
    StationCreate, StationResponse, StationUpdate,
    ZoneCreate, ZoneResponse, ZoneUpdate,
    SecurityGateCreate, SecurityGateResponse
)

router = APIRouter()

@router.post("/stations", response_model=StationResponse, status_code=status.HTTP_201_CREATED)
def create_station(station: StationCreate, db: Session = Depends(get_db)):
    existing = db.query(Station).filter(Station.code == station.code).first()
    if existing:
        raise HTTPException(status_code=400, detail="车站代码已存在")
    
    db_station = Station(
        code=station.code,
        name=station.name,
        max_capacity=station.max_capacity
    )
    db.add(db_station)
    db.commit()
    db.refresh(db_station)
    return db_station

@router.get("/stations/{station_code}", response_model=StationResponse)
def get_station(station_code: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    return station

@router.get("/stations", response_model=List[StationResponse])
def list_stations(db: Session = Depends(get_db)):
    return db.query(Station).all()

@router.put("/stations/{station_code}", response_model=StationResponse)
def update_station(station_code: str, station_update: StationUpdate, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    update_data = station_update.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(station, key, value)
    
    db.commit()
    db.refresh(station)
    return station

@router.delete("/stations/{station_code}", status_code=status.HTTP_204_NO_CONTENT)
def delete_station(station_code: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    db.delete(station)
    db.commit()

@router.post("/stations/{station_code}/zones", response_model=ZoneResponse, status_code=status.HTTP_201_CREATED)
def create_zone(station_code: str, zone: ZoneCreate, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    db_zone = Zone(
        station_id=station.id,
        name=zone.name,
        zone_type=zone.zone_type,
        max_capacity=zone.max_capacity
    )
    db.add(db_zone)
    db.commit()
    db.refresh(db_zone)
    return db_zone

@router.get("/stations/{station_code}/zones", response_model=List[ZoneResponse])
def list_zones(station_code: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    return db.query(Zone).filter(Zone.station_id == station.id).all()

@router.get("/stations/{station_code}/zones/{zone_id}", response_model=ZoneResponse)
def get_zone(station_code: str, zone_id: int, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    zone = db.query(Zone).filter(Zone.id == zone_id, Zone.station_id == station.id).first()
    if not zone:
        raise HTTPException(status_code=404, detail="区域不存在")
    return zone

@router.put("/stations/{station_code}/zones/{zone_id}", response_model=ZoneResponse)
def update_zone(station_code: str, zone_id: int, zone_update: ZoneUpdate, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    zone = db.query(Zone).filter(Zone.id == zone_id, Zone.station_id == station.id).first()
    if not zone:
        raise HTTPException(status_code=404, detail="区域不存在")
    
    update_data = zone_update.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(zone, key, value)
    
    db.commit()
    db.refresh(zone)
    return zone

@router.delete("/stations/{station_code}/zones/{zone_id}", status_code=status.HTTP_204_NO_CONTENT)
def delete_zone(station_code: str, zone_id: int, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    zone = db.query(Zone).filter(Zone.id == zone_id, Zone.station_id == station.id).first()
    if not zone:
        raise HTTPException(status_code=404, detail="区域不存在")
    
    db.delete(zone)
    db.commit()

@router.post("/stations/{station_code}/zones/{zone_id}/gates", response_model=SecurityGateResponse, status_code=status.HTTP_201_CREATED)
def create_gate(station_code: str, zone_id: int, gate: SecurityGateCreate, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    zone = db.query(Zone).filter(Zone.id == zone_id, Zone.station_id == station.id).first()
    if not zone:
        raise HTTPException(status_code=404, detail="区域不存在")
    
    existing = db.query(SecurityGate).filter(
        SecurityGate.zone_id == zone_id,
        SecurityGate.gate_number == gate.gate_number
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="安检通道号已存在")
    
    db_gate = SecurityGate(
        zone_id=zone_id,
        gate_number=gate.gate_number
    )
    db.add(db_gate)
    db.commit()
    db.refresh(db_gate)
    return db_gate

@router.get("/stations/{station_code}/zones/{zone_id}/gates", response_model=List[SecurityGateResponse])
def list_gates(station_code: str, zone_id: int, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    zone = db.query(Zone).filter(Zone.id == zone_id, Zone.station_id == station.id).first()
    if not zone:
        raise HTTPException(status_code=404, detail="区域不存在")
    
    return db.query(SecurityGate).filter(SecurityGate.zone_id == zone_id).all()

@router.get("/stations/{station_code}/zones/{zone_id}/gates/{gate_id}", response_model=SecurityGateResponse)
def get_gate(station_code: str, zone_id: int, gate_id: int, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    zone = db.query(Zone).filter(Zone.id == zone_id, Zone.station_id == station.id).first()
    if not zone:
        raise HTTPException(status_code=404, detail="区域不存在")
    
    gate = db.query(SecurityGate).filter(
        SecurityGate.id == gate_id,
        SecurityGate.zone_id == zone_id
    ).first()
    if not gate:
        raise HTTPException(status_code=404, detail="安检通道不存在")
    return gate
