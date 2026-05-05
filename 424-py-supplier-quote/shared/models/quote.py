from datetime import datetime, timedelta
from decimal import Decimal
from typing import Optional
from uuid import UUID

from pydantic import Field, field_validator

from shared.models.base import BaseModel, TimestampMixin, UUIDMixin
from shared.models.enums import QuoteStatus


QUOTE_VALIDITY_DAYS: int = 30


class QuoteVersion(BaseModel):
    version: int = Field(..., ge=1, description="版本号")
    unit_price: Decimal = Field(..., ge=Decimal("0"), decimal_places=4, description="单价")
    delivery_days: int = Field(..., ge=0, description="承诺交货天数")
    remarks: str = Field(default="", max_length=1000, description="备注")
    submitted_at: datetime = Field(..., description="提交时间")
    is_late: bool = Field(default=False, description="是否为迟到报价")


class QuoteHistory(BaseModel):
    version: int = Field(..., ge=1, description="版本号")
    changed_at: datetime = Field(..., description="修改时间")
    changes: dict[str, str] = Field(..., description="变更内容差异")


class Quote(UUIDMixin, TimestampMixin):
    purchase_id: UUID = Field(..., description="采购需求ID")
    supplier_id: UUID = Field(..., description="供应商ID")
    status: QuoteStatus = Field(default=QuoteStatus.DRAFT, description="报价状态")

    current_version: int = Field(default=1, ge=1, description="当前版本号")
    versions: list[QuoteVersion] = Field(default_factory=list, description="版本历史")
    history: list[QuoteHistory] = Field(default_factory=list, description="变更历史记录")

    valid_until: datetime = Field(..., description="报价有效期截止时间")

    @field_validator("valid_until", mode="before")
    @classmethod
    def set_default_valid_until(cls, v: Optional[datetime]) -> datetime:
        if v is None:
            return datetime.utcnow() + timedelta(days=QUOTE_VALIDITY_DAYS)
        return v

    @property
    def latest_version(self) -> Optional[QuoteVersion]:
        if not self.versions:
            return None
        return self.versions[-1]

    def is_expired(self, at: Optional[datetime] = None) -> bool:
        check_time = at or datetime.utcnow()
        return check_time > self.valid_until
