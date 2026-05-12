from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from server.database import get_db
from server.schemas import DailySettlementResponse
from server.services import DailySettlementService
import datetime

router = APIRouter(prefix="/settlements", tags=["settlements"])


@router.post("/generate", response_model=DailySettlementResponse)
def generate_daily_settlement(
    shop_id: int,
    settlement_date: Optional[datetime.date] = None,
    db: Session = Depends(get_db),
):
    settlement = DailySettlementService.get_or_create_draft(db, shop_id, settlement_date)
    return _build_settlement_response(settlement)


@router.get("/{settlement_id}", response_model=DailySettlementResponse)
def get_settlement(settlement_id: int, db: Session = Depends(get_db)):
    settlement = DailySettlementService.get_by_id(db, settlement_id)
    if not settlement:
        raise HTTPException(status_code=404, detail="日结记录不存在")
    return _build_settlement_response(settlement)


@router.post("/{settlement_id}/confirm", response_model=DailySettlementResponse)
def confirm_settlement(settlement_id: int, db: Session = Depends(get_db)):
    settlement = DailySettlementService.confirm(db, settlement_id)
    return _build_settlement_response(settlement)


def _build_settlement_response(settlement) -> DailySettlementResponse:
    resp = DailySettlementResponse.model_validate(settlement)
    if settlement.shop:
        resp.shop_name = settlement.shop.name
    return resp
