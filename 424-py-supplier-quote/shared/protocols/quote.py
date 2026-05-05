from datetime import datetime
from decimal import Decimal
from typing import Optional
from uuid import UUID

from pydantic import Field

from shared.models.base import BaseModel
from shared.models.enums import QuoteStatus, QualificationLevel
from shared.models.quote import Quote, QuoteHistory, QuoteVersion


class QuoteSubmitRequest(BaseModel):
    unit_price: Decimal = Field(..., ge=Decimal("0"), decimal_places=4, description="单价")
    delivery_days: int = Field(..., ge=0, description="承诺交货天数")
    remarks: str = Field(default="", max_length=1000, description="备注")


class QuoteUpdateRequest(BaseModel):
    unit_price: Optional[Decimal] = Field(default=None, ge=Decimal("0"), decimal_places=4, description="单价")
    delivery_days: Optional[int] = Field(default=None, ge=0, description="承诺交货天数")
    remarks: Optional[str] = Field(default=None, max_length=1000, description="备注")


class QuoteResponse(Quote):
    supplier_name: str = Field(..., description="供应商名称")
    qualification_level: QualificationLevel = Field(..., description="供应商资质等级")


class QuoteListResponse(BaseModel):
    quotes: list[QuoteResponse] = Field(..., description="报价列表")
    total: int = Field(..., ge=0, description="总数量")


class QuoteHistoryResponse(BaseModel):
    quote_id: UUID = Field(..., description="报价ID")
    versions: list[QuoteVersion] = Field(..., description="版本列表")
    history: list[QuoteHistory] = Field(..., description="变更历史")


class QuoteComparisonItem(BaseModel):
    supplier_id: UUID = Field(..., description="供应商ID")
    supplier_name: str = Field(..., description="供应商名称")
    qualification_level: QualificationLevel = Field(..., description="资质等级")
    unit_price: Decimal = Field(..., ge=Decimal("0"), decimal_places=4, description="报价单价")
    delivery_days: int = Field(..., ge=0, description="承诺交货天数")
    rank: int = Field(default=1, ge=1, description="排名")
    is_late: bool = Field(default=False, description="是否为迟到报价")
    is_awarded: bool = Field(default=False, description="是否中标")


class QuoteComparisonReport(BaseModel):
    purchase_id: UUID = Field(..., description="采购需求ID")
    purchase_title: str = Field(..., description="采购需求标题")
    product_name: str = Field(..., description="商品名称")
    quotes: list[QuoteComparisonItem] = Field(..., description="报价列表")
    price_distribution: dict[str, int] = Field(..., description="价格分布统计")
    avg_price: Decimal = Field(..., ge=Decimal("0"), decimal_places=4, description="平均价格")
    min_price: Decimal = Field(..., ge=Decimal("0"), decimal_places=4, description="最低价格")
    max_price: Decimal = Field(..., ge=Decimal("0"), decimal_places=4, description="最高价格")


class AwardResultMasked(BaseModel):
    purchase_id: UUID = Field(..., description="采购需求ID")
    purchase_title: str = Field(..., description="采购需求标题")
    is_awarded: bool = Field(..., description="是否中标")
    awarded_supplier_name: str = Field(..., description="中标供应商名称")
    price_comparison: str = Field(..., description="价格比较（higher/lower/same）")
