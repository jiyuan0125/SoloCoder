from datetime import date, datetime
from enum import Enum
from typing import List, Optional
from pydantic import BaseModel, Field, validator


class PurchaseRequestStatus(str, Enum):
    PENDING = "pending"
    APPROVED = "approved"
    REJECTED = "rejected"
    COMPLETED = "completed"


class Store(BaseModel):
    id: str
    name: str
    address: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class Ingredient(BaseModel):
    id: str
    name: str
    unit: str
    safety_stock: float = Field(gt=0)

    @validator("safety_stock")
    def validate_safety_stock(cls, v):
        if v <= 0:
            raise ValueError("安全库存线必须大于0")
        return v


class InventoryItem(BaseModel):
    ingredient_id: str
    ingredient_name: str
    quantity: float = Field(ge=0)
    unit_price: float = Field(ge=0)
    expiry_date: Optional[date] = None
    updated_at: datetime = Field(default_factory=datetime.now)
    is_low_stock: bool = False


class PurchaseRequest(BaseModel):
    id: str
    store_id: str
    ingredient_id: str
    ingredient_name: str
    requested_quantity: float = Field(gt=0)
    unit_price: float = Field(ge=0)
    expected_arrival_date: date
    status: PurchaseRequestStatus = PurchaseRequestStatus.PENDING
    created_at: datetime = Field(default_factory=datetime.now)
    approved_at: Optional[datetime] = None
    rejected_reason: Optional[str] = None

    @validator("expected_arrival_date")
    def validate_expected_arrival_date(cls, v):
        if v < date.today():
            raise ValueError("采购期望到货日期不能是过去的")
        return v


class PurchaseArrival(BaseModel):
    id: str
    purchase_request_id: str
    store_id: str
    ingredient_id: str
    ingredient_name: str
    arrived_quantity: float = Field(gt=0)
    unit_price: float = Field(ge=0)
    arrived_at: datetime = Field(default_factory=datetime.now)
    expiry_date: Optional[date] = None


class WasteRecord(BaseModel):
    id: str
    store_id: str
    ingredient_id: str
    ingredient_name: str
    requested_quantity: float = Field(gt=0)
    actual_deducted: float = Field(ge=0)
    unit_price: float = Field(ge=0)
    waste_date: date = Field(default_factory=date.today)
    reason: Optional[str] = None
    created_at: datetime = Field(default_factory=datetime.now)


class LowStockAlert(BaseModel):
    id: str
    store_id: str
    ingredient_id: str
    ingredient_name: str
    current_quantity: float
    safety_stock: float
    alert_date: date = Field(default_factory=date.today)
    created_at: datetime = Field(default_factory=datetime.now)


class AggregationResult(BaseModel):
    store_id: str
    store_name: str
    start_date: date
    end_date: date
    total_purchase_amount: float
    total_waste_amount: float
    waste_rate: float
    is_high_waste_rate: bool


class StoreRanking(BaseModel):
    rank: int
    store_id: str
    store_name: str
    waste_rate: float
    is_high_waste_rate: bool
