from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from app.database import get_db
from app.models import Route, RouteSegment, Station
from app.schemas import RouteCreate, RouteResponse

router = APIRouter(prefix="/routes", tags=["routes"])


@router.get("", response_model=List[RouteResponse])
def list_routes(db: Session = Depends(get_db)):
    return db.query(Route).all()


@router.post("", response_model=RouteResponse)
def create_route(route: RouteCreate, db: Session = Depends(get_db)):
    existing = db.query(Route).filter(Route.code == route.code).first()
    if existing:
        raise HTTPException(status_code=400, detail="Route code already exists")
    
    origin = db.query(Station).filter(Station.id == route.origin_station_id).first()
    if not origin:
        raise HTTPException(status_code=404, detail="Origin station not found")
    
    dest = db.query(Station).filter(Station.id == route.destination_station_id).first()
    if not dest:
        raise HTTPException(status_code=404, detail="Destination station not found")
    
    db_route = Route(
        code=route.code,
        name=route.name,
        origin_station_id=route.origin_station_id,
        destination_station_id=route.destination_station_id
    )
    db.add(db_route)
    db.flush()
    
    for seg in route.segments:
        route_seg = RouteSegment(
            route_id=db_route.id,
            sequence=seg.sequence,
            mode=seg.mode,
            origin_station_id=seg.origin_station_id,
            destination_station_id=seg.destination_station_id,
            estimated_hours=seg.estimated_hours,
            is_transit_point=seg.is_transit_point
        )
        db.add(route_seg)
    
    db.commit()
    db.refresh(db_route)
    return db_route


@router.get("/{route_code}", response_model=RouteResponse)
def get_route(route_code: str, db: Session = Depends(get_db)):
    route = db.query(Route).filter(Route.code == route_code).first()
    if not route:
        raise HTTPException(status_code=404, detail="Route not found")
    return route
