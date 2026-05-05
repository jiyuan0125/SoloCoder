from typing import Optional
from uuid import UUID

from fastapi import APIRouter, Depends, HTTPException, Query

from shared.constants.error_codes import ErrorCode
from shared.models.enums import SupplierStatus
from shared.protocols.common import ApiResponse
from shared.protocols.supplier import (
    SupplierApproveRequest,
    SupplierCreateRequest,
    SupplierListResponse,
    SupplierResponse,
    SupplierReviewRequest,
    SupplierUpdateRequest,
)
from server.services.exceptions import BusinessException
from server.services.supplier_service import SupplierService

router = APIRouter(prefix="/suppliers", tags=["suppliers"])


def get_supplier_service() -> SupplierService:
    return SupplierService()


@router.post("", response_model=ApiResponse[SupplierResponse])
async def create_supplier(
    request: SupplierCreateRequest,
    service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[SupplierResponse]:
    try:
        supplier = service.create_supplier(request)
        return ApiResponse.create_success(data=SupplierResponse.model_validate(supplier))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("/{supplier_id}", response_model=ApiResponse[SupplierResponse])
async def get_supplier(
    supplier_id: UUID,
    service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[SupplierResponse]:
    try:
        supplier = service.get_supplier(supplier_id)
        return ApiResponse.create_success(data=SupplierResponse.model_validate(supplier))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.put("/{supplier_id}", response_model=ApiResponse[SupplierResponse])
async def update_supplier(
    supplier_id: UUID,
    request: SupplierUpdateRequest,
    service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[SupplierResponse]:
    try:
        supplier = service.update_supplier(supplier_id, request)
        return ApiResponse.create_success(data=SupplierResponse.model_validate(supplier))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.post("/{supplier_id}/approve", response_model=ApiResponse[SupplierResponse])
async def approve_supplier(
    supplier_id: UUID,
    request: SupplierApproveRequest,
    service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[SupplierResponse]:
    try:
        supplier = service.approve_supplier(supplier_id, request.approved, request.remarks)
        return ApiResponse.create_success(data=SupplierResponse.model_validate(supplier))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.post("/{supplier_id}/suspend", response_model=ApiResponse[SupplierResponse])
async def suspend_supplier(
    supplier_id: UUID,
    service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[SupplierResponse]:
    try:
        supplier = service.suspend_supplier(supplier_id)
        return ApiResponse.create_success(data=SupplierResponse.model_validate(supplier))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.post("/{supplier_id}/reactivate", response_model=ApiResponse[SupplierResponse])
async def reactivate_supplier(
    supplier_id: UUID,
    service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[SupplierResponse]:
    try:
        supplier = service.reactivate_supplier(supplier_id)
        return ApiResponse.create_success(data=SupplierResponse.model_validate(supplier))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.post("/{supplier_id}/review", response_model=ApiResponse[SupplierResponse])
async def review_qualification(
    supplier_id: UUID,
    request: SupplierReviewRequest,
    service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[SupplierResponse]:
    try:
        review = service.review_qualification(supplier_id, request)
        supplier = service.get_supplier(supplier_id)
        return ApiResponse.create_success(data=SupplierResponse.model_validate(supplier))
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("", response_model=ApiResponse[SupplierListResponse])
async def list_suppliers(
    status: Optional[SupplierStatus] = Query(default=None, description="按状态筛选"),
    service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[SupplierListResponse]:
    suppliers = service.list_suppliers(status)
    response = SupplierListResponse(
        suppliers=[SupplierResponse.model_validate(s) for s in suppliers],
        total=len(suppliers),
    )
    return ApiResponse.create_success(data=response)
