from datetime import datetime
from decimal import Decimal
from typing import Optional
from uuid import UUID

from pydantic import Field

from shared.models.base import BaseModel
from shared.models.enums import OrderStatus
from shared.models.order import OrderItem, PurchaseOrder


class OrderResponse(PurchaseOrder):
    supplier_name: str = Field(..., description="供应商名称")


class OrderListResponse(BaseModel):
    orders: list[OrderResponse] = Field(..., description="订单列表")
    total: int = Field(..., ge=0, description="总数量")


class OrderConfirmRequest(BaseModel):
    remarks: str = Field(default="", max_length=1000, description="备注")
