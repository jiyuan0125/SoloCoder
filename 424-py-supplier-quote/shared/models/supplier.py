from datetime import date, datetime
from typing import Optional
from uuid import UUID

from pydantic import Field

from shared.models.base import BaseModel, TimestampMixin, UUIDMixin
from shared.models.enums import QualificationLevel, ReviewResult, SupplierStatus


class SupplierBase(BaseModel):
    name: str = Field(..., min_length=1, max_length=200, description="供应商名称")
    contact_person: str = Field(..., min_length=1, max_length=100, description="联系人")
    phone: str = Field(..., min_length=1, max_length=50, description="联系电话")
    address: str = Field(..., min_length=1, max_length=500, description="地址")


class Supplier(SupplierBase, UUIDMixin, TimestampMixin):
    qualification_level: QualificationLevel = Field(
        default=QualificationLevel.C, description="资质等级"
    )
    status: SupplierStatus = Field(default=SupplierStatus.PENDING, description="状态")
    last_review_date: Optional[date] = Field(default=None, description="上次资质评审日期")
    next_review_date: Optional[date] = Field(default=None, description="下次资质评审日期")


class SupplierQualificationReview(UUIDMixin, TimestampMixin):
    supplier_id: UUID = Field(..., description="供应商ID")
    review_date: date = Field(..., description="评审日期")
    previous_level: QualificationLevel = Field(..., description="评审前资质等级")
    result: ReviewResult = Field(..., description="评审结果")
    new_level: QualificationLevel = Field(..., description="评审后资质等级")
    reviewer: str = Field(..., min_length=1, max_length=100, description="评审人")
    remarks: str = Field(default="", max_length=1000, description="备注")
