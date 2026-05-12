from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from server.database import get_db
from server.schemas import (
    SaleCreate,
    SaleResponse,
    SaleItemResponse,
    SaleItemBatchResponse,
    RefundRequestCreate,
    RefundRequestResponse,
    RefundItemResponse,
    RefundApproval,
)
from server.services import (
    SaleService,
    RefundService,
)

router = APIRouter(prefix="/sales", tags=["sales"])


@router.post("", response_model=SaleResponse, status_code=201)
def create_sale(data: SaleCreate, db: Session = Depends(get_db)):
    sale = SaleService.create(db, data)
    return _build_sale_response(sale)


@router.get("", response_model=List[SaleResponse])
def list_sales(shop_id: Optional[int] = None, skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    sales = SaleService.get_all(db, shop_id=shop_id, skip=skip, limit=limit)
    return [_build_sale_response(s) for s in sales]


@router.get("/{sale_id}", response_model=SaleResponse)
def get_sale(sale_id: int, db: Session = Depends(get_db)):
    sale = SaleService.get_by_id(db, sale_id)
    if not sale:
        raise HTTPException(status_code=404, detail="销售记录不存在")
    return _build_sale_response(sale)


@router.get("/{sale_id}/details")
def get_sale_details(sale_id: int, db: Session = Depends(get_db)):
    sale = SaleService.get_by_id(db, sale_id)
    if not sale:
        raise HTTPException(status_code=404, detail="销售记录不存在")
    return _build_sale_response(sale)


@router.get("/{sale_id}/refund", response_model=RefundRequestResponse)
def get_sale_refund_status(sale_id: int, db: Session = Depends(get_db)):
    refund = RefundService.get_by_sale(db, sale_id)
    if not refund:
        raise HTTPException(status_code=404, detail="该订单没有退货记录")
    return _build_refund_response(refund)


@router.post("/{sale_id}/refund", response_model=RefundRequestResponse, status_code=201)
def apply_refund(sale_id: int, data: RefundRequestCreate, db: Session = Depends(get_db)):
    if data.sale_id != sale_id:
        raise HTTPException(status_code=400, detail="sale_id不匹配")
    refund = RefundService.apply(db, data)
    return _build_refund_response(refund)


@router.post("/refunds/{refund_id}/approve", response_model=RefundRequestResponse)
def approve_refund(refund_id: int, data: RefundApproval, db: Session = Depends(get_db)):
    refund = RefundService.approve(db, refund_id, data)
    return _build_refund_response(refund)


@router.post("/refunds/{refund_id}/reject", response_model=RefundRequestResponse)
def reject_refund(refund_id: int, data: RefundApproval, rejection_reason: str, db: Session = Depends(get_db)):
    refund = RefundService.reject(db, refund_id, data, rejection_reason)
    return _build_refund_response(refund)


@router.post("/refunds/{refund_id}/process", response_model=RefundRequestResponse)
def process_refund(refund_id: int, db: Session = Depends(get_db)):
    refund = RefundService.process_refund(db, refund_id)
    return _build_refund_response(refund)


@router.get("/refunds/{refund_id}", response_model=RefundRequestResponse)
def get_refund(refund_id: int, db: Session = Depends(get_db)):
    refund = RefundService.get_by_id(db, refund_id)
    if not refund:
        raise HTTPException(status_code=404, detail="退货申请不存在")
    return _build_refund_response(refund)


def _build_sale_response(sale) -> SaleResponse:
    resp = SaleResponse.model_validate(sale)
    if sale.shop:
        resp.shop_name = sale.shop.name
    if sale.passenger:
        resp.passenger_name = sale.passenger.name
        resp.passport_number = sale.passenger.passport_number
        resp.flight_number = sale.passenger.flight_number
        resp.flight_departure_time = sale.passenger.flight_departure_time
    
    items = []
    for si in sale.items:
        item = SaleItemResponse.model_validate(si)
        if si.product:
            item.product_name = si.product.name
            item.product_sku = si.product.sku
            item.product_category = si.product.category
        item.batches = [SaleItemBatchResponse.model_validate(b) for b in si.batches]
        items.append(item)
    resp.items = items
    return resp


def _build_refund_response(refund) -> RefundRequestResponse:
    resp = RefundRequestResponse.model_validate(refund)
    if refund.sale:
        resp.sale_number = refund.sale.sale_number
    
    items = []
    for ri in refund.items:
        item = RefundItemResponse.model_validate(ri)
        if refund.sale:
            for si in refund.sale.items:
                if si.id == ri.sale_item_id and si.product:
                    item.product_name = si.product.name
                    break
        items.append(item)
    resp.items = items
    return resp
