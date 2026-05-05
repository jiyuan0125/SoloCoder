from datetime import datetime, timezone
from typing import Any

from fastapi import APIRouter, Depends, HTTPException, status

from shared.models import (
    AlertResponse,
    BatchQueryRequest,
    BatchQueryResponse,
    CustomStatusUpdateRequest,
    TrackingNode,
    TrackingNodeCreate,
    Waybill,
    WaybillCreate,
    WaybillQueryResponse,
    WaybillSplitRequest,
    WaybillSplitResponse,
)
from shared.constants import ErrorCode, ERROR_MESSAGES
from server.exceptions import ShippingServiceException
from server.storage import ShippingService, get_service


router = APIRouter()


def get_shipping_service() -> ShippingService:
    return get_service()


def handle_exception(exc: ShippingServiceException) -> HTTPException:
    return HTTPException(
        status_code=status.HTTP_400_BAD_REQUEST,
        detail={
            "error_code": exc.error_code.value,
            "message": exc.message,
        },
    )


@router.post("/waybills", response_model=Waybill, status_code=status.HTTP_201_CREATED)
async def create_waybill(
    waybill_create: WaybillCreate,
    service: ShippingService = Depends(get_shipping_service),
) -> Waybill:
    try:
        return service.create_waybill(waybill_create)
    except ShippingServiceException as exc:
        raise handle_exception(exc)


@router.get("/waybills/{waybill_number}", response_model=WaybillQueryResponse)
async def get_waybill(
    waybill_number: str,
    service: ShippingService = Depends(get_shipping_service),
) -> WaybillQueryResponse:
    try:
        return service.query_waybill(waybill_number)
    except ShippingServiceException as exc:
        raise handle_exception(exc)


@router.post("/waybills/batch", response_model=BatchQueryResponse)
async def batch_query_waybills(
    request: BatchQueryRequest,
    service: ShippingService = Depends(get_shipping_service),
) -> BatchQueryResponse:
    try:
        return service.batch_query_waybills(request.waybill_numbers)
    except ShippingServiceException as exc:
        raise handle_exception(exc)


@router.post(
    "/waybills/{waybill_number}/nodes",
    response_model=TrackingNode,
    status_code=status.HTTP_201_CREATED,
)
async def add_tracking_node(
    waybill_number: str,
    node_create: TrackingNodeCreate,
    service: ShippingService = Depends(get_shipping_service),
) -> TrackingNode:
    try:
        return service.add_tracking_node(waybill_number, node_create)
    except ShippingServiceException as exc:
        raise handle_exception(exc)


@router.post("/waybills/split", response_model=WaybillSplitResponse)
async def split_waybill(
    split_request: WaybillSplitRequest,
    service: ShippingService = Depends(get_shipping_service),
) -> WaybillSplitResponse:
    try:
        return service.split_waybill(split_request)
    except ShippingServiceException as exc:
        raise handle_exception(exc)


@router.get("/waybills/{waybill_number}/subs", response_model=list[Waybill])
async def get_sub_waybills(
    waybill_number: str,
    service: ShippingService = Depends(get_shipping_service),
) -> list[Waybill]:
    try:
        waybill = service.get_waybill(waybill_number)
        if waybill is None:
            from server.exceptions import WaybillNotFoundException
            raise WaybillNotFoundException(waybill_number)
        return service.get_sub_waybills(waybill_number)
    except ShippingServiceException as exc:
        raise handle_exception(exc)


@router.put("/waybills/{waybill_number}/custom-status", response_model=Waybill)
async def update_custom_status(
    waybill_number: str,
    request: CustomStatusUpdateRequest,
    service: ShippingService = Depends(get_shipping_service),
) -> Waybill:
    try:
        return service.update_custom_status(
            waybill_number, request.custom_status, request.operator
        )
    except ShippingServiceException as exc:
        raise handle_exception(exc)


@router.post("/alerts/check", response_model=list[AlertResponse])
async def check_abnormal_waybills(
    service: ShippingService = Depends(get_shipping_service),
) -> list[AlertResponse]:
    current_time = datetime.now(timezone.utc)
    try:
        return service.check_abnormal_waybills(current_time)
    except ShippingServiceException as exc:
        raise handle_exception(exc)


@router.post("/alerts/generate", response_model=list[AlertResponse])
async def generate_alerts(
    service: ShippingService = Depends(get_shipping_service),
) -> list[AlertResponse]:
    current_time = datetime.now(timezone.utc)
    try:
        return service.generate_alerts(current_time)
    except ShippingServiceException as exc:
        raise handle_exception(exc)


@router.get("/health", response_model=dict[str, Any])
async def health_check() -> dict[str, Any]:
    return {"status": "healthy", "timestamp": datetime.now(timezone.utc).isoformat()}
