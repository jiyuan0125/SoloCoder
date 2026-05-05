from typing import Generic, TypeVar, Optional, Any
from decimal import Decimal

from pydantic import BaseModel

from shared.models import CalculationResult, MonthlyStatistics
from shared.errors import ErrorCode


T = TypeVar("T")


class ApiResponse(BaseModel, Generic[T]):
    success: bool
    data: Optional[T] = None
    message: Optional[str] = None


class CalculationResponse(ApiResponse[CalculationResult]):
    pass


class StatisticsResponse(ApiResponse[MonthlyStatistics]):
    pass


class RemoteAreaConfig(BaseModel):
    areas: list[str]
    surcharge_rate: Decimal


class DiscountTier(BaseModel):
    threshold: Decimal
    discount_rate: Decimal


class DiscountConfig(BaseModel):
    tiers: list[DiscountTier]


class PricingConfig(BaseModel):
    land_first_weight_price: Decimal
    land_continue_weight_price: Decimal
    sea_20ft_price: Decimal
    sea_40ft_price: Decimal
    air_volumetric_divisor: Decimal
    land_max_perimeter_meters: Decimal
    air_max_perimeter_meters: Decimal
    insurance_rate: Decimal
    insurance_min_fee: Decimal
    max_compensation_multiplier: Decimal


class FullConfig(BaseModel):
    pricing: PricingConfig
    remote_areas: RemoteAreaConfig
    discounts: DiscountConfig


class ConfigResponse(ApiResponse[FullConfig]):
    pass


class ErrorDetail(BaseModel):
    code: str
    message: str
    details: Optional[dict[str, Any]] = None


class ErrorResponse(BaseModel):
    success: bool = False
    error: ErrorDetail
