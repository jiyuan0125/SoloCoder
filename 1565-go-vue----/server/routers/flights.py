from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime
from ..database import get_db
from ..models import Flight, Route, FlightType
from ..schemas import FlightCreate, FlightUpdate, FlightResponse
from ..services.flow_manager import update_route_status, can_accept_flight, prioritize_waiting_queue
from ..services.conflict_detector import detect_conflicts

router = APIRouter(prefix="/flights", tags=["flights"])


@router.get("/", response_model=List[FlightResponse])
def list_flights(db: Session = Depends(get_db)):
    return db.query(Flight).all()


@router.post("/", response_model=FlightResponse)
def create_flight(flight: FlightCreate, db: Session = Depends(get_db)):
    existing = db.query(Flight).filter(Flight.flight_number == flight.flight_number).first()
    if existing:
        raise HTTPException(status_code=400, detail=f"Flight {flight.flight_number} already exists")
    
    route = db.query(Route).filter(Route.id == flight.route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail="Route not found")
    
    is_waiting = not can_accept_flight(db, route)
    
    db_flight = Flight(
        **flight.model_dump(),
        is_waiting=is_waiting
    )
    db.add(db_flight)
    db.flush()
    
    if not is_waiting:
        active_flights = db.query(Flight).filter(
            Flight.route_id == route.id,
            Flight.is_completed == False,
            Flight.is_waiting == False
        ).all()
        
        conflicts = detect_conflicts(db, route, active_flights)
        for conflict in conflicts:
            db.add(conflict)
    
    db.commit()
    db.refresh(db_flight)
    
    update_route_status(db, route, reason=f"航班 {flight.flight_number} 加入{'等待队列' if is_waiting else '航路'}")
    
    return db_flight


@router.get("/{flight_id}", response_model=FlightResponse)
def get_flight(flight_id: int, db: Session = Depends(get_db)):
    flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not flight:
        raise HTTPException(status_code=404, detail="Flight not found")
    return flight


@router.put("/{flight_id}", response_model=FlightResponse)
def update_flight(flight_id: int, flight_update: FlightUpdate, db: Session = Depends(get_db)):
    db_flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not db_flight:
        raise HTTPException(status_code=404, detail="Flight not found")
    
    update_data = flight_update.model_dump(exclude_unset=True)
    for key, value in update_data.items():
        setattr(db_flight, key, value)
    
    db.commit()
    db.refresh(db_flight)
    return db_flight


@router.delete("/{flight_id}")
def delete_flight(flight_id: int, db: Session = Depends(get_db)):
    db_flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not db_flight:
        raise HTTPException(status_code=404, detail="Flight not found")
    
    route_id = db_flight.route_id
    db.delete(db_flight)
    db.commit()
    
    route = db.query(Route).filter(Route.id == route_id).first()
    if route:
        waiting = prioritize_waiting_queue(db, route_id)
        if waiting and can_accept_flight(db, route):
            next_flight = waiting[0]
            next_flight.is_waiting = False
            db.commit()
        
        update_route_status(db, route, reason=f"航班 {db_flight.flight_number} 移除")
    
    return {"message": "Flight deleted"}


@router.post("/{flight_id}/complete", response_model=FlightResponse)
def complete_flight(flight_id: int, db: Session = Depends(get_db)):
    db_flight = db.query(Flight).filter(Flight.id == flight_id).first()
    if not db_flight:
        raise HTTPException(status_code=404, detail="Flight not found")
    
    db_flight.is_completed = True
    db_flight.exit_time = datetime.utcnow()
    db.commit()
    db.refresh(db_flight)
    
    route = db.query(Route).filter(Route.id == db_flight.route_id).first()
    if route:
        waiting = prioritize_waiting_queue(db, route.id)
        if waiting and can_accept_flight(db, route):
            next_flight = waiting[0]
            next_flight.is_waiting = False
            next_flight.actual_entry_time = datetime.utcnow()
            db.commit()
            
            active_flights = db.query(Flight).filter(
                Flight.route_id == route.id,
                Flight.is_completed == False,
                Flight.is_waiting == False
            ).all()
            
            from ..services.conflict_detector import detect_conflicts
            conflicts = detect_conflicts(db, route, active_flights)
            for conflict in conflicts:
                db.add(conflict)
            db.commit()
        
        update_route_status(db, route, reason=f"航班 {db_flight.flight_number} 完成")
    
    return db_flight
