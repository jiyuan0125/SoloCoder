from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from server.database import get_db
from server.models import Berth, Ship, HandlingOperation, YardArea
from server.schemas import (
    BerthCreate, BerthUpdate, BerthResponse,
    ShipCreate, ShipUpdate, ShipResponse,
    HandlingOperationCreate, HandlingOperationUpdate, HandlingOperationResponse,
    YardAreaCreate, YardAreaUpdate, YardAreaResponse,
    SchedulingRecommendation
)
from server.services import PortService

router = APIRouter()


@router.post("/berths/", response_model=BerthResponse)
def create_berth(berth: BerthCreate, db: Session = Depends(get_db)):
    existing = db.query(Berth).filter(Berth.name == berth.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="Berth with this name already exists")
    
    db_berth = Berth(**berth.model_dump())
    db.add(db_berth)
    db.commit()
    db.refresh(db_berth)
    db_berth.is_available = db_berth.check_available()
    return db_berth


@router.get("/berths/", response_model=List[BerthResponse])
def list_berths(available_only: Optional[bool] = None, db: Session = Depends(get_db)):
    query = db.query(Berth)
    
    if available_only is not None:
        berths = query.all()
        result = []
        for berth in berths:
            is_available = berth.check_available()
            if (available_only and is_available) or (not available_only and not is_available):
                berth.is_available = is_available
                result.append(berth)
        return result
    
    berths = query.all()
    for berth in berths:
        berth.is_available = berth.check_available()
    return berths


@router.get("/berths/{berth_id}", response_model=BerthResponse)
def get_berth(berth_id: int, db: Session = Depends(get_db)):
    berth = db.query(Berth).filter(Berth.id == berth_id).first()
    if not berth:
        raise HTTPException(status_code=404, detail="Berth not found")
    berth.is_available = berth.check_available()
    return berth


@router.put("/berths/{berth_id}", response_model=BerthResponse)
def update_berth(berth_id: int, berth_update: BerthUpdate, db: Session = Depends(get_db)):
    berth = db.query(Berth).filter(Berth.id == berth_id).first()
    if not berth:
        raise HTTPException(status_code=404, detail="Berth not found")
    
    for key, value in berth_update.model_dump(exclude_unset=True).items():
        setattr(berth, key, value)
    
    db.commit()
    db.refresh(berth)
    berth.is_available = berth.check_available()
    return berth


@router.delete("/berths/{berth_id}")
def delete_berth(berth_id: int, db: Session = Depends(get_db)):
    berth = db.query(Berth).filter(Berth.id == berth_id).first()
    if not berth:
        raise HTTPException(status_code=404, detail="Berth not found")
    
    has_active_ops = db.query(HandlingOperation).filter(
        HandlingOperation.berth_id == berth_id,
        HandlingOperation.status.in_(["pending", "in_progress", "waiting"])
    ).first()
    
    if has_active_ops:
        raise HTTPException(status_code=400, detail="Berth has active operations")
    
    db.delete(berth)
    db.commit()
    return {"message": "Berth deleted"}


@router.post("/ships/", response_model=ShipResponse)
def create_ship(ship: ShipCreate, db: Session = Depends(get_db)):
    existing = db.query(Ship).filter(Ship.name == ship.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="Ship with this name already exists")
    
    db_ship = Ship(**ship.model_dump())
    db.add(db_ship)
    db.commit()
    db.refresh(db_ship)
    return db_ship


@router.get("/ships/", response_model=List[ShipResponse])
def list_ships(status: Optional[str] = None, db: Session = Depends(get_db)):
    query = db.query(Ship)
    if status:
        query = query.filter(Ship.status == status)
    return query.all()


@router.get("/ships/{ship_id}", response_model=ShipResponse)
def get_ship(ship_id: int, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="Ship not found")
    return ship


@router.put("/ships/{ship_id}", response_model=ShipResponse)
def update_ship(ship_id: int, ship_update: ShipUpdate, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="Ship not found")
    
    update_data = ship_update.model_dump(exclude_unset=True)
    
    if "status" in update_data:
        return PortService.update_ship_status(db, ship_id, update_data["status"])
    
    if "draught" in update_data:
        return PortService.update_ship_draught(db, ship_id, update_data["draught"])
    
    for key, value in update_data.items():
        setattr(ship, key, value)
    
    db.commit()
    db.refresh(ship)
    return ship


@router.delete("/ships/{ship_id}")
def delete_ship(ship_id: int, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="Ship not found")
    
    has_active_ops = db.query(HandlingOperation).filter(
        HandlingOperation.ship_id == ship_id,
        HandlingOperation.status.in_(["pending", "in_progress", "waiting"])
    ).first()
    
    if has_active_ops:
        raise HTTPException(status_code=400, detail="Ship has active operations")
    
    db.delete(ship)
    db.commit()
    return {"message": "Ship deleted"}


@router.get("/ships/{ship_id}/schedule", response_model=SchedulingRecommendation)
def get_ship_schedule(ship_id: int, db: Session = Depends(get_db)):
    ship = db.query(Ship).filter(Ship.id == ship_id).first()
    if not ship:
        raise HTTPException(status_code=404, detail="Ship not found")
    
    recommended, waiting = PortService.get_scheduling_recommendation(db, ship)
    
    for berth in recommended:
        berth.is_available = berth.check_available()
    
    return SchedulingRecommendation(
        recommended_berths=recommended,
        waiting_queue=waiting
    )


@router.post("/operations/", response_model=HandlingOperationResponse)
def create_operation(operation: HandlingOperationCreate, db: Session = Depends(get_db)):
    return PortService.create_handling_operation(
        db=db,
        ship_id=operation.ship_id,
        priority=operation.priority,
        cargo_volume=operation.cargo_volume,
        description=operation.description,
        berth_id=operation.berth_id,
        yard_area_id=operation.yard_area_id
    )


@router.get("/operations/", response_model=List[HandlingOperationResponse])
def list_operations(
    status: Optional[str] = None,
    ship_id: Optional[int] = None,
    berth_id: Optional[int] = None,
    db: Session = Depends(get_db)
):
    query = db.query(HandlingOperation)
    
    if status:
        query = query.filter(HandlingOperation.status == status)
    if ship_id:
        query = query.filter(HandlingOperation.ship_id == ship_id)
    if berth_id:
        query = query.filter(HandlingOperation.berth_id == berth_id)
    
    return query.order_by(
        HandlingOperation.priority.desc(),
        HandlingOperation.created_at.asc()
    ).all()


@router.get("/operations/{operation_id}", response_model=HandlingOperationResponse)
def get_operation(operation_id: int, db: Session = Depends(get_db)):
    operation = db.query(HandlingOperation).filter(HandlingOperation.id == operation_id).first()
    if not operation:
        raise HTTPException(status_code=404, detail="Operation not found")
    return operation


@router.put("/operations/{operation_id}", response_model=HandlingOperationResponse)
def update_operation(operation_id: int, op_update: HandlingOperationUpdate, db: Session = Depends(get_db)):
    operation = db.query(HandlingOperation).filter(HandlingOperation.id == operation_id).first()
    if not operation:
        raise HTTPException(status_code=404, detail="Operation not found")
    
    update_data = op_update.model_dump(exclude_unset=True)
    
    if "status" in update_data and update_data["status"] == "completed":
        return PortService.complete_handling_operation(db, operation_id)
    
    for key, value in update_data.items():
        setattr(operation, key, value)
    
    db.commit()
    db.refresh(operation)
    return operation


@router.delete("/operations/{operation_id}")
def delete_operation(operation_id: int, db: Session = Depends(get_db)):
    operation = db.query(HandlingOperation).filter(HandlingOperation.id == operation_id).first()
    if not operation:
        raise HTTPException(status_code=404, detail="Operation not found")
    
    if operation.status in ["in_progress"]:
        raise HTTPException(status_code=400, detail="Cannot delete in-progress operation")
    
    db.delete(operation)
    db.commit()
    return {"message": "Operation deleted"}


@router.post("/yard_areas/", response_model=YardAreaResponse)
def create_yard_area(yard: YardAreaCreate, db: Session = Depends(get_db)):
    existing = db.query(YardArea).filter(YardArea.name == yard.name).first()
    if existing:
        raise HTTPException(status_code=400, detail="Yard area with this name already exists")
    
    db_yard = YardArea(**yard.model_dump())
    db.add(db_yard)
    db.commit()
    db.refresh(db_yard)
    usage = db_yard.calculate_usage(db)
    db_yard.current_usage = usage
    db_yard.remaining_capacity = db_yard.total_capacity - usage
    return db_yard


@router.get("/yard_areas/", response_model=List[YardAreaResponse])
def list_yard_areas(available_only: Optional[bool] = None, db: Session = Depends(get_db)):
    query = db.query(YardArea)
    
    if available_only is not None:
        query = query.filter(YardArea.is_available == available_only)
    
    yards = query.all()
    for yard in yards:
        usage = yard.calculate_usage(db)
        yard.current_usage = usage
        yard.remaining_capacity = yard.total_capacity - usage
    
    return yards


@router.get("/yard_areas/{yard_id}", response_model=YardAreaResponse)
def get_yard_area(yard_id: int, db: Session = Depends(get_db)):
    yard = db.query(YardArea).filter(YardArea.id == yard_id).first()
    if not yard:
        raise HTTPException(status_code=404, detail="Yard area not found")
    
    usage = yard.calculate_usage(db)
    yard.current_usage = usage
    yard.remaining_capacity = yard.total_capacity - usage
    return yard


@router.put("/yard_areas/{yard_id}", response_model=YardAreaResponse)
def update_yard_area(yard_id: int, yard_update: YardAreaUpdate, db: Session = Depends(get_db)):
    yard = db.query(YardArea).filter(YardArea.id == yard_id).first()
    if not yard:
        raise HTTPException(status_code=404, detail="Yard area not found")
    
    update_data = yard_update.model_dump(exclude_unset=True)
    
    if "is_available" in update_data:
        if update_data["is_available"]:
            PortService.mark_yard_area_available(db, yard_id)
        else:
            PortService.mark_yard_area_unavailable(db, yard_id)
        db.refresh(yard)
    
    for key, value in update_data.items():
        if key != "is_available":
            setattr(yard, key, value)
    
    if "is_available" not in update_data:
        db.commit()
        db.refresh(yard)
    
    usage = yard.calculate_usage(db)
    yard.current_usage = usage
    yard.remaining_capacity = yard.total_capacity - usage
    return yard


@router.delete("/yard_areas/{yard_id}")
def delete_yard_area(yard_id: int, db: Session = Depends(get_db)):
    yard = db.query(YardArea).filter(YardArea.id == yard_id).first()
    if not yard:
        raise HTTPException(status_code=404, detail="Yard area not found")
    
    has_active_ops = db.query(HandlingOperation).filter(
        HandlingOperation.yard_area_id == yard_id,
        HandlingOperation.status.in_(["pending", "in_progress", "waiting"])
    ).first()
    
    if has_active_ops:
        raise HTTPException(status_code=400, detail="Yard area has active operations")
    
    db.delete(yard)
    db.commit()
    return {"message": "Yard area deleted"}
