from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime

from ..database import get_db
from ..models import Cargo, Compartment, LoadAssignment, Flight
from ..schemas import LoadAssignmentResponse
from ..utils import (
    check_cargo_compartment_compatibility,
    check_compartment_mixed_cargo,
    check_pending_assignments_mixed_cargo,
    check_front_rear_weight_balance,
    check_min_load_rate,
    create_low_load_alert
)

router = APIRouter(prefix="/load", tags=["load"])


@router.post("/{compartment_number}/assign/{cargo_number}", response_model=LoadAssignmentResponse)
def assign_cargo_to_compartment(
    compartment_number: str,
    cargo_number: str,
    db: Session = Depends(get_db)
):
    db_cargo = db.query(Cargo).filter(Cargo.cargo_number == cargo_number).first()
    if not db_cargo:
        raise HTTPException(status_code=404, detail="货运单不存在")
    
    if db_cargo.status != "flight_assigned":
        raise HTTPException(status_code=400, detail="该货运单状态不允许分配舱位")
    
    if db_cargo.chargeable_weight is None:
        raise HTTPException(status_code=400, detail="货运单重量信息缺失")
    
    db_compartment = db.query(Compartment).filter(
        Compartment.compartment_number == compartment_number
    ).first()
    if not db_compartment:
        raise HTTPException(status_code=404, detail="舱位不存在")
    
    if db_cargo.flight_id != db_compartment.flight_id:
        raise HTTPException(status_code=400, detail="货运单与舱位不属于同一航班")
    
    if db_cargo.chargeable_weight > db_compartment.remaining_capacity_weight:
        raise HTTPException(
            status_code=400,
            detail=f"舱位剩余容量不足。需要: {db_cargo.chargeable_weight}kg, 剩余: {db_compartment.remaining_capacity_weight}kg"
        )
    
    ok, msg = check_cargo_compartment_compatibility(db_cargo, db_compartment)
    if not ok:
        raise HTTPException(status_code=400, detail=msg)
    
    ok, msg = check_compartment_mixed_cargo(db, db_compartment, db_cargo.cargo_type)
    if not ok:
        raise HTTPException(status_code=400, detail=msg)
    
    existing_assignment = db.query(LoadAssignment).filter(
        LoadAssignment.cargo_id == db_cargo.id,
        LoadAssignment.is_confirmed == False
    ).first()
    if existing_assignment:
        raise HTTPException(status_code=400, detail="该货运单已有未确认的配载分配")
    
    assignment = LoadAssignment(
        cargo_id=db_cargo.id,
        compartment_id=db_compartment.id,
        assigned_weight=db_cargo.chargeable_weight,
        is_confirmed=False
    )
    db.add(assignment)
    
    db_compartment.remaining_capacity_weight -= db_cargo.chargeable_weight
    
    db.commit()
    db.refresh(assignment)
    db.refresh(db_compartment)
    
    db_cargo.status = "assigned"
    db.commit()
    db.refresh(db_cargo)
    
    return LoadAssignmentResponse(
        id=assignment.id,
        cargo_id=assignment.cargo_id,
        cargo_number=db_cargo.cargo_number,
        compartment_id=assignment.compartment_id,
        compartment_number=db_compartment.compartment_number,
        assigned_weight=assignment.assigned_weight,
        is_confirmed=assignment.is_confirmed,
        confirmed_at=assignment.confirmed_at
    )


@router.post("/{compartment_number}/confirm", response_model=List[LoadAssignmentResponse])
def confirm_compartment_load(
    compartment_number: str,
    db: Session = Depends(get_db)
):
    db_compartment = db.query(Compartment).filter(
        Compartment.compartment_number == compartment_number
    ).first()
    if not db_compartment:
        raise HTTPException(status_code=404, detail="舱位不存在")
    
    pending_assignments = db.query(LoadAssignment).filter(
        LoadAssignment.compartment_id == db_compartment.id,
        LoadAssignment.is_confirmed == False
    ).all()
    
    if not pending_assignments:
        raise HTTPException(status_code=400, detail="该舱位没有待确认的配载")
    
    flight = db.query(Flight).filter(Flight.id == db_compartment.flight_id).first()
    
    ok, msg, diff = check_front_rear_weight_balance(db, flight.id)
    if not ok:
        raise HTTPException(status_code=400, detail=msg)
    
    ok, msg = check_pending_assignments_mixed_cargo(db, db_compartment, pending_assignments)
    if not ok:
        raise HTTPException(status_code=400, detail=msg)
    
    for assignment in pending_assignments:
        assignment.is_confirmed = True
        assignment.confirmed_at = datetime.utcnow()
    
    db.commit()
    
    ok, current_rate, min_rate = check_min_load_rate(db, flight)
    if not ok:
        create_low_load_alert(db, flight, current_rate, min_rate)
    
    confirmed_assignments = db.query(LoadAssignment).filter(
        LoadAssignment.compartment_id == db_compartment.id,
        LoadAssignment.is_confirmed == True
    ).all()
    
    result = []
    for assignment in confirmed_assignments:
        cargo = db.query(Cargo).filter(Cargo.id == assignment.cargo_id).first()
        cargo.status = "loaded"
        db.commit()
        
        result.append(LoadAssignmentResponse(
            id=assignment.id,
            cargo_id=assignment.cargo_id,
            cargo_number=cargo.cargo_number if cargo else "",
            compartment_id=assignment.compartment_id,
            compartment_number=db_compartment.compartment_number,
            assigned_weight=assignment.assigned_weight,
            is_confirmed=assignment.is_confirmed,
            confirmed_at=assignment.confirmed_at
        ))
    
    return result


@router.get("/{compartment_number}/assignments", response_model=List[LoadAssignmentResponse])
def list_compartment_assignments(
    compartment_number: str,
    db: Session = Depends(get_db)
):
    db_compartment = db.query(Compartment).filter(
        Compartment.compartment_number == compartment_number
    ).first()
    if not db_compartment:
        raise HTTPException(status_code=404, detail="舱位不存在")
    
    assignments = db.query(LoadAssignment).filter(
        LoadAssignment.compartment_id == db_compartment.id
    ).all()
    
    result = []
    for assignment in assignments:
        cargo = db.query(Cargo).filter(Cargo.id == assignment.cargo_id).first()
        result.append(LoadAssignmentResponse(
            id=assignment.id,
            cargo_id=assignment.cargo_id,
            cargo_number=cargo.cargo_number if cargo else "",
            compartment_id=assignment.compartment_id,
            compartment_number=db_compartment.compartment_number,
            assigned_weight=assignment.assigned_weight,
            is_confirmed=assignment.is_confirmed,
            confirmed_at=assignment.confirmed_at
        ))
    
    return result


@router.get("/cargo/{cargo_number}", response_model=List[LoadAssignmentResponse])
def list_cargo_assignments(
    cargo_number: str,
    db: Session = Depends(get_db)
):
    db_cargo = db.query(Cargo).filter(Cargo.cargo_number == cargo_number).first()
    if not db_cargo:
        raise HTTPException(status_code=404, detail="货运单不存在")
    
    assignments = db.query(LoadAssignment).filter(
        LoadAssignment.cargo_id == db_cargo.id
    ).all()
    
    result = []
    for assignment in assignments:
        compartment = db.query(Compartment).filter(
            Compartment.id == assignment.compartment_id
        ).first()
        result.append(LoadAssignmentResponse(
            id=assignment.id,
            cargo_id=assignment.cargo_id,
            cargo_number=db_cargo.cargo_number,
            compartment_id=assignment.compartment_id,
            compartment_number=compartment.compartment_number if compartment else "",
            assigned_weight=assignment.assigned_weight,
            is_confirmed=assignment.is_confirmed,
            confirmed_at=assignment.confirmed_at
        ))
    
    return result
