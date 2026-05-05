from typing import List

from fastapi import APIRouter, HTTPException, status

from shared.models import (
    ReturnApplyRequest,
    ReturnApplyResponse,
    WarehouseReceiveRequest,
    InspectionRequest,
    StockInRequest,
    ReturnOrderResponse,
)
from server.models import ReturnOrder
from server.services.return_service import ReturnService

router = APIRouter()

_service = ReturnService()


def _to_response(ret: ReturnOrder) -> ReturnOrderResponse:
    return ReturnOrderResponse(
        return_order_id=ret.return_order_id,
        outbound_order_id=ret.outbound_order_id,
        status=ret.status,
        items=ret.items,
        reason=ret.reason,
        customer_note=ret.customer_note,
        created_at=ret.created_at,
        expiry_date=ret.expiry_date,
        warehouse_received_at=ret.warehouse_received_at,
        warehouse_received_by=ret.warehouse_received_by,
        inspected_at=ret.inspected_at,
        inspected_by=ret.inspected_by,
        stocked_in_at=ret.stocked_in_at,
        stocked_in_by=ret.stocked_in_by,
    )


@router.post("/apply", response_model=ReturnApplyResponse, status_code=status.HTTP_201_CREATED)
async def apply_return(request: ReturnApplyRequest) -> ReturnApplyResponse:
    try:
        return _service.apply_return(request)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.post("/receive", response_model=ReturnOrderResponse)
async def warehouse_receive(request: WarehouseReceiveRequest) -> ReturnOrderResponse:
    try:
        ret = _service.warehouse_receive(request)
        return _to_response(ret)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.post("/inspect", response_model=ReturnOrderResponse)
async def inspect_return(request: InspectionRequest) -> ReturnOrderResponse:
    try:
        ret = _service.inspect(request)
        return _to_response(ret)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.post("/stock-in", response_model=ReturnOrderResponse)
async def stock_in_return(request: StockInRequest) -> ReturnOrderResponse:
    try:
        ret = _service.stock_in(request)
        return _to_response(ret)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.get("/{return_id}", response_model=ReturnOrderResponse)
async def get_return_order(return_id: str) -> ReturnOrderResponse:
    try:
        ret = _service.get_return_order(return_id)
        return _to_response(ret)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=str(e),
        )


@router.get("/", response_model=List[ReturnOrderResponse])
async def list_return_orders() -> List[ReturnOrderResponse]:
    returns = _service.list_return_orders()
    return [_to_response(r) for r in returns]
