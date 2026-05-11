from datetime import date, datetime
from typing import Optional, List
from pydantic import BaseModel, Field, field_validator


class StoreBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    address: Optional[str] = None


class StoreCreate(StoreBase):
    pass


class StoreResponse(StoreBase):
    id: int
    created_at: datetime

    class Config:
        from_attributes = True


class IngredientBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=100)
    stock_quantity: float = Field(..., ge=0)
    unit_price: float = Field(..., ge=0)
    expiry_date: Optional[date] = None
    safety_stock: float = Field(..., ge=0)


class IngredientCreate(IngredientBase):
    store_id: int


class IngredientUpdate(BaseModel):
    name: Optional[str] = None
    stock_quantity: Optional[float] = Field(None, ge=0)
    unit_price: Optional[float] = Field(None, ge=0)
    expiry_date: Optional[date] = None
    safety_stock: Optional[float] = Field(None, ge=0)


class IngredientResponse(IngredientBase):
    id: int
    store_id: int
    is_low_stock: bool
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class PurchaseOrderBase(BaseModel):
    store_id: int
    ingredient_id: int
    requested_quantity: float = Field(..., gt=0)
    expected_arrival_date: date
    remarks: Optional[str] = None

    @field_validator("expected_arrival_date")
    @classmethod
    def validate_expected_arrival_date(cls, v: date) -> date:
        if v < date.today():
            raise ValueError("期望到货日期不能是过去的日期")
        return v


class PurchaseOrderCreate(PurchaseOrderBase):
    pass


class PurchaseOrderUpdate(BaseModel):
    expected_arrival_date: Optional[date] = None
    remarks: Optional[str] = None


class PurchaseOrderApprove(BaseModel):
    approved: bool


class PurchaseOrderReceive(BaseModel):
    received_quantity: float = Field(..., gt=0)
    actual_arrival_date: date = Field(default_factory=date.today)


class PurchaseOrderResponse(BaseModel):
    id: int
    store_id: int
    ingredient_id: int
    ingredient_name: Optional[str] = None
    requested_quantity: float
    received_quantity: Optional[float] = None
    expected_arrival_date: date
    actual_arrival_date: Optional[date] = None
    status: str
    remarks: Optional[str] = None
    created_at: datetime
    updated_at: datetime

    class Config:
        from_attributes = True


class WastageBase(BaseModel):
    store_id: int
    ingredient_id: int
    quantity: float = Field(..., gt=0)
    wastage_date: date = Field(default_factory=date.today)
    reason: Optional[str] = None


class WastageCreate(WastageBase):
    pass


class WastageResponse(BaseModel):
    id: int
    store_id: int
    ingredient_id: int
    ingredient_name: Optional[str] = None
    quantity: float
    actual_deducted: float
    wastage_date: date
    reason: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True


class LowStockAlertResponse(BaseModel):
    id: int
    store_id: int
    ingredient_id: int
    ingredient_name: Optional[str] = None
    alert_date: date
    current_stock: float
    safety_stock: float
    is_active: bool
    created_at: datetime

    class Config:
        from_attributes = True


class MetricsRequest(BaseModel):
    store_ids: Optional[List[int]] = None
    start_date: date
    end_date: date

    @field_validator("end_date")
    @classmethod
    def validate_date_range(cls, v: date, info) -> date:
        start_date = info.data.get("start_date")
        if start_date and v < start_date:
            raise ValueError("结束日期不能早于开始日期")
        return v


class StoreMetrics(BaseModel):
    store_id: int
    store_name: str
    purchase_total_amount: float
    wastage_total_amount: float
    wastage_rate: float
    wastage_rate_high: bool


class MetricsResponse(BaseModel):
    start_date: date
    end_date: date
    metrics: List[StoreMetrics]


class WastageRankingItem(BaseModel):
    store_id: int
    store_name: str
    wastage_rate: float
    wastage_rate_high: bool
    rank: int


class WastageRankingResponse(BaseModel):
    start_date: date
    end_date: date
    ranking: List[WastageRankingItem]


class ErrorResponse(BaseModel):
    detail: str
