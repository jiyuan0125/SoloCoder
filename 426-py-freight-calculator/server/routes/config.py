from decimal import Decimal
from typing import List, Optional

from fastapi import APIRouter, HTTPException

from shared.responses import (
    ConfigResponse,
    FullConfig,
    PricingConfig,
    RemoteAreaConfig,
    DiscountConfig,
    DiscountTier,
    ErrorResponse,
    ErrorDetail,
)
from shared.errors import ErrorCode

from server.config import app_config


router = APIRouter(prefix="/config", tags=["config"])


@router.get("/", response_model=ConfigResponse)
async def get_full_config() -> ConfigResponse:
    pricing = app_config.pricing
    remote = app_config.remote_areas
    discounts = app_config.discounts
    
    return ConfigResponse(
        success=True,
        data=FullConfig(
            pricing=PricingConfig(
                land_first_weight_price=pricing.land_first_weight_price,
                land_continue_weight_price=pricing.land_continue_weight_price,
                sea_20ft_price=pricing.sea_20ft_price,
                sea_40ft_price=pricing.sea_40ft_price,
                air_volumetric_divisor=pricing.air_volumetric_divisor,
                land_max_perimeter_meters=pricing.land_max_perimeter_meters,
                air_max_perimeter_meters=pricing.air_max_perimeter_meters,
                insurance_rate=pricing.insurance_rate,
                insurance_min_fee=pricing.insurance_min_fee,
                max_compensation_multiplier=pricing.max_compensation_multiplier,
            ),
            remote_areas=RemoteAreaConfig(
                areas=list(remote.areas),
                surcharge_rate=remote.surcharge_rate,
            ),
            discounts=DiscountConfig(
                tiers=[
                    DiscountTier(threshold=t.threshold, discount_rate=t.discount_rate)
                    for t in discounts.tiers
                ],
            ),
        ),
        message="获取配置成功",
    )


@router.put("/pricing", response_model=ConfigResponse)
async def update_pricing(
    land_first_weight_price: Optional[Decimal] = None,
    land_continue_weight_price: Optional[Decimal] = None,
    sea_20ft_price: Optional[Decimal] = None,
    sea_40ft_price: Optional[Decimal] = None,
    air_volumetric_divisor: Optional[Decimal] = None,
    land_max_perimeter_meters: Optional[Decimal] = None,
    air_max_perimeter_meters: Optional[Decimal] = None,
    insurance_rate: Optional[Decimal] = None,
    insurance_min_fee: Optional[Decimal] = None,
    max_compensation_multiplier: Optional[Decimal] = None,
) -> ConfigResponse:
    updates: dict[str, object] = {}
    
    if land_first_weight_price is not None:
        updates["land_first_weight_price"] = land_first_weight_price
    if land_continue_weight_price is not None:
        updates["land_continue_weight_price"] = land_continue_weight_price
    if sea_20ft_price is not None:
        updates["sea_20ft_price"] = sea_20ft_price
    if sea_40ft_price is not None:
        updates["sea_40ft_price"] = sea_40ft_price
    if air_volumetric_divisor is not None:
        updates["air_volumetric_divisor"] = air_volumetric_divisor
    if land_max_perimeter_meters is not None:
        updates["land_max_perimeter_meters"] = land_max_perimeter_meters
    if air_max_perimeter_meters is not None:
        updates["air_max_perimeter_meters"] = air_max_perimeter_meters
    if insurance_rate is not None:
        updates["insurance_rate"] = insurance_rate
    if insurance_min_fee is not None:
        updates["insurance_min_fee"] = insurance_min_fee
    if max_compensation_multiplier is not None:
        updates["max_compensation_multiplier"] = max_compensation_multiplier
    
    app_config.update_pricing(**updates)
    
    return await get_full_config()


@router.put("/remote-areas", response_model=ConfigResponse)
async def update_remote_areas(
    areas: Optional[List[str]] = None,
    surcharge_rate: Optional[Decimal] = None,
) -> ConfigResponse:
    if areas is not None and len(areas) == 0:
        raise HTTPException(
            status_code=400,
            detail=ErrorResponse(
                error=ErrorDetail(
                    code=ErrorCode.REMOTE_AREA_CONFIG_ERROR.value,
                    message="偏远地区列表不能为空",
                )
            ).model_dump(),
        )
    
    app_config.update_remote_areas(areas=areas, surcharge_rate=surcharge_rate)
    
    return await get_full_config()


@router.put("/discounts", response_model=ConfigResponse)
async def update_discounts(
    thresholds: List[Decimal],
    rates: List[Decimal],
) -> ConfigResponse:
    if len(thresholds) != len(rates):
        raise HTTPException(
            status_code=400,
            detail=ErrorResponse(
                error=ErrorDetail(
                    code=ErrorCode.DISCOUNT_CONFIG_ERROR.value,
                    message="阈值数量和折扣率数量必须一致",
                )
            ).model_dump(),
        )
    
    if len(thresholds) == 0:
        raise HTTPException(
            status_code=400,
            detail=ErrorResponse(
                error=ErrorDetail(
                    code=ErrorCode.DISCOUNT_CONFIG_ERROR.value,
                    message="折扣配置不能为空",
                )
            ).model_dump(),
        )
    
    tiers = [
        DiscountTier(threshold=t, discount_rate=r)
        for t, r in zip(thresholds, rates)
    ]
    
    app_config.update_discounts(tiers)
    
    return await get_full_config()
