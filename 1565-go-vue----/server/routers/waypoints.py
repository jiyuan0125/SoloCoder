from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from ..database import get_db
from ..models import Waypoint
from ..schemas import WaypointCreate, WaypointUpdate, WaypointResponse

router = APIRouter(prefix="/waypoints", tags=["waypoints"])


@router.get("/", response_model=List[WaypointResponse])
def list_waypoints(db: Session = Depends(get_db)):
    return db.query(Waypoint).all()


@router.post("/", response_model=WaypointResponse)
def create_waypoint(waypoint: WaypointCreate, db: Session = Depends(get_db)):
    existing = db.query(Waypoint).filter(Waypoint.code == waypoint.code).first()
    if existing:
        raise HTTPException(status_code=400, detail=f"Waypoint with code {waypoint.code} already exists")
    
    db_waypoint = Waypoint(**waypoint.model_dump())
    db.add(db_waypoint)
    db.commit()
    db.refresh(db_waypoint)
    return db_waypoint


@router.get("/{waypoint_id}", response_model=WaypointResponse)
def get_waypoint(waypoint_id: int, db: Session = Depends(get_db)):
    waypoint = db.query(Waypoint).filter(Waypoint.id == waypoint_id).first()
    if not waypoint:
        raise HTTPException(status_code=404, detail="Waypoint not found")
    return waypoint


@router.put("/{waypoint_id}", response_model=WaypointResponse)
def update_waypoint(waypoint_id: int, waypoint: WaypointUpdate, db: Session = Depends(get_db)):
    db_waypoint = db.query(Waypoint).filter(Waypoint.id == waypoint_id).first()
    if not db_waypoint:
        raise HTTPException(status_code=404, detail="Waypoint not found")
    
    update_data = waypoint.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_waypoint, key, value)
    
    db.commit()
    db.refresh(db_waypoint)
    return db_waypoint


@router.delete("/{waypoint_id}")
def delete_waypoint(waypoint_id: int, db: Session = Depends(get_db)):
    db_waypoint = db.query(Waypoint).filter(Waypoint.id == waypoint_id).first()
    if not db_waypoint:
        raise HTTPException(status_code=404, detail="Waypoint not found")
    
    db.delete(db_waypoint)
    db.commit()
    return {"message": "Waypoint deleted"}
