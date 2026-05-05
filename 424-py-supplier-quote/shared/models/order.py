from datetime import datetime
from decimal import Decimal
from typing import Optional
from uuid import UUID

from pydantic import Field

from shared.models.base import BaseModel, TimestampMixin, UUIDMixin
from shared.models.enums import OrderStatus


class OrderItem(BaseModel):
    product_name: str = Field(..., min_length=1, max_length=200, description="商品名称")
    product_code: Optional[str] = Field(default=None, max_length=100, description="商品编码")
    quantity: int = Field(..., gt=0, description="采购数量")
    unit: str = Field(..., min_length=1, max_length=20, description="单位")
    unit_price: Decimal = Field(..., ge=Decimal("0"), decimal_places=4, description="单价")
    total_amount: Decimal = Field(..., ge=Decimal("0"), decimal_places=4, description="小计金额")


class PurchaseOrder(UUIDMixin, TimestampMixin):
    order_number: str = Field(..., min_length=1, max_length=50, description="订单编号")
    purchase_id: UUID = Field(..., description="关联采购需求ID")
    quote_id: UUID = Field(..., description="关联报价ID")
    supplier_id: UUID = Field(..., description="供应商ID")

    items: list[OrderItem] = Field(..., min_length=1, description="订单商品列表")
    total_amount: Decimal = Field(..., ge=Decimal("0"), decimal_places=4, description="订单总金额")

    status: OrderStatus = Field(default=OrderStatus.DRAFT, description="订单状态")
    delivery_days: int = Field(..., ge=0, description="承诺交货天数")
    expected_delivery_date: Optional[datetime] = Field(default=None, description="预计交货日期")

    confirmed_at: Optional[datetime] = Field(default=None, description="确认时间")
    completed_at: Optional[datetime] = Field(default=None, description="完成时间")
    remarks: str = Field(default="", max_length=1000, description="备注")
