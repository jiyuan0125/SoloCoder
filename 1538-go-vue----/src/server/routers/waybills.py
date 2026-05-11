from typing import List
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from src.core.database import get_db
from src.core.models import Waybill
from src.core.schemas import WaybillCreate, WaybillStatusUpdate, WaybillResponse
from src.core.services import create_waybill, update_waybill_status
from src.core.enums import WaybillStatus

router = APIRouter()


@router.post("", response_model=WaybillResponse, status_code=status.HTTP_201_CREATED)
def create_new_waybill(waybill_data: WaybillCreate, db: Session = Depends(get_db)):
    try:
        return create_waybill(
            db=db,
            producer_id=waybill_data.producer_id,
            disposer_id=waybill_data.disposer_id,
            waste_type=waybill_data.waste_type,
            waste_name=waybill_data.waste_name,
            quantity=waybill_data.quantity,
            hw_code=waybill_data.hw_code,
            unit=waybill_data.unit
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))


@router.get("", response_model=List[WaybillResponse])
def list_waybills(
    producer_id: int = None,
    disposer_id: int = None,
    status: WaybillStatus = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Waybill)
    if producer_id:
        query = query.filter(Waybill.producer_id == producer_id)
    if disposer_id:
        query = query.filter(Waybill.disposer_id == disposer_id)
    if status:
        query = query.filter(Waybill.status == status.value)
    return query.offset(skip).limit(limit).all()


@router.get("/{waybill_id}", response_model=WaybillResponse)
def get_waybill(waybill_id: int, db: Session = Depends(get_db)):
    waybill = db.query(Waybill).filter(Waybill.id == waybill_id).first()
    if not waybill:
        raise HTTPException(status_code=404, detail="联单不存在")
    return waybill


@router.post("/{waybill_id}/status", response_model=WaybillResponse)
def update_status(
    waybill_id: int,
    status_data: WaybillStatusUpdate,
    db: Session = Depends(get_db)
):
    try:
        return update_waybill_status(
            db=db,
            waybill_id=waybill_id,
            new_status=status_data.new_status,
            reject_reason=status_data.reject_reason
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
