from datetime import datetime
from enum import Enum, IntEnum
from typing import Any, Dict, Generic, List, Optional, TypeVar

from pydantic import BaseModel, Field, field_validator, ValidationInfo

T = TypeVar("T")


class AlertType(str, Enum):
    LOW_STOCK = "low_stock"
    OVERSTOCK = "overstock"


class AlertLevel(str, Enum):
    WARNING = "warning"
    CRITICAL = "critical"


class TransactionType(str, Enum):
    INBOUND = "inbound"
    OUTBOUND = "outbound"
    TRANSFER = "transfer"


class InventoryStatus(str, Enum):
    NORMAL = "normal"
    LOW_STOCK = "low_stock"
    OVERSTOCK = "overstock"


class Month(IntEnum):
    JANUARY = 1
    FEBRUARY = 2
    MARCH = 3
    APRIL = 4
    MAY = 5
    JUNE = 6
    JULY = 7
    AUGUST = 8
    SEPTEMBER = 9
    OCTOBER = 10
    NOVEMBER = 11
    DECEMBER = 12


class SeasonalSafetyStock(BaseModel):
    month: Month
    min_stock: int = Field(ge=0)
    max_stock: int = Field(ge=0)

    @field_validator("max_stock")
    @classmethod
    def check_max_stock(cls, v: int, info: ValidationInfo) -> int:
        if "min_stock" in info.data and v <= info.data["min_stock"]:
            raise ValueError("max_stock must be greater than min_stock")
        return v


class SafetyStock(BaseModel):
    is_seasonal: bool = False
    min_stock: int = Field(default=0, ge=0)
    max_stock: int = Field(default=1000, ge=0)
    seasonal_config: Optional[Dict[Month, SeasonalSafetyStock]] = None

    @field_validator("max_stock")
    @classmethod
    def check_max_stock(cls, v: int, info: ValidationInfo) -> int:
        if not info.data.get("is_seasonal") and "min_stock" in info.data:
            if v <= info.data["min_stock"]:
                raise ValueError("max_stock must be greater than min_stock")
        return v

    def get_current_safety_stock(self, current_month: int) -> SeasonalSafetyStock:
        if self.is_seasonal and self.seasonal_config:
            month_enum = Month(current_month)
            if month_enum in self.seasonal_config:
                return self.seasonal_config[month_enum]
        return SeasonalSafetyStock(
            month=Month(current_month),
            min_stock=self.min_stock,
            max_stock=self.max_stock,
        )


class ProductCreate(BaseModel):
    sku: str = Field(min_length=1, max_length=50)
    name: str = Field(min_length=1, max_length=200)
    category: str = Field(min_length=1, max_length=100)
    initial_stock: int = Field(default=0, ge=0)
    safety_stock: Optional[SafetyStock] = None


class ProductUpdate(BaseModel):
    name: Optional[str] = Field(None, min_length=1, max_length=200)
    category: Optional[str] = Field(None, min_length=1, max_length=100)


class Product(BaseModel):
    sku: str
    name: str
    category: str
    current_stock: int
    safety_stock: SafetyStock
    created_at: datetime
    updated_at: datetime

    def get_inventory_status(self, current_month: int) -> InventoryStatus:
        ss = self.safety_stock.get_current_safety_stock(current_month)
        if self.current_stock < ss.min_stock:
            return InventoryStatus.LOW_STOCK
        if self.current_stock > ss.max_stock:
            return InventoryStatus.OVERSTOCK
        return InventoryStatus.NORMAL


class AlertRecord(BaseModel):
    id: str = Field(min_length=1)
    sku: str
    alert_type: AlertType
    alert_level: AlertLevel
    current_stock: int
    min_stock: int
    max_stock: int
    suggestion_quantity: int
    triggered_at: datetime
    resolved: bool = False
    resolved_at: Optional[datetime] = None


class ReplenishmentOrderDraft(BaseModel):
    id: str = Field(min_length=1)
    sku: str
    alert_id: str
    suggested_quantity: int
    confirmed: bool = False
    created_at: datetime
    confirmed_at: Optional[datetime] = None


class InventoryTransaction(BaseModel):
    id: str = Field(min_length=1)
    sku: str
    transaction_type: TransactionType
    quantity: int
    previous_stock: int
    new_stock: int
    reference_id: Optional[str] = None
    notes: Optional[str] = None
    created_at: datetime


class InventoryHealthReportItem(BaseModel):
    sku: str
    name: str
    category: str
    current_stock: int
    min_stock: int
    max_stock: int
    status: InventoryStatus


class InventoryHealthReport(BaseModel):
    id: str = Field(min_length=1)
    generated_at: datetime
    total_products: int
    normal_count: int
    low_stock_count: int
    overstock_count: int
    items: List[InventoryHealthReportItem]


class TurnoverAnalysisItem(BaseModel):
    sku: str
    name: str
    category: str
    total_outbound: int
    average_stock: float
    turnover_rate: float
    turnover_days: float


class InventoryTurnoverAnalysis(BaseModel):
    id: str = Field(min_length=1)
    start_date: datetime
    end_date: datetime
    items: List[TurnoverAnalysisItem]


class APIResponse(BaseModel, Generic[T]):
    success: bool
    data: Optional[T] = None
    error: Optional[str] = None
    error_code: Optional[str] = None


class PaginationParams(BaseModel):
    page: int = Field(default=1, ge=1)
    page_size: int = Field(default=20, ge=1, le=100)


class AlertQueryParams(BaseModel):
    sku: Optional[str] = None
    category: Optional[str] = None
    alert_type: Optional[AlertType] = None
    alert_level: Optional[AlertLevel] = None
    resolved: Optional[bool] = None
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None


class BulkSafetyStockUpdate(BaseModel):
    category: str
    safety_stock: SafetyStock


class StockAdjustment(BaseModel):
    quantity: int = Field(ge=1)
    reference_id: Optional[str] = None
    notes: Optional[str] = None
