from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session

from ..dependencies import get_db
from ...core.schemas import (
    PurchaseOrderCreate,
    PurchaseOrderUpdate,
    PurchaseOrderApprove,
    PurchaseOrderReceive,
    PurchaseOrderResponse,
)
from ...core.service import PurchaseOrderService, BusinessException


router = APIRouter(prefix="/purchase-orders", tags=["purchase_orders"])


@router.post("", response_model=PurchaseOrderResponse)
def create_purchase_order(
    po_create: PurchaseOrderCreate, db: Session = Depends(get_db)
):
    try:
        po = PurchaseOrderService(db).create(po_create)
        return PurchaseOrderService(db)._to_response(po)
    except BusinessException as e:
        raise HTTPException(status_code=400, detail=e.detail)


@router.get("", response_model=List[PurchaseOrderResponse])
def list_purchase_orders(
    store_id: Optional[int] = Query(None),
    status: Optional[str] = Query(None),
    db: Session = Depends(get_db),
):
    service = PurchaseOrderService(db)
    if store_id:
        return service.list_by_store(store_id, status)
    return service.list_all(status)


@router.get("/{po_id}", response_model=PurchaseOrderResponse)
def get_purchase_order(po_id: int, db: Session = Depends(get_db)):
    po = PurchaseOrderService(db).get_by_id(po_id)
    if not po:
        raise HTTPException(
            status_code=404, detail=f"采购订单 ID {po_id} 不存在"
        )
    return PurchaseOrderService(db)._to_response(po)


@router.patch("/{po_id}", response_model=PurchaseOrderResponse)
def update_purchase_order(
    po_id: int,
    update_data: PurchaseOrderUpdate,
    db: Session = Depends(get_db),
):
    try:
        po = PurchaseOrderService(db).update(po_id, update_data)
        return PurchaseOrderService(db)._to_response(po)
    except BusinessException as e:
        raise HTTPException(status_code=400, detail=e.detail)


@router.post("/{po_id}/approve", response_model=PurchaseOrderResponse)
def approve_purchase_order(
    po_id: int,
    approve_data: PurchaseOrderApprove,
    db: Session = Depends(get_db),
):
    try:
        po = PurchaseOrderService(db).approve(po_id, approve_data.approved)
        return PurchaseOrderService(db)._to_response(po)
    except BusinessException as e:
        raise HTTPException(status_code=400, detail=e.detail)


@router.post("/{po_id}/receive", response_model=PurchaseOrderResponse)
def receive_purchase_order(
    po_id: int,
    receive_data: PurchaseOrderReceive,
    db: Session = Depends(get_db),
):
    try:
        po = PurchaseOrderService(db).receive(
            po_id, receive_data.received_quantity, receive_data.actual_arrival_date
        )
        return PurchaseOrderService(db)._to_response(po)
    except BusinessException as e:
        raise HTTPException(status_code=400, detail=e.detail)
