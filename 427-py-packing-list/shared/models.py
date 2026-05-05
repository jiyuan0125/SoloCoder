from datetime import datetime
from enum import Enum
from typing import Optional, List, Dict, Any
from uuid import UUID, uuid4
from pydantic import BaseModel, Field, field_validator


class ProductType(str, Enum):
    NORMAL = "normal"
    FRAGILE = "fragile"
    LIQUID = "liquid"


class PackageStatus(str, Enum):
    PENDING = "pending"
    PACKED = "packed"
    SHIPPED = "shipped"
    NEED_REPACK = "need_repack"


class OrderStatus(str, Enum):
    CREATED = "created"
    PACKING = "packing"
    PACKED = "packed"
    SHIPPED = "shipped"
    MODIFIED = "modified"


class Product(BaseModel):
    product_id: str = Field(..., min_length=1, max_length=64)
    name: str = Field(..., min_length=1, max_length=256)
    weight_kg: float = Field(..., gt=0, le=1000)
    length_cm: float = Field(..., gt=0, le=500)
    width_cm: float = Field(..., gt=0, le=500)
    height_cm: float = Field(..., gt=0, le=500)
    product_type: ProductType = Field(default=ProductType.NORMAL)

    @property
    def volume_m3(self) -> float:
        return (self.length_cm * self.width_cm * self.height_cm) / 1_000_000


class OrderItem(BaseModel):
    product: Product
    quantity: int = Field(..., ge=1)


class Order(BaseModel):
    order_id: str = Field(..., min_length=1, max_length=64)
    items: List[OrderItem] = Field(..., min_length=1)
    status: OrderStatus = Field(default=OrderStatus.CREATED)
    created_at: datetime = Field(default_factory=datetime.now)
    modified_at: Optional[datetime] = Field(default=None)


class PackageItem(BaseModel):
    product: Product
    quantity: int = Field(..., ge=1)
    position_layer: int = Field(..., ge=1)


class BoxType(BaseModel):
    box_type_id: str = Field(..., min_length=1, max_length=64)
    name: str = Field(..., min_length=1, max_length=128)
    length_cm: float = Field(..., gt=0)
    width_cm: float = Field(..., gt=0)
    height_cm: float = Field(..., gt=0)
    max_weight_kg: float = Field(default=30.0, gt=0)

    @property
    def volume_m3(self) -> float:
        return (self.length_cm * self.width_cm * self.height_cm) / 1_000_000

    @property
    def effective_volume_m3(self) -> float:
        return self.volume_m3 * 0.8


class Package(BaseModel):
    package_id: UUID = Field(default_factory=uuid4)
    box_number: str = Field(..., min_length=1, max_length=64)
    box_type: BoxType
    items: List[PackageItem] = Field(default_factory=list)
    status: PackageStatus = Field(default=PackageStatus.PENDING)
    tracking_number: Optional[str] = Field(default=None, min_length=1, max_length=128)
    has_fragile_label: bool = Field(default=False)
    shipped_at: Optional[datetime] = Field(default=None)
    created_at: datetime = Field(default_factory=datetime.now)

    @property
    def total_weight_kg(self) -> float:
        return sum(item.product.weight_kg * item.quantity for item in self.items)

    @property
    def total_volume_m3(self) -> float:
        return sum(item.product.volume_m3 * item.quantity for item in self.items)

    @property
    def filler_volume_m3(self) -> float:
        effective_volume = self.box_type.effective_volume_m3
        return max(0.0, effective_volume - self.total_volume_m3)


class PackingList(BaseModel):
    packing_id: UUID = Field(default_factory=uuid4)
    order_id: str = Field(..., min_length=1, max_length=64)
    packages: List[Package] = Field(default_factory=list)
    created_at: datetime = Field(default_factory=datetime.now)
    is_template: bool = Field(default=False)
    template_name: Optional[str] = Field(default=None, min_length=1, max_length=128)

    @property
    def total_boxes(self) -> int:
        return len(self.packages)

    @property
    def total_weight_kg(self) -> float:
        return sum(pkg.total_weight_kg for pkg in self.packages)

    @property
    def total_volume_m3(self) -> float:
        return sum(pkg.total_volume_m3 for pkg in self.packages)


class PackagingMaterialInventory(BaseModel):
    box_type: BoxType
    stock_quantity: int = Field(..., ge=0)
    used_quantity: int = Field(default=0, ge=0)


class PackingSuggestion(BaseModel):
    suggested_box_type: BoxType
    items: List[OrderItem]
    estimated_weight_kg: float
    estimated_volume_m3: float
    confidence_score: float = Field(..., ge=0, le=1)


class CreateOrderRequest(BaseModel):
    order_id: str = Field(..., min_length=1, max_length=64)
    items: List[OrderItem] = Field(..., min_length=1)


class UpdateOrderRequest(BaseModel):
    items: Optional[List[OrderItem]] = Field(default=None)


class CreatePackageRequest(BaseModel):
    box_type_id: str = Field(..., min_length=1, max_length=64)


class AddItemToPackageRequest(BaseModel):
    product_id: str = Field(..., min_length=1, max_length=64)
    quantity: int = Field(..., ge=1)


class UpdateTrackingNumberRequest(BaseModel):
    tracking_number: str = Field(..., min_length=1, max_length=128)


class SavePackingTemplateRequest(BaseModel):
    template_name: str = Field(..., min_length=1, max_length=128)


class ApplyPackingTemplateRequest(BaseModel):
    template_id: UUID


class ApiResponse(BaseModel):
    success: bool
    data: Optional[Dict[str, Any]] = Field(default=None)
    error_code: Optional[str] = Field(default=None)
    message: Optional[str] = Field(default=None)


class BoxTypeCreateRequest(BaseModel):
    box_type_id: str = Field(..., min_length=1, max_length=64)
    name: str = Field(..., min_length=1, max_length=128)
    length_cm: float = Field(..., gt=0)
    width_cm: float = Field(..., gt=0)
    height_cm: float = Field(..., gt=0)
    max_weight_kg: float = Field(default=30.0, gt=0)
    initial_stock: int = Field(default=100, ge=0)
