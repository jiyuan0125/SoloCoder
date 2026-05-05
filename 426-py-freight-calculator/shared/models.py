from decimal import Decimal
from datetime import datetime
from enum import Enum
from typing import Optional, List

from pydantic import BaseModel, Field, field_validator


class TransportMode(str, Enum):
    LAND = "land"
    AIR = "air"
    SEA = "sea"


class ContainerType(str, Enum):
    CONTAINER_20FT = "20ft"
    CONTAINER_40FT = "40ft"


class Package(BaseModel):
    weight_kg: Decimal = Field(..., gt=0, description="实际重量(公斤)")
    length_cm: Decimal = Field(..., gt=0, description="长度(厘米)")
    width_cm: Decimal = Field(..., gt=0, description="宽度(厘米)")
    height_cm: Decimal = Field(..., gt=0, description="高度(厘米)")
    origin: str = Field(..., min_length=1, description="发货地址")
    destination: str = Field(..., min_length=1, description="收货地址")
    declared_value: Optional[Decimal] = Field(None, ge=0, description="申报货值，用于保价")
    container_type: Optional[ContainerType] = Field(None, description="海运集装箱类型")

    @field_validator("declared_value")
    @classmethod
    def validate_declared_value(cls, v: Optional[Decimal]) -> Optional[Decimal]:
        if v is not None and v < Decimal("0"):
            raise ValueError("申报货值不能为负数")
        return v


class SplitPackage(BaseModel):
    weight_kg: Decimal
    length_cm: Decimal
    width_cm: Decimal
    height_cm: Decimal
    base_freight: Decimal
    total_freight: Decimal


class PackageDetail(BaseModel):
    package_index: int
    base_freight: Decimal
    surcharge: Decimal
    insurance_fee: Decimal
    discount: Decimal
    final_amount: Decimal
    max_compensation: Decimal
    is_split: bool = False
    split_packages: Optional[List[SplitPackage]] = None


class CalculationRequest(BaseModel):
    transport_mode: TransportMode
    packages: List[Package] = Field(..., min_length=1, max_length=100)
    customer_id: Optional[str] = Field(None, description="客户ID，用于累计运费折扣")

    @field_validator("packages")
    @classmethod
    def validate_packages_count(cls, v: List[Package]) -> List[Package]:
        if len(v) > 100:
            raise ValueError("批量计算一次最多100个包裹")
        return v


class CalculationResult(BaseModel):
    request_id: str
    total_base_freight: Decimal
    total_surcharge: Decimal
    total_insurance_fee: Decimal
    total_discount: Decimal
    total_final_amount: Decimal
    package_details: List[PackageDetail]
    calculated_at: datetime
    is_cached: bool = False


class MonthlyStatistics(BaseModel):
    year_month: str
    total_packages: int
    total_base_freight: Decimal
    total_surcharge: Decimal
    total_insurance_fee: Decimal
    total_discount: Decimal
    total_final_amount: Decimal
    transport_mode_breakdown: dict[TransportMode, int]
