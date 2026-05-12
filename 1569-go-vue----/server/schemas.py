import datetime
from typing import List, Optional
from pydantic import BaseModel, Field
from server.models import (
    ProductCategory,
    SaleStatus,
    RefundStatus,
    RefundType,
    SettlementStatus,
)


class ShopBase(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    location: str = Field(min_length=1, max_length=200)
    manager_name: Optional[str] = Field(None, max_length=100)
    phone: Optional[str] = Field(None, max_length=20)


class ShopCreate(ShopBase):
    pass


class ShopUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=100)
    location: Optional[str] = Field(None, min_length=1, max_length=200)
    manager_name: Optional[str] = Field(None, max_length=100)
    phone: Optional[str] = Field(None, max_length=20)


class ShopResponse(ShopBase):
    id: int
    created_at: datetime.datetime
    updated_at: datetime.datetime

    class Config:
        from_attributes = True


class ProductBase(BaseModel):
    name: str = Field(min_length=1, max_length=150)
    sku: str = Field(min_length=1, max_length=50)
    brand: Optional[str] = Field(None, max_length=100)
    category: ProductCategory = ProductCategory.OTHER
    retail_price: int = Field(ge=0)
    duty_free_price: int = Field(ge=0)
    description: Optional[str] = None
    is_active: bool = True


class ProductCreate(ProductBase):
    shop_id: int


class ProductUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=150)
    brand: Optional[str] = Field(None, max_length=100)
    category: Optional[ProductCategory] = None
    retail_price: Optional[int] = Field(None, ge=0)
    duty_free_price: Optional[int] = Field(None, ge=0)
    description: Optional[str] = None
    is_active: Optional[bool] = None


class ProductResponse(ProductBase):
    id: int
    shop_id: int
    created_at: datetime.datetime
    updated_at: datetime.datetime

    class Config:
        from_attributes = True


class InventoryBatchBase(BaseModel):
    product_id: int
    batch_number: str = Field(min_length=1, max_length=50)
    quantity: int = Field(gt=0)
    expiry_date: datetime.date


class InventoryBatchCreate(InventoryBatchBase):
    pass


class InventoryBatchResponse(InventoryBatchBase):
    id: int
    available_quantity: int
    received_date: datetime.datetime
    created_at: datetime.datetime

    class Config:
        from_attributes = True


class ProductWithInventory(ProductResponse):
    total_available: int = 0
    is_expired: bool = False
    is_near_expiry: bool = False
    batches: List[InventoryBatchResponse] = []


class PassengerBase(BaseModel):
    name: str = Field(min_length=1, max_length=100)
    passport_number: str = Field(min_length=1, max_length=50)
    flight_number: str = Field(min_length=1, max_length=20)
    flight_departure_time: datetime.datetime


class PassengerCreate(PassengerBase):
    pass


class PassengerResponse(PassengerBase):
    id: int
    created_at: datetime.datetime

    class Config:
        from_attributes = True


class SaleItemCreate(BaseModel):
    product_id: int
    quantity: int = Field(gt=0)


class SaleCreate(BaseModel):
    shop_id: int
    passenger: PassengerCreate
    items: List[SaleItemCreate]
    payment_method: str = Field(min_length=1, max_length=50)


class SaleItemBatchResponse(BaseModel):
    id: int
    batch_id: int
    quantity: int
    is_returned: bool
    returned_quantity: int

    class Config:
        from_attributes = True


class SaleItemResponse(BaseModel):
    id: int
    product_id: int
    product_name: Optional[str] = None
    product_sku: Optional[str] = None
    product_category: Optional[ProductCategory] = None
    quantity: int
    unit_price: int
    subtotal: int
    batches: List[SaleItemBatchResponse] = []

    class Config:
        from_attributes = True


class SaleResponse(BaseModel):
    id: int
    shop_id: int
    shop_name: Optional[str] = None
    passenger_id: int
    passenger_name: Optional[str] = None
    passport_number: Optional[str] = None
    flight_number: Optional[str] = None
    flight_departure_time: Optional[datetime.datetime] = None
    sale_number: str
    total_amount: int
    payment_method: str
    status: SaleStatus
    sale_time: datetime.datetime
    created_at: datetime.datetime
    items: List[SaleItemResponse] = []

    class Config:
        from_attributes = True


class SettlementSummaryResponse(BaseModel):
    id: int
    payment_method: str
    total_amount: int
    transaction_count: int

    class Config:
        from_attributes = True


class DailySettlementResponse(BaseModel):
    id: int
    shop_id: int
    shop_name: Optional[str] = None
    settlement_date: datetime.date
    status: SettlementStatus
    created_at: datetime.datetime
    confirmed_at: Optional[datetime.datetime] = None
    summaries: List[SettlementSummaryResponse] = []

    class Config:
        from_attributes = True


class RefundItemResponse(BaseModel):
    id: int
    sale_item_id: int
    product_name: Optional[str] = None
    quantity: int
    unit_price: int
    subtotal: int

    class Config:
        from_attributes = True


class RefundRequestResponse(BaseModel):
    id: int
    sale_id: int
    sale_number: Optional[str] = None
    refund_number: str
    refund_type: RefundType
    status: RefundStatus
    reason: Optional[str] = None
    is_sealed: bool
    total_refund_amount: int
    applicant_name: Optional[str] = None
    supervisor_name: Optional[str] = None
    applied_at: datetime.datetime
    approved_at: Optional[datetime.datetime] = None
    rejected_at: Optional[datetime.datetime] = None
    refunded_at: Optional[datetime.datetime] = None
    completed_at: Optional[datetime.datetime] = None
    rejection_reason: Optional[str] = None
    items: List[RefundItemResponse] = []

    class Config:
        from_attributes = True


class RefundRequestCreate(BaseModel):
    sale_id: int
    reason: Optional[str] = None
    is_sealed: bool = True
    applicant_name: Optional[str] = None


class RefundApproval(BaseModel):
    supervisor_name: str = Field(min_length=1, max_length=100)
