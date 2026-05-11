from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.database import get_db
from server.schemas import FeeSettlement, FeeSettlementCreate
from server import services

router = APIRouter(prefix="/api/agent-services/{service_id}/fee-settlement", tags=["fees"])


@router.post("", response_model=FeeSettlement)
def create_fee_settlement(
    service_id: int,
    fee_in: FeeSettlementCreate,
    db: Session = Depends(get_db),
):
    return services.create_fee_settlement(db, service_id, fee_in)


@router.get("", response_model=FeeSettlement)
def get_fee_settlement(service_id: int, db: Session = Depends(get_db)):
    fee = services.get_fee_settlement(db, service_id)
    if not fee:
        raise HTTPException(status_code=404, detail="No fee settlement found")
    return fee


@router.post("/settle", response_model=FeeSettlement)
def settle_fee(service_id: int, db: Session = Depends(get_db)):
    return services.settle_fee(db, service_id)
