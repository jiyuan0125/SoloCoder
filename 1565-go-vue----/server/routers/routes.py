from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List, Optional
from ..database import get_db
from ..models import Route, RouteWaypoint, Conflict, FlowRecord
from ..schemas import RouteCreate, RouteUpdate, RouteResponse, RouteStatus, ConflictResponse, FlowRecordResponse
from ..services.flow_manager import update_route_status
from ..services.conflict_detector import reevaluate_conflicts

router = APIRouter(prefix="/routes", tags=["routes"])


@router.get("/", response_model=List[RouteResponse])
def list_routes(db: Session = Depends(get_db)):
    return db.query(Route).all()


@router.post("/", response_model=RouteResponse)
def create_route(route: RouteCreate, db: Session = Depends(get_db)):
    existing = db.query(Route).filter(Route.code == route.code).first()
    if existing:
        raise HTTPException(status_code=400, detail=f"Route with code {route.code} already exists")
    
    route_data = route.model_dump(exclude={"waypoints"})
    db_route = Route(**route_data)
    db.add(db_route)
    db.flush()
    
    for wp in route.waypoints:
        rw = RouteWaypoint(route_id=db_route.id, **wp.model_dump())
        db.add(rw)
    
    db.commit()
    db.refresh(db_route)
    return db_route


@router.get("/{route_id}", response_model=RouteResponse)
def get_route(route_id: int, db: Session = Depends(get_db)):
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="Route not found")
    return route


@router.put("/{route_id}", response_model=RouteResponse)
def update_route(route_id: int, route_update: RouteUpdate, db: Session = Depends(get_db)):
    db_route = db.query(Route).filter(Route.id == route_id).first()
    if not db_route:
        raise HTTPException(status_code=404, detail="Route not found")
    
    update_data = route_update.model_dump(exclude_unset=True, exclude={"waypoints"})
    for key, value in update_data.items():
        setattr(db_route, key, value)
    
    if route_update.waypoints is not None:
        db.query(RouteWaypoint).filter(RouteWaypoint.route_id == route_id).delete()
        for wp in route_update.waypoints:
            rw = RouteWaypoint(route_id=route_id, **wp.model_dump())
            db.add(rw)
    
    db.commit()
    db.refresh(db_route)
    
    reevaluate_conflicts(db, db_route)
    
    return db_route


@router.delete("/{route_id}")
def delete_route(route_id: int, db: Session = Depends(get_db)):
    db_route = db.query(Route).filter(Route.id == route_id).first()
    if not db_route:
        raise HTTPException(status_code=404, detail="Route not found")
    
    db.delete(db_route)
    db.commit()
    return {"message": "Route deleted"}


@router.get("/{route_id}/status", response_model=RouteStatus)
def get_route_status(route_id: int, db: Session = Depends(get_db)):
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="Route not found")
    
    return update_route_status(db, route)


@router.get("/{route_id}/flow", response_model=List[FlowRecordResponse])
def get_route_flow(
    route_id: int,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="Route not found")
    
    records = db.query(FlowRecord).filter(
        FlowRecord.route_id == route_id
    ).order_by(FlowRecord.timestamp.desc()).limit(limit).all()
    
    return records


@router.get("/{route_id}/conflicts", response_model=List[ConflictResponse])
def get_route_conflicts(
    route_id: int,
    resolved: Optional[bool] = None,
    db: Session = Depends(get_db)
):
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="Route not found")
    
    query = db.query(Conflict).filter(Conflict.route_id == route_id)
    
    if resolved is not None:
        query = query.filter(Conflict.resolved == resolved)
    
    return query.order_by(Conflict.detected_at.desc()).all()


@router.post("/{route_id}/reevaluate", response_model=List[ConflictResponse])
def reevaluate_route_conflicts(route_id: int, db: Session = Depends(get_db)):
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="Route not found")
    
    return reevaluate_conflicts(db, route)
