from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from server.core.database import get_db
from server.schemas import (
    PurchaseApprove,
    PurchaseOrder,
    PurchaseReceive,
    PurchaseRequestCreate,
    PurchaseRequestResponse,
    SparePartCreate,
    SparePartResponse,
    SparePartUpdate,
    SpareUsageCreate,
    SpareUsageResponse,
)
from server.services import InventoryService

router = APIRouter(tags=["备件库存"])


@router.post("/spare-parts", response_model=SparePartResponse)
def create_spare_part(data: SparePartCreate, db: Session = Depends(get_db)):
    service = InventoryService(db)
    return service.create_spare_part(data.model_dump())


@router.get("/spare-parts", response_model=List[SparePartResponse])
def list_spare_parts(db: Session = Depends(get_db)):
    service = InventoryService(db)
    return service.get_spare_parts()


@router.get("/spare-parts/low-stock", response_model=List[SparePartResponse])
def list_low_stock_parts(db: Session = Depends(get_db)):
    service = InventoryService(db)
    return service.check_low_stock()


@router.get("/spare-parts/{part_id}", response_model=SparePartResponse)
def get_spare_part(part_id: int, db: Session = Depends(get_db)):
    service = InventoryService(db)
    part = service.get_spare_part(part_id)
    if not part:
        raise HTTPException(status_code=404, detail="备件不存在")
    return part


@router.put("/spare-parts/{part_id}", response_model=SparePartResponse)
def update_spare_part(part_id: int, data: SparePartUpdate, db: Session = Depends(get_db)):
    service = InventoryService(db)
    part = service.update_spare_part(part_id, data.model_dump(exclude_unset=True))
    if not part:
        raise HTTPException(status_code=404, detail="备件不存在")
    return part


@router.post("/spare-usages", response_model=SpareUsageResponse)
def record_spare_usage(data: SpareUsageCreate, db: Session = Depends(get_db)):
    service = InventoryService(db)
    try:
        usage = service.record_usage(
            lighthouse_id=data.lighthouse_id,
            spare_part_id=data.spare_part_id,
            quantity=data.quantity,
            usage_type=data.usage_type,
            work_order_id=data.work_order_id,
            notes=data.notes,
        )
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
    if not usage:
        raise HTTPException(status_code=404, detail="备件不存在")
    return usage


@router.get("/spare-usages", response_model=List[SpareUsageResponse])
def list_spare_usages(lighthouse_id: Optional[int] = None, spare_part_id: Optional[int] = None, db: Session = Depends(get_db)):
    service = InventoryService(db)
    return service.get_spare_usages(lighthouse_id, spare_part_id)


@router.post("/purchase-requests", response_model=PurchaseRequestResponse)
def create_purchase_request(data: PurchaseRequestCreate, db: Session = Depends(get_db)):
    service = InventoryService(db)
    items = [item.model_dump() for item in data.items]
    return service.create_purchase_request(data.requested_by, data.reason, items)


@router.get("/purchase-requests", response_model=List[PurchaseRequestResponse])
def list_purchase_requests(status: Optional[str] = None, db: Session = Depends(get_db)):
    service = InventoryService(db)
    return service.get_purchase_requests(status)


@router.get("/purchase-requests/{request_id}", response_model=PurchaseRequestResponse)
def get_purchase_request(request_id: int, db: Session = Depends(get_db)):
    service = InventoryService(db)
    request = service.get_purchase_request(request_id)
    if not request:
        raise HTTPException(status_code=404, detail="采购申请不存在")
    return request


@router.post("/purchase-requests/{request_id}/approve", response_model=PurchaseRequestResponse)
def approve_purchase(request_id: int, data: PurchaseApprove, db: Session = Depends(get_db)):
    service = InventoryService(db)
    request = service.approve_purchase(request_id, data.approved_by)
    if not request:
        raise HTTPException(status_code=404, detail="采购申请不存在或状态不正确")
    return request


@router.post("/purchase-requests/{request_id}/order", response_model=PurchaseRequestResponse)
def order_purchase(request_id: int, data: PurchaseOrder, db: Session = Depends(get_db)):
    service = InventoryService(db)
    request = service.order_purchase(request_id, data.ordered_by)
    if not request:
        raise HTTPException(status_code=404, detail="采购申请不存在或状态不正确")
    return request


@router.post("/purchase-requests/{request_id}/receive", response_model=PurchaseRequestResponse)
def receive_purchase(request_id: int, data: PurchaseReceive, db: Session = Depends(get_db)):
    service = InventoryService(db)
    request = service.receive_purchase(request_id, data.received_by)
    if not request:
        raise HTTPException(status_code=404, detail="采购申请不存在或状态不正确")
    return request


@router.post("/spare-parts/{part_id}/auto-purchase")
def auto_create_purchase(part_id: int, requested_by: str = "system", db: Session = Depends(get_db)):
    from server.models import SparePart

    part = db.query(SparePart).filter(SparePart.id == part_id).first()
    if not part:
        raise HTTPException(status_code=404, detail="备件不存在")

    if part.stock_quantity >= part.safety_stock:
        raise HTTPException(status_code=400, detail="库存充足，无需采购")

    if part.in_transit_quantity > 0:
        raise HTTPException(status_code=400, detail="已有在途采购")

    service = InventoryService(db)
    purchase_qty = max(part.safety_stock * 2, 1)
    request = service.create_purchase_request(
        requested_by=requested_by,
        reason=f"安全库存不足，当前{part.stock_quantity}，安全库存{part.safety_stock}",
        items=[{"spare_part_id": part_id, "quantity": purchase_qty, "unit_price": part.price}],
    )
    return request
