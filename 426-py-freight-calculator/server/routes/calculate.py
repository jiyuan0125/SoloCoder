import uuid
from datetime import datetime
from decimal import Decimal
from typing import List, Any

from fastapi import APIRouter, HTTPException

from shared.models import (
    CalculationRequest,
    CalculationResult,
    PackageDetail,
    TransportMode,
    ContainerType,
)
from shared.responses import CalculationResponse, ErrorResponse, ErrorDetail
from shared.errors import ErrorCode, error_messages

from server.calculator import freight_calculator
from server.cache import freight_cache
from server.store import data_store
from server.config import app_config


router = APIRouter(prefix="/calculate", tags=["calculate"])


def _build_cache_key(request: CalculationRequest) -> dict[str, Any]:
    return {
        "transport_mode": request.transport_mode.value,
        "packages": [
            {
                "weight_kg": str(p.weight_kg),
                "length_cm": str(p.length_cm),
                "width_cm": str(p.width_cm),
                "height_cm": str(p.height_cm),
                "origin": p.origin,
                "destination": p.destination,
                "declared_value": str(p.declared_value) if p.declared_value else None,
                "container_type": p.container_type.value if p.container_type else None,
            }
            for p in request.packages
        ],
        "customer_id": request.customer_id,
    }


def _validate_request(request: CalculationRequest) -> None:
    if request.transport_mode == TransportMode.SEA:
        for pkg in request.packages:
            if pkg.container_type is None:
                raise HTTPException(
                    status_code=400,
                    detail=ErrorResponse(
                        error=ErrorDetail(
                            code=ErrorCode.INVALID_CONTAINER_TYPE.value,
                            message=error_messages[ErrorCode.INVALID_CONTAINER_TYPE],
                        )
                    ).model_dump(),
                )


@router.post("/", response_model=CalculationResponse)
async def calculate_freight(request: CalculationRequest) -> CalculationResponse:
    _validate_request(request)
    
    cache_key_data = _build_cache_key(request)
    cached_result = freight_cache.get(cache_key_data)
    
    if cached_result is not None:
        cached_result.is_cached = True
        return CalculationResponse(
            success=True,
            data=cached_result,
            message="使用缓存结果",
        )
    
    cumulative_amount = data_store.get_customer_cumulative(request.customer_id)
    discount_rate = app_config.get_discount_rate(cumulative_amount)
    
    package_details: List[PackageDetail] = []
    
    total_base_freight = Decimal("0")
    total_surcharge = Decimal("0")
    total_insurance_fee = Decimal("0")
    total_discount = Decimal("0")
    total_final_amount = Decimal("0")
    
    for idx, pkg in enumerate(request.packages):
        detail = freight_calculator.calculate_package(
            pkg,
            request.transport_mode,
            discount_rate,
        )
        detail.package_index = idx
        
        package_details.append(detail)
        
        total_base_freight += detail.base_freight
        total_surcharge += detail.surcharge
        total_insurance_fee += detail.insurance_fee
        total_discount += detail.discount
        total_final_amount += detail.final_amount
    
    result = CalculationResult(
        request_id=str(uuid.uuid4()),
        total_base_freight=total_base_freight,
        total_surcharge=total_surcharge,
        total_insurance_fee=total_insurance_fee,
        total_discount=total_discount,
        total_final_amount=total_final_amount,
        package_details=package_details,
        calculated_at=datetime.now(),
        is_cached=False,
    )
    
    freight_cache.set(cache_key_data, result)
    
    data_store.add_record(
        result=result,
        transport_mode=request.transport_mode,
        customer_id=request.customer_id,
    )
    
    return CalculationResponse(
        success=True,
        data=result,
        message="计算完成",
    )
