from datetime import datetime
from decimal import Decimal
from typing import Optional
from uuid import UUID

from pydantic import Field

from shared.models.base import BaseModel, TimestampMixin, UUIDMixin
from shared.models.enums import PurchaseStatus


class PurchaseItem(BaseModel):
    product_name: str = Field(..., min_length=1, max_length=200, description="商品名称")
    product_code: Optional[str] = Field(default=None, max_length=100, description="商品编码")
    quantity: int = Field(..., gt=0, description="采购数量")
    unit: str = Field(..., min_length=1, max_length=20, description="单位")
    estimated_price: Optional[Decimal] = Field(
        default=None, ge=Decimal("0"), decimal_places=4, description="预估单价"
    )
    description: Optional[str] = Field(default=None, max_length=500, description="商品描述")


class PurchaseRequirement(UUIDMixin, TimestampMixin):
    title: str = Field(..., min_length=1, max_length=200, description="采购需求标题")
    description: Optional[str] = Field(default=None, max_length=2000, description="需求描述")
    items: list[PurchaseItem] = Field(..., min_length=1, description="采购商品列表")
    status: PurchaseStatus = Field(default=PurchaseStatus.DRAFT, description="状态")
    quote_deadline: datetime = Field(..., description="报价截止时间")
    published_at: Optional[datetime] = Field(default=None, description="发布时间")
    awarded_at: Optional[datetime] = Field(default=None, description="定标时间")
    awarded_supplier_id: Optional[UUID] = Field(default=None, description="中标供应商ID")
