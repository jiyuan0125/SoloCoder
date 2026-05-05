from datetime import date
from typing import Optional
from uuid import UUID

from pydantic import Field

from shared.models.base import BaseModel
from shared.models.enums import QualificationLevel, ReviewResult, SupplierStatus
from shared.models.supplier import Supplier, SupplierQualificationReview


class SupplierCreateRequest(BaseModel):
    name: str = Field(..., min_length=1, max_length=200, description="供应商名称")
    contact_person: str = Field(..., min_length=1, max_length=100, description="联系人")
    phone: str = Field(..., min_length=1, max_length=50, description="联系电话")
    address: str = Field(..., min_length=1, max_length=500, description="地址")


class SupplierUpdateRequest(BaseModel):
    name: Optional[str] = Field(default=None, min_length=1, max_length=200, description="供应商名称")
    contact_person: Optional[str] = Field(default=None, min_length=1, max_length=100, description="联系人")
    phone: Optional[str] = Field(default=None, min_length=1, max_length=50, description="联系电话")
    address: Optional[str] = Field(default=None, min_length=1, max_length=500, description="地址")


class SupplierApproveRequest(BaseModel):
    approved: bool = Field(..., description="是否通过")
    remarks: str = Field(default="", max_length=1000, description="备注")


class SupplierReviewRequest(BaseModel):
    result: ReviewResult = Field(..., description="评审结果")
    reviewer: str = Field(..., min_length=1, max_length=100, description="评审人")
    remarks: str = Field(default="", max_length=1000, description="备注")


class SupplierResponse(Supplier):
    pass


class SupplierListResponse(BaseModel):
    suppliers: list[SupplierResponse] = Field(..., description="供应商列表")
    total: int = Field(..., ge=0, description="总数量")


class SupplierReviewResponse(SupplierQualificationReview):
    pass
