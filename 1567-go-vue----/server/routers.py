from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from . import schemas, services
from .database import get_db

router = APIRouter()


@router.post("/flights", response_model=schemas.FlightResponse, status_code=status.HTTP_201_CREATED)
def create_flight(flight: schemas.FlightCreate, db: Session = Depends(get_db)):
    from .models import Flight
    existing = db.query(Flight).filter(Flight.id == flight.id).first()
    if existing:
        raise HTTPException(status_code=400, detail="航班ID已存在")
    existing_number = db.query(Flight).filter(Flight.flight_number == flight.flight_number).first()
    if existing_number:
        raise HTTPException(status_code=400, detail="航班号已存在")
    return services.create_flight(db, flight)


@router.get("/flights", response_model=List[schemas.FlightResponse])
def list_flights(db: Session = Depends(get_db)):
    return services.get_all_flights(db)


@router.get("/flights/{flight_id}", response_model=schemas.FlightResponse)
def get_flight(flight_id: str, db: Session = Depends(get_db)):
    flight = services.get_flight(db, flight_id)
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")
    return flight


@router.patch("/flights/{flight_id}", response_model=schemas.FlightResponse)
def update_flight(flight_id: str, update: schemas.FlightUpdate, db: Session = Depends(get_db)):
    flight = services.update_flight(db, flight_id, update)
    if not flight:
        raise HTTPException(status_code=404, detail="航班不存在")
    return flight


@router.get("/flights/{flight_id}/baggage/baggage-list", response_model=schemas.FlightBaggageListResponse)
def get_flight_baggage_list(flight_id: str, db: Session = Depends(get_db)):
    result = services.get_flight_baggage_list(db, flight_id)
    if not result:
        raise HTTPException(status_code=404, detail="航班不存在")
    return result


@router.get("/flights/{flight_id}/baggage/sorting", response_model=schemas.FlightSortingResponse)
def get_flight_sorting(flight_id: str, db: Session = Depends(get_db)):
    result = services.get_flight_sorting(db, flight_id)
    if not result:
        raise HTTPException(status_code=404, detail="航班不存在")
    return result


@router.get("/flights/{flight_id}/baggage/loading", response_model=schemas.FlightLoadingResponse)
def get_flight_loading(flight_id: str, db: Session = Depends(get_db)):
    result = services.get_flight_loading(db, flight_id)
    if not result:
        raise HTTPException(status_code=404, detail="航班不存在")
    return result


@router.post("/baggage", response_model=schemas.BaggageResponse, status_code=status.HTTP_201_CREATED)
def create_baggage(baggage: schemas.BaggageCreate, db: Session = Depends(get_db)):
    from .models import Baggage
    existing_id = db.query(Baggage).filter(Baggage.id == baggage.id).first()
    if existing_id:
        raise HTTPException(status_code=400, detail="行李ID已存在")
    existing_tag = db.query(Baggage).filter(Baggage.tag_number == baggage.tag_number).first()
    if existing_tag:
        raise HTTPException(status_code=400, detail="行李牌号已存在")
    return services.create_baggage(db, baggage)


@router.get("/baggage/{baggage_id}", response_model=schemas.BaggageResponse)
def get_baggage(baggage_id: str, db: Session = Depends(get_db)):
    baggage = services.get_baggage(db, baggage_id)
    if not baggage:
        raise HTTPException(status_code=404, detail="行李不存在")
    return baggage


@router.get("/baggage/tag/{tag_number}", response_model=schemas.BaggageResponse)
def get_baggage_by_tag(tag_number: str, db: Session = Depends(get_db)):
    baggage = services.get_baggage_by_tag(db, tag_number)
    if not baggage:
        raise HTTPException(status_code=404, detail="行李不存在")
    return baggage


@router.patch("/baggage/{baggage_id}/flight/{flight_id}", response_model=schemas.BaggageResponse)
def bind_flight_to_baggage(baggage_id: str, flight_id: str, db: Session = Depends(get_db)):
    baggage = services.update_baggage_flight(db, baggage_id, flight_id)
    if not baggage:
        raise HTTPException(status_code=404, detail="行李或航班不存在")
    return baggage


@router.post("/baggage/{baggage_id}/security-check")
def security_check(baggage_id: str, request: schemas.SecurityCheckRequest, db: Session = Depends(get_db)):
    success, message = services.process_security_check(db, baggage_id, request)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/baggage/{baggage_id}/sorting")
def sorting(baggage_id: str, request: schemas.SortingRequest, db: Session = Depends(get_db)):
    success, message = services.process_sorting(db, baggage_id, request)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/baggage/{baggage_id}/loading")
def loading(baggage_id: str, request: schemas.LoadingRequest, db: Session = Depends(get_db)):
    success, message = services.process_loading(db, baggage_id, request)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/baggage/{baggage_id}/in-transit")
def in_transit(baggage_id: str, db: Session = Depends(get_db)):
    success, message = services.process_in_transit(db, baggage_id)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/baggage/{baggage_id}/arrival")
def arrival(baggage_id: str, db: Session = Depends(get_db)):
    success, message = services.process_arrival(db, baggage_id)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/baggage/{baggage_id}/conveyor-belt")
def conveyor_belt(baggage_id: str, request: schemas.ConveyorBeltRequest, db: Session = Depends(get_db)):
    success, message = services.process_conveyor_belt(db, baggage_id, request)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/baggage/{baggage_id}/pickup")
def pickup(baggage_id: str, db: Session = Depends(get_db)):
    success, message = services.process_pickup(db, baggage_id)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/baggage/{baggage_id}/report-lost")
def report_lost(baggage_id: str, db: Session = Depends(get_db)):
    success, message = services.report_lost(db, baggage_id)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/baggage/{baggage_id}/report-found")
def report_found(baggage_id: str, db: Session = Depends(get_db)):
    success, message = services.report_found(db, baggage_id)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return {"success": True, "message": message}


@router.post("/baggage/{baggage_id}/claims", response_model=schemas.ClaimResponse, status_code=status.HTTP_201_CREATED)
def file_claim(baggage_id: str, claim: schemas.ClaimCreate, db: Session = Depends(get_db)):
    success, message, db_claim = services.file_claim(db, baggage_id, claim)
    if not success:
        raise HTTPException(status_code=400, detail=message)
    return db_claim


@router.get("/alerts", response_model=List[schemas.AlertResponse])
def list_alerts(db: Session = Depends(get_db)):
    from .models import Alert
    return db.query(Alert).all()


@router.get("/baggage/{baggage_id}/events", response_model=List[schemas.BaggageEventResponse])
def list_baggage_events(baggage_id: str, db: Session = Depends(get_db)):
    baggage = services.get_baggage(db, baggage_id)
    if not baggage:
        raise HTTPException(status_code=404, detail="行李不存在")
    from .models import BaggageEvent
    return db.query(BaggageEvent).filter(BaggageEvent.baggage_id == baggage_id).order_by(BaggageEvent.event_time).all()
