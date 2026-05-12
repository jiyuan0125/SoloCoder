from datetime import datetime
from typing import List, Optional

from pydantic import BaseModel, Field


class SparePartBase(BaseModel):
    name: str
    code: str
    category: Optional[str] = None
    unit: str
    stock_quantity: float = 0
    safety_stock: float = 0
    in_transit_quantity: float = 0
    price: float = 0
    supplier: Optional[str] = None


class SparePartCreate(SparePartBase):
    pass


class SparePartUpdate(BaseModel):
    name: Optional[str] = None
    code: Optional[str] = None
    category: Optional[str] = None
    unit: Optional[str] = None
    stock_quantity: Optional[float] = None
    safety_stock: Optional[float] = None
    in_transit_quantity: Optional[float] = None
    price: Optional[float] = None
    supplier: Optional[str] = None


class SparePartResponse(SparePartBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class SpareUsageBase(BaseModel):
    lighthouse_id: int
    spare_part_id: int
    work_order_id: Optional[int] = None
    quantity: float
    usage_type: str = "maintenance"
    notes: Optional[str] = None


class SpareUsageCreate(SpareUsageBase):
    pass


class SpareUsageResponse(SpareUsageBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class PurchaseRequestItemBase(BaseModel):
    spare_part_id: int
    quantity: float
    unit_price: float = 0


class PurchaseRequestItemCreate(PurchaseRequestItemBase):
    pass


class PurchaseRequestItemResponse(PurchaseRequestItemBase):
    id: int
    subtotal: float = 0
    spare_part: Optional[SparePartResponse] = None

    class Config:
        from_attributes = True


class PurchaseRequestBase(BaseModel):
    requested_by: str
    reason: Optional[str] = None
    items: List[PurchaseRequestItemCreate]


class PurchaseRequestCreate(PurchaseRequestBase):
    pass


class PurchaseRequestResponse(BaseModel):
    id: int
    request_no: str
    status: str
    requested_by: str
    requested_at: datetime
    reason: Optional[str] = None
    total_amount: float = 0
    approved_by: Optional[str] = None
    approved_at: Optional[datetime] = None
    ordered_by: Optional[str] = None
    ordered_at: Optional[datetime] = None
    received_by: Optional[str] = None
    received_at: Optional[datetime] = None
    items: List[PurchaseRequestItemResponse] = []

    class Config:
        from_attributes = True


class PurchaseApprove(BaseModel):
    approved_by: str


class PurchaseOrder(BaseModel):
    ordered_by: str


class PurchaseReceive(BaseModel):
    received_by: str
