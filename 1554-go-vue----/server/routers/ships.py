from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy.exc import IntegrityError

from server.database import get_db
from server.schemas import (
    ShipCreate, ShipUpdate, ShipResponse, ShipDetailResponse,
    LicenseResponse, InspectionResponse, PenaltyResponse, RectificationResponse
)
from server import services

router = APIRouter(prefix="/ships", tags=["ships"])


@router.get("", response_model=List[ShipResponse])
def list_ships(
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    return services.get_ships(db, skip=skip, limit=limit)


@router.post("", response_model=ShipResponse)
def create_ship(ship: ShipCreate, db: Session = Depends(get_db)):
    try:
        db_ship = services.create_ship(db, ship)
        db.commit()
        return db_ship
    except IntegrityError:
        db.rollback()
        raise HTTPException(status_code=400, detail="IMO编号已存在")


@router.get("/{ship_id}", response_model=ShipDetailResponse)
def get_ship(ship_id: int, db: Session = Depends(get_db)):
    detail = services.get_ship_detail(db, ship_id)
    if not detail:
        raise HTTPException(status_code=404, detail="船舶不存在")
    
    services.check_overdue_penalties(db)
    db.commit()
    
    ship = detail["ship"]
    return ShipDetailResponse(
        id=ship.id,
        imo_number=ship.imo_number,
        name=ship.name,
        registration_port=ship.registration_port,
        gross_tonnage=ship.gross_tonnage,
        length=ship.length,
        width=ship.width,
        owner_name=ship.owner_name,
        created_at=ship.created_at,
        updated_at=ship.updated_at,
        is_high_priority=detail["is_high_priority"],
        licenses=[LicenseResponse.model_validate(l) for l in detail["licenses"]],
        inspections=[InspectionResponse.model_validate(i) for i in detail["inspections"]],
        penalties=[PenaltyResponse.model_validate(p) for p in detail["penalties"]],
        rectifications=[RectificationResponse.model_validate(r) for r in detail["rectifications"]]
    )


@router.put("/{ship_id}", response_model=ShipResponse)
def update_ship(ship_id: int, ship: ShipUpdate, db: Session = Depends(get_db)):
    db_ship = services.update_ship(db, ship_id, ship)
    if not db_ship:
        raise HTTPException(status_code=404, detail="The ship does not exist")
    db.commit()
    return db_ship


@router.post("/{ship_id}/associate")
def associate_records(
    ship_id: int,
    temp_identifier: str = Query(...),
    db: Session = Depends(get_db)
):
    ship = services.get_ship(db, ship_id)
    if not ship:
        raise HTTPException(status_code=404, detail="The ship does not exist")
    
    count = services.associate_records_with_ship(db, ship_id, temp_identifier)
    db.commit()
    return {"associated_records": count}
