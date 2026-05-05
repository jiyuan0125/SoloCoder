from datetime import datetime
from typing import Optional
from uuid import UUID

from pydantic import Field

from shared.models.base import BaseModel
from shared.models.enums import PurchaseStatus
from shared.models.purchase import PurchaseItem, PurchaseRequirement


class PurchaseCreateRequest(BaseModel):
    title: str = Field(..., min_length=1, max_length=200, description="采购需求标题")
    description: Optional[str] = Field(default=None, max_length=2000, description="需求描述")
    items: list[PurchaseItem] = Field(..., min_length=1, description="采购商品列表")
    quote_deadline: datetime = Field(..., description="报价截止时间")


class PurchaseUpdateRequest(BaseModel):
    title: Optional[str] = Field(default=None, min_length=1, max_length=200, description="采购需求标题")
    description: Optional[str] = Field(default=None, max_length=2000, description="需求描述")
    items: Optional[list[PurchaseItem]] = Field(default=None, min_length=1, description="采购商品列表")
    quote_deadline: Optional[datetime] = Field(default=None, description="报价截止时间")


class PurchasePublishRequest(BaseModel):
    pass


class PurchaseResponse(PurchaseRequirement):
    pass


class PurchaseListResponse(BaseModel):
    purchases: list[PurchaseResponse] = Field(..., description="采购需求列表")
    total: int = Field(..., ge=0, description="总数量")
