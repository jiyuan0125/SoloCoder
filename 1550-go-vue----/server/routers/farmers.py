from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Farmer, Notification
from ..schemas import (
    Farmer as FarmerSchema,
    FarmerCreate,
    Notification as NotificationSchema
)
from ..services.notification_service import NotificationService

router = APIRouter(prefix="/farmers", tags=["farmers"])


@router.post("/", response_model=FarmerSchema, status_code=201)
def create_farmer(farmer: FarmerCreate, db: Session = Depends(get_db)):
    station_ids = []
    if farmer.station_ids:
        try:
            import json
            station_ids = json.loads(farmer.station_ids)
        except:
            try:
                station_ids = [int(x.strip()) for x in farmer.station_ids.split(",")]
            except:
                station_ids = []
    
    return NotificationService.create_farmer(
        db=db,
        name=farmer.name,
        station_ids=station_ids,
        email=farmer.email,
        phone=farmer.phone
    )


@router.get("/", response_model=List[FarmerSchema])
def list_farmers(
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    return db.query(Farmer).offset(skip).limit(limit).all()


@router.get("/{farmer_id}", response_model=FarmerSchema)
def get_farmer(farmer_id: int, db: Session = Depends(get_db)):
    farmer = db.query(Farmer).filter(Farmer.id == farmer_id).first()
    if not farmer:
        raise HTTPException(status_code=404, detail="养殖户不存在")
    return farmer


@router.get("/{farmer_id}/notifications", response_model=List[NotificationSchema])
def get_farmer_notifications(
    farmer_id: int,
    limit: int = 50,
    db: Session = Depends(get_db)
):
    farmer = db.query(Farmer).filter(Farmer.id == farmer_id).first()
    if not farmer:
        raise HTTPException(status_code=404, detail="养殖户不存在")
    return NotificationService.get_notifications_for_farmer(db, farmer_id, limit)


@router.put("/notifications/{notification_id}/read", response_model=NotificationSchema)
def mark_notification_read(notification_id: int, db: Session = Depends(get_db)):
    notification = NotificationService.mark_notification_read(db, notification_id)
    if not notification:
        raise HTTPException(status_code=404, detail="通知不存在")
    return notification
