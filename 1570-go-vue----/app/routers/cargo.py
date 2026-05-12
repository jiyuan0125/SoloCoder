from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime

from ..database import get_db
from ..models import Cargo, Flight
from ..schemas import CargoAccept, CargoWeigh, CargoResponse
from ..utils import calculate_chargeable_weight, calculate_freight_charge, calculate_storage_charge

router = APIRouter(prefix="/cargo", tags=["cargo"])


@router.post("/{cargo_number}/accept", response_model=CargoResponse)
def accept_cargo(cargo_number: str, cargo_data: CargoAccept, db: Session = Depends(get_db)):
    db_cargo = db.query(Cargo).filter(Cargo.cargo_number == cargo_number).first()
    if db_cargo:
        raise HTTPException(status_code=400, detail="货运单号已存在")
    
    cargo_type_lower = cargo_data.cargo_type.lower()
    
    is_dangerous = cargo_type_lower in ["dangerous", "hazardous", "危险品"]
    is_live_animal = cargo_type_lower in ["live_animal", "活体动物", "活物"]
    
    if is_dangerous and not cargo_data.declaration_number:
        raise HTTPException(status_code=400, detail="危险品必须提供申报单号")
    
    if is_live_animal and not cargo_data.quarantine_certificate_number:
        raise HTTPException(status_code=400, detail="活体动物必须提供检疫证明编号")
    
    db_cargo = Cargo(
        cargo_number=cargo_number,
        shipper=cargo_data.shipper,
        consignee=cargo_data.consignee,
        cargo_type=cargo_data.cargo_type,
        description=cargo_data.description,
        status="accepted",
        declaration_number=cargo_data.declaration_number,
        quarantine_certificate_number=cargo_data.quarantine_certificate_number
    )
    db.add(db_cargo)
    db.commit()
    db.refresh(db_cargo)
    
    return db_cargo


@router.post("/{cargo_number}/weigh", response_model=CargoResponse)
def weigh_cargo(cargo_number: str, weigh_data: CargoWeigh, db: Session = Depends(get_db)):
    db_cargo = db.query(Cargo).filter(Cargo.cargo_number == cargo_number).first()
    if not db_cargo:
        raise HTTPException(status_code=404, detail="货运单不存在")
    
    if db_cargo.status != "accepted":
        raise HTTPException(status_code=400, detail="该货运单状态不允许称重")
    
    volume_weight = weigh_data.volume_weight or 0
    
    chargeable_weight = calculate_chargeable_weight(weigh_data.actual_weight, volume_weight)
    
    db_cargo.actual_weight = weigh_data.actual_weight
    db_cargo.volume = weigh_data.volume
    db_cargo.volume_weight = volume_weight
    db_cargo.chargeable_weight = chargeable_weight
    db_cargo.status = "weighed"
    
    db.commit()
    db.refresh(db_cargo)
    
    return db_cargo


@router.post("/{cargo_number}/assign-flight/{flight_number}", response_model=CargoResponse)
def assign_cargo_to_flight(cargo_number: str, flight_number: str, db: Session = Depends(get_db)):
    db_cargo = db.query(Cargo).filter(Cargo.cargo_number == cargo_number).first()
    if not db_cargo:
        raise HTTPException(status_code=404, detail="货运单不存在")
    
    if db_cargo.status != "weighed":
        raise HTTPException(status_code=400, detail="该货运单状态不允许分配航班")
    
    db_flight = db.query(Flight).filter(Flight.flight_number == flight_number).first()
    if not db_flight:
        raise HTTPException(status_code=404, detail="航班不存在")
    
    if db_cargo.chargeable_weight is None:
        raise HTTPException(status_code=400, detail="货运单尚未称重")
    
    freight_charge = calculate_freight_charge(db_cargo.chargeable_weight, db_flight.unit_price)
    
    db_cargo.flight_id = db_flight.id
    db_cargo.price_per_kg = db_flight.unit_price
    db_cargo.freight_charge = freight_charge
    db_cargo.status = "flight_assigned"
    
    db.commit()
    db.refresh(db_cargo)
    
    return db_cargo


@router.post("/{cargo_number}/arrive", response_model=CargoResponse)
def mark_cargo_arrived(cargo_number: str, db: Session = Depends(get_db)):
    db_cargo = db.query(Cargo).filter(Cargo.cargo_number == cargo_number).first()
    if not db_cargo:
        raise HTTPException(status_code=404, detail="货运单不存在")
    
    db_cargo.arrival_time = datetime.utcnow()
    db_cargo.status = "arrived"
    db_cargo.storage_charge = 0.0
    
    db.commit()
    db.refresh(db_cargo)
    
    return db_cargo


@router.post("/{cargo_number}/calculate-storage", response_model=CargoResponse)
def calculate_cargo_storage_charge(cargo_number: str, db: Session = Depends(get_db)):
    db_cargo = db.query(Cargo).filter(Cargo.cargo_number == cargo_number).first()
    if not db_cargo:
        raise HTTPException(status_code=404, detail="货运单不存在")
    
    if db_cargo.arrival_time is None:
        raise HTTPException(status_code=400, detail="该货运单尚未到达")
    
    if db_cargo.chargeable_weight is None:
        raise HTTPException(status_code=400, detail="货运单重量信息缺失")
    
    storage_charge = calculate_storage_charge(db_cargo.arrival_time, db_cargo.chargeable_weight)
    db_cargo.storage_charge = storage_charge
    db.commit()
    db.refresh(db_cargo)
    
    return db_cargo


@router.get("/{cargo_number}", response_model=CargoResponse)
def get_cargo(cargo_number: str, db: Session = Depends(get_db)):
    db_cargo = db.query(Cargo).filter(Cargo.cargo_number == cargo_number).first()
    if not db_cargo:
        raise HTTPException(status_code=404, detail="货运单不存在")
    return db_cargo


@router.get("", response_model=List[CargoResponse])
def list_cargo(db: Session = Depends(get_db)):
    return db.query(Cargo).all()
