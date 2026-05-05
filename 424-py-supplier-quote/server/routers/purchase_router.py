from typing import Optional
from uuid import UUID

from fastapi import APIRouter, Depends, Query

from shared.models.enums import PurchaseStatus
from shared.protocols.common import ApiResponse
from shared.protocols.purchase import (
    PurchaseCreateRequest,
    PurchaseListResponse,
    PurchaseResponse,
    PurchaseUpdateRequest,
)
from server.services.exceptions import BusinessException
from server.services.purchase_service import PurchaseService

router = APIRouter(prefix="/purchases", tags=["purchases"])


def get_purchase_service() -> PurchaseService:
    return PurchaseService()


@router.post("", response_model=ApiResponse[PurchaseResponse])
async def create_purchase(
    request: PurchaseCreateRequest,
    service: PurchaseService = Depends(get_purchase_service),
) -> ApiResponse[PurchaseResponse]:
    try:
        purchase = service.create_purchase(request)
        return ApiResponse.create_success(data=PurchaseResponse.model_validate(purchase))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("/{purchase_id}", response_model=ApiResponse[PurchaseResponse])
async def get_purchase(
    purchase_id: UUID,
    service: PurchaseService = Depends(get_purchase_service),
) -> ApiResponse[PurchaseResponse]:
    try:
        purchase = service.get_purchase(purchase_id)
        return ApiResponse.create_success(data=PurchaseResponse.model_validate(purchase))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.put("/{purchase_id}", response_model=ApiResponse[PurchaseResponse])
async def update_purchase(
    purchase_id: UUID,
    request: PurchaseUpdateRequest,
    service: PurchaseService = Depends(get_purchase_service),
) -> ApiResponse[PurchaseResponse]:
    try:
        purchase = service.update_purchase(purchase_id, request)
        return ApiResponse.create_success(data=PurchaseResponse.model_validate(purchase))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.post("/{purchase_id}/publish", response_model=ApiResponse[PurchaseResponse])
async def publish_purchase(
    purchase_id: UUID,
    service: PurchaseService = Depends(get_purchase_service),
) -> ApiResponse[PurchaseResponse]:
    try:
        purchase = service.publish_purchase(purchase_id)
        return ApiResponse.create_success(data=PurchaseResponse.model_validate(purchase))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.post("/{purchase_id}/close", response_model=ApiResponse[PurchaseResponse])
async def close_purchase(
    purchase_id: UUID,
    service: PurchaseService = Depends(get_purchase_service),
) -> ApiResponse[PurchaseResponse]:
    try:
        purchase = service.close_purchase(purchase_id)
        return ApiResponse.create_success(data=PurchaseResponse.model_validate(purchase))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("", response_model=ApiResponse[PurchaseListResponse])
async def list_purchases(
    status: Optional[PurchaseStatus] = Query(default=None, description="按状态筛选"),
    published_only: bool = Query(default=False, description="仅显示已发布的"),
    service: PurchaseService = Depends(get_purchase_service),
) -> ApiResponse[PurchaseListResponse]:
    if published_only:
        purchases = service.list_published_purchases()
    else:
        purchases = service.list_purchases(status)

    response = PurchaseListResponse(
        purchases=[PurchaseResponse.model_validate(p) for p in purchases],
        total=len(purchases),
    )
    return ApiResponse.create_success(data=response)
