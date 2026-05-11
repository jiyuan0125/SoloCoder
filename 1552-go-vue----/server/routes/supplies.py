from typing import List
from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import Supply, SupplyCreate
from server import services

router = APIRouter(prefix="/api/agent-services/{service_id}/supplies", tags=["supplies"])


@router.post("", response_model=Supply)
def add_supply(
    service_id: int,
    supply_in: SupplyCreate,
    db: Session = Depends(get_db),
):
    return services.create_supply(db, service_id, supply_in)


@router.get("", response_model=List[Supply])
def list_supplies(service_id: int, db: Session = Depends(get_db)):
    return services.get_supplies(db, service_id)


@router.post("/{supply_id}/deliver", response_model=Supply)
def deliver_supply(
    service_id: int,
    supply_id: int,
    db: Session = Depends(get_db),
):
    return services.deliver_supply(db, supply_id)


@router.post("/{supply_id}/cancel", response_model=Supply)
def cancel_supply(
    service_id: int,
    supply_id: int,
    db: Session = Depends(get_db),
):
    return services.cancel_supply(db, supply_id)
