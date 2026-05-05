from enum import Enum
from decimal import Decimal
from datetime import datetime
from typing import Optional, Union
from pydantic import BaseModel, Field, field_validator


class LocationStatus(str, Enum):
    IDLE = "idle"
    OCCUPIED = "occupied"
    LOCKED = "locked"
    UNDER_INVENTORY = "under_inventory"
    UNDER_MAINTENANCE = "under_maintenance"


class ErrorCode(str, Enum):
    SUCCESS = "success"
    INVALID_LOCATION_CODE = "invalid_location_code"
    LOCATION_NOT_FOUND = "location_not_found"
    PRODUCT_NOT_FOUND = "product_not_found"
    LOCATION_OCCUPIED = "location_occupied"
    LOCATION_NOT_OCCUPIED = "location_not_occupied"
    LOCATION_LOCKED = "location_locked"
    LOCATION_UNDER_MAINTENANCE = "location_under_maintenance"
    LOCATION_UNDER_INVENTORY = "location_under_inventory"
    INSUFFICIENT_CAPACITY = "insufficient_capacity"
    INSUFFICIENT_QUANTITY = "insufficient_quantity"
    PRODUCT_MISMATCH = "product_mismatch"
    OVER_CAPACITY = "over_capacity"
    INVENTORY_ALREADY_STARTED = "inventory_already_started"
    INVENTORY_NOT_STARTED = "inventory_not_started"
    ZONE_NOT_FOUND = "zone_not_found"
    SHELF_NOT_FOUND = "shelf_not_found"
    ZONE_ALREADY_EXISTS = "zone_already_exists"
    SHELF_ALREADY_EXISTS = "shelf_already_exists"
    PRODUCT_ALREADY_EXISTS = "product_already_exists"
    INVALID_REQUEST = "invalid_request"
    INTERNAL_ERROR = "internal_error"


ERROR_CODE_DESCRIPTIONS: dict[ErrorCode, str] = {
    ErrorCode.SUCCESS: "操作成功",
    ErrorCode.INVALID_LOCATION_CODE: "无效的库位编码格式",
    ErrorCode.LOCATION_NOT_FOUND: "库位不存在",
    ErrorCode.PRODUCT_NOT_FOUND: "商品不存在",
    ErrorCode.LOCATION_OCCUPIED: "库位已被占用",
    ErrorCode.LOCATION_NOT_OCCUPIED: "库位未被占用",
    ErrorCode.LOCATION_LOCKED: "库位已锁定，无法进行操作",
    ErrorCode.LOCATION_UNDER_MAINTENANCE: "库位正在维护中，无法进行出入库操作",
    ErrorCode.LOCATION_UNDER_INVENTORY: "库位正在盘点中，无法进行出入库操作",
    ErrorCode.INSUFFICIENT_CAPACITY: "库位容量不足",
    ErrorCode.INSUFFICIENT_QUANTITY: "库存数量不足",
    ErrorCode.PRODUCT_MISMATCH: "库位已存放其他种类商品",
    ErrorCode.OVER_CAPACITY: "超出库位容量上限",
    ErrorCode.INVENTORY_ALREADY_STARTED: "盘点已在进行中",
    ErrorCode.INVENTORY_NOT_STARTED: "盘点会话不存在或未开始",
    ErrorCode.ZONE_NOT_FOUND: "区域不存在",
    ErrorCode.SHELF_NOT_FOUND: "货架不存在",
    ErrorCode.ZONE_ALREADY_EXISTS: "区域已存在",
    ErrorCode.SHELF_ALREADY_EXISTS: "货架已存在",
    ErrorCode.PRODUCT_ALREADY_EXISTS: "商品已存在",
    ErrorCode.INVALID_REQUEST: "请求参数无效",
    ErrorCode.INTERNAL_ERROR: "服务器内部错误",
}


def get_error_description(error_code: ErrorCode | None) -> str:
    if error_code is None:
        return "未知错误"
    return ERROR_CODE_DESCRIPTIONS.get(error_code, str(error_code))


class Zone(BaseModel):
    code: str = Field(..., min_length=1, max_length=5, pattern=r"^[A-Za-z]+$")
    name: str = Field(..., min_length=1, max_length=100)
    description: Optional[str] = Field(None, max_length=500)

    @field_validator("code")
    @classmethod
    def uppercase_code(cls, v: str) -> str:
        return v.upper()


class Shelf(BaseModel):
    zone_code: str = Field(..., min_length=1, max_length=5)
    shelf_number: int = Field(..., ge=1)
    name: Optional[str] = Field(None, max_length=100)
    total_locations: int = Field(..., ge=1)
    layers: int = Field(..., ge=1)
    columns: int = Field(..., ge=1)


class Product(BaseModel):
    sku: str = Field(..., min_length=1, max_length=50)
    name: str = Field(..., min_length=1, max_length=200)
    description: Optional[str] = Field(None, max_length=500)
    unit_price: Decimal = Field(..., gt=Decimal("0"), decimal_places=4)
    volume: Decimal = Field(..., gt=Decimal("0"), decimal_places=4)


class Location(BaseModel):
    code: str = Field(..., min_length=1, max_length=20)
    zone_code: str = Field(..., min_length=1, max_length=5)
    shelf_number: int = Field(..., ge=1)
    layer: int = Field(..., ge=1)
    column: int = Field(..., ge=1)
    status: LocationStatus = Field(default=LocationStatus.IDLE)
    product_sku: Optional[str] = Field(None, max_length=50)
    quantity: int = Field(default=0, ge=0)
    max_capacity: int = Field(..., ge=1)
    volume_capacity: Decimal = Field(default=Decimal("1000"), gt=Decimal("0"), decimal_places=4)
    first_in_time: Optional[datetime] = Field(None)
    last_in_time: Optional[datetime] = Field(None)

    @classmethod
    def generate_code(
        cls, zone_code: str, shelf_number: int, layer: int, column: int
    ) -> str:
        return f"{zone_code}-{shelf_number:02d}-{layer:02d}-{column:02d}"

    @field_validator("code")
    @classmethod
    def validate_code_format(cls, v: str) -> str:
        parts = v.split("-")
        if len(parts) != 4:
            raise ValueError("Location code must be in format: ZONE-SHELF-LAYER-COLUMN")
        return v


class InventoryRecord(BaseModel):
    id: str = Field(..., min_length=1, max_length=50)
    location_code: str = Field(..., min_length=1, max_length=20)
    product_sku: Optional[str] = Field(None, max_length=50)
    expected_quantity: int = Field(..., ge=0)
    actual_quantity: int = Field(..., ge=0)
    difference: int = Field(default=0)
    created_at: datetime = Field(default_factory=datetime.now)
    completed_at: Optional[datetime] = Field(None)
    notes: Optional[str] = Field(None, max_length=500)


class InventorySession(BaseModel):
    id: str = Field(..., min_length=1, max_length=50)
    zone_code: Optional[str] = Field(None, max_length=5)
    location_codes: list[str] = Field(default_factory=list)
    status: str = Field(default="pending")
    created_at: datetime = Field(default_factory=datetime.now)
    completed_at: Optional[datetime] = Field(None)
    records: list[InventoryRecord] = Field(default_factory=list)


class StockInRequest(BaseModel):
    product_sku: str = Field(..., min_length=1, max_length=50)
    quantity: int = Field(..., gt=0)
    preferred_zone: Optional[str] = Field(None, max_length=5)
    location_code: Optional[str] = Field(None, max_length=20)


class StockInResult(BaseModel):
    product_sku: str = Field(..., min_length=1, max_length=50)
    location_code: str = Field(..., min_length=1, max_length=20)
    quantity: int = Field(..., gt=0)
    allocated_at: datetime = Field(default_factory=datetime.now)


class StockOutRequest(BaseModel):
    product_sku: str = Field(..., min_length=1, max_length=50)
    quantity: int = Field(..., gt=0)


class StockOutResult(BaseModel):
    product_sku: str = Field(..., min_length=1, max_length=50)
    location_code: str = Field(..., min_length=1, max_length=20)
    quantity: int = Field(..., gt=0)
    released_at: datetime = Field(default_factory=datetime.now)


class AllocationStrategy(str, Enum):
    ADJACENT_PREFERRED = "adjacent_preferred"
    ANY_AVAILABLE = "any_available"


class AlertLevel(str, Enum):
    INFO = "info"
    WARNING = "warning"
    CRITICAL = "critical"


class Alert(BaseModel):
    id: str = Field(..., min_length=1, max_length=50)
    level: AlertLevel = Field(...)
    message: str = Field(..., min_length=1, max_length=500)
    location_code: Optional[str] = Field(None, max_length=20)
    product_sku: Optional[str] = Field(None, max_length=50)
    created_at: datetime = Field(default_factory=datetime.now)
    resolved: bool = Field(default=False)
    resolved_at: Optional[datetime] = Field(None)


class ApiResponse[T](BaseModel):
    success: bool = Field(default=True)
    error_code: Optional[ErrorCode] = Field(None)
    error_message: Optional[str] = Field(None, max_length=500)
    data: Optional[T] = Field(None)


class LocationQueryResult(BaseModel):
    location: Location = Field(...)
    product: Optional[Product] = Field(None)


class ProductLocationsResult(BaseModel):
    product: Product = Field(...)
    locations: list[Location] = Field(default_factory=list)
    total_quantity: int = Field(default=0)


class ZoneCreateRequest(BaseModel):
    code: str = Field(..., min_length=1, max_length=5, pattern=r"^[A-Za-z]+$")
    name: str = Field(..., min_length=1, max_length=100)
    description: Optional[str] = Field(None, max_length=500)


class ShelfCreateRequest(BaseModel):
    zone_code: str = Field(..., min_length=1, max_length=5)
    shelf_number: int = Field(..., ge=1)
    name: Optional[str] = Field(None, max_length=100)
    layers: int = Field(..., ge=1)
    columns: int = Field(..., ge=1)
    default_capacity_per_location: int = Field(default=100, ge=1)
    default_volume_capacity_per_location: Decimal = Field(default=Decimal("1000"), gt=Decimal("0"), decimal_places=4)


class ProductCreateRequest(BaseModel):
    sku: str = Field(..., min_length=1, max_length=50)
    name: str = Field(..., min_length=1, max_length=200)
    description: Optional[str] = Field(None, max_length=500)
    unit_price: Decimal = Field(..., gt=Decimal("0"), decimal_places=4)
    volume: Decimal = Field(..., gt=Decimal("0"), decimal_places=4)


class InventoryStartRequest(BaseModel):
    zone_code: Optional[str] = Field(None, max_length=5)
    location_codes: Optional[list[str]] = Field(None)
    include_all: bool = Field(default=False)


class InventoryItem(BaseModel):
    location_code: str = Field(..., min_length=1, max_length=20)
    actual_quantity: int = Field(..., ge=0)


class InventoryCompleteRequest(BaseModel):
    session_id: str = Field(..., min_length=1, max_length=50)
    items: list[InventoryItem] = Field(...)


class LocationUpdateStatusRequest(BaseModel):
    status: LocationStatus = Field(...)


class LocationSetCapacityRequest(BaseModel):
    max_capacity: int = Field(..., ge=1)


class ZoneDetail(BaseModel):
    zone: Zone = Field(...)
    shelf_count: int = Field(default=0)
    location_count: int = Field(default=0)


class ShelfDetail(BaseModel):
    shelf: Shelf = Field(...)
    zone: Zone = Field(...)
    locations: list[Location] = Field(default_factory=list)
