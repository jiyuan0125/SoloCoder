from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List

from ..database import get_db
from ..models import Flight, Compartment
from ..schemas import FlightCreate, FlightResponse, CompartmentCreate, CompartmentResponse

router = APIRouter(prefix="/flights", tags=["flights"])


@router.post("", response_model=FlightResponse)
def create_flight(flight: FlightCreate, db: Session = Depends(get_db)):
    db_flight = db.query(Flight).filter(Flight.flight_number == flight.flight_number).first()
    if db_flight:
        raise HTTPException(status_code=400, detail="航班号已存在")
    
    db_flight = Flight(
        flight_number=flight.flight_number,
        route=flight.route,
        scheduled_departure=flight.scheduled_departure,
        scheduled_arrival=flight.scheduled_arrival,
        unit_price=flight.unit_price,
        min_load_rate=flight.min_load_rate
    )
    db.add(db_flight)
    db.commit()
    db.refresh(db_flight)
    
    return db_flight


@router.get("/{flight_number}", response_model=FlightResponse)
def get_flight(flight_number: str, db: Session = Depends(get_db)):
    db_flight = db.query(Flight).filter(Flight.flight_number == flight_number).first()
    if not db_flight:
        raise HTTPException(status_code=404, detail="航班不存在")
    return db_flight


@router.get("", response_model=List[FlightResponse])
def list_flights(db: Session = Depends(get_db)):
    return db.query(Flight).all()


@router.post("/compartments", response_model=CompartmentResponse)
def create_compartment(compartment: CompartmentCreate, db: Session = Depends(get_db)):
    db_flight = db.query(Flight).filter(Flight.flight_number == compartment.flight_number).first()
    if not db_flight:
        raise HTTPException(status_code=404, detail="航班不存在")
    
    db_compartment = db.query(Compartment).filter(
        Compartment.compartment_number == compartment.compartment_number
    ).first()
    if db_compartment:
        raise HTTPException(status_code=400, detail="舱位号已存在")
    
    db_compartment = Compartment(
        compartment_number=compartment.compartment_number,
        flight_id=db_flight.id,
        position=compartment.position,
        total_capacity_weight=compartment.total_capacity_weight,
        remaining_capacity_weight=compartment.total_capacity_weight,
        is_temperature_controlled=compartment.is_temperature_controlled
    )
    db.add(db_compartment)
    db.commit()
    db.refresh(db_compartment)
    
    return db_compartment


@router.get("/compartments/{compartment_number}", response_model=CompartmentResponse)
def get_compartment(compartment_number: str, db: Session = Depends(get_db)):
    db_compartment = db.query(Compartment).filter(
        Compartment.compartment_number == compartment_number
    ).first()
    if not db_compartment:
        raise HTTPException(status_code=404, detail="舱位不存在")
    return db_compartment


@router.get("/{flight_number}/compartments", response_model=List[CompartmentResponse])
def list_flight_compartments(flight_number: str, db: Session = Depends(get_db)):
    db_flight = db.query(Flight).filter(Flight.flight_number == flight_number).first()
    if not db_flight:
        raise HTTPException(status_code=404, detail="航班不存在")
    
    return db.query(Compartment).filter(Compartment.flight_id == db_flight.id).all()
