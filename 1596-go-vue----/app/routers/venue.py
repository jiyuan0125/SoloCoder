from datetime import date
from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from app.database import get_db
from app.models import Venue, VenueType, Maintenance
from app.schemas import (
    VenueTypeCreate, VenueTypeResponse, VenueCreate, VenueResponse,
    MaintenanceCreate, MaintenanceResponse, VenueAvailability
)
from app.services.booking_service import get_available_slots

router = APIRouter(prefix="/api/venues", tags=["场地管理"])


@router.post("/types", response_model=VenueTypeResponse)
def create_venue_type(data: VenueTypeCreate, db: Session = Depends(get_db)):
    existing = db.query(VenueType).filter(VenueType.name == data.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="该场地类型已存在")
    
    venue_type = VenueType(**data.dict())
    db.add(venue_type)
    db.commit()
    db.refresh(venue_type)
    return venue_type


@router.get("/types", response_model=List[VenueTypeResponse])
def list_venue_types(db: Session = Depends(get_db)):
    return db.query(VenueType).all()


@router.post("/", response_model=VenueResponse)
def create_venue(data: VenueCreate, db: Session = Depends(get_db)):
    venue_type = db.query(VenueType).filter(VenueType.id == data.venue_type_id).first()
    if not venue_type:
        raise HTTPException(status_code=404, detail="场地类型不存在")
    
    venue = Venue(**data.dict())
    db.add(venue)
    db.commit()
    db.refresh(venue)
    return venue


@router.get("/", response_model=List[VenueResponse])
def list_venues(venue_type_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(Venue)
    if venue_type_id:
        query = query.filter(Venue.venue_type_id == venue_type_id)
    return query.all()


@router.get("/{venue_id}", response_model=VenueResponse)
def get_venue(venue_id: int, db: Session = Depends(get_db)):
    venue = db.query(Venue).filter(Venue.id == venue_id).first()
    if not venue:
        raise HTTPException(status_code=404, detail="场地不存在")
    return venue


@router.get("/{venue_id}/availability")
def get_venue_availability(venue_id: int, booking_date: date, db: Session = Depends(get_db)):
    venue = db.query(Venue).filter(Venue.id == venue_id).first()
    if not venue:
        raise HTTPException(status_code=404, detail="场地不存在")
    
    slots = get_available_slots(db, venue_id, booking_date)
    return {
        "venue_id": venue_id,
        "venue_name": venue.name,
        "booking_date": booking_date,
        "slots": slots,
    }


@router.post("/maintenance", response_model=MaintenanceResponse)
def create_maintenance(data: MaintenanceCreate, db: Session = Depends(get_db)):
    venue = db.query(Venue).filter(Venue.id == data.venue_id).first()
    if not venue:
        raise HTTPException(status_code=404, detail="场地不存在")
    
    if data.start_time >= data.end_time:
        raise HTTPException(status_code=400, detail="结束时间必须晚于开始时间")
    
    existing = db.query(Maintenance).filter(
        Maintenance.venue_id == data.venue_id,
        Maintenance.start_time < data.end_time,
        Maintenance.end_time > data.start_time,
    ).first()
    if existing:
        raise HTTPException(status_code=400, detail="该时段已有维修安排")
    
    maintenance = Maintenance(**data.dict())
    db.add(maintenance)
    db.commit()
    db.refresh(maintenance)
    return maintenance


@router.get("/maintenance", response_model=List[MaintenanceResponse])
def list_maintenance(venue_id: Optional[int] = None, db: Session = Depends(get_db)):
    query = db.query(Maintenance)
    if venue_id:
        query = query.filter(Maintenance.venue_id == venue_id)
    return query.order_by(Maintenance.start_time.desc()).all()
