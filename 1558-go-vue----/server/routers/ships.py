from typing import List

from fastapi import APIRouter, Depends, Query
from sqlalchemy.orm import Session

from server.config import get_db
from server.schemas.schemas import ShipCreate, ShipResponse
from server.services import create_ship, get_ships, get_ship

router = APIRouter(prefix="/ships", tags=["ships"])


@router.post("", response_model=ShipResponse)
def create_ship_endpoint(
    ship_data: ShipCreate,
    actor: str = Query(..., description="操作人"),
    db: Session = Depends(get_db),
):
    return create_ship(db, ship_data, actor)


@router.get("", response_model=List[ShipResponse])
def list_ships(
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    return get_ships(db, skip, limit)


@router.get("/{ship_id}", response_model=ShipResponse)
def get_ship_endpoint(
    ship_id: int,
    db: Session = Depends(get_db),
):
    return get_ship(db, ship_id)
