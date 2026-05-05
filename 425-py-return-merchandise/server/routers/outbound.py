from typing import List

from fastapi import APIRouter, HTTPException, status

from shared.models import (
    OutboundOrderCreate,
    OutboundOrderResponse,
    OutboundOrder,
)
from server.services.outbound_service import OutboundService

router = APIRouter()

_service = OutboundService()


def _to_response(order: OutboundOrder) -> OutboundOrderResponse:
    return OutboundOrderResponse(
        order_id=order.order_id,
        customer_id=order.customer_id,
        items=order.items,
        total_amount=order.total_amount,
        status=order.status,
        created_at=order.created_at,
        updated_at=order.updated_at,
        shipped_at=order.shipped_at,
        delivered_at=order.delivered_at,
    )


@router.post("/", response_model=OutboundOrderResponse, status_code=status.HTTP_201_CREATED)
async def create_outbound_order(request: OutboundOrderCreate) -> OutboundOrderResponse:
    try:
        order = _service.create_order(request)
        return _to_response(order)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.get("/{order_id}", response_model=OutboundOrderResponse)
async def get_outbound_order(order_id: str) -> OutboundOrderResponse:
    order = _service.get_order(order_id)
    if not order:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"出库单 {order_id} 不存在",
        )
    return _to_response(order)


@router.get("/", response_model=List[OutboundOrderResponse])
async def list_outbound_orders() -> List[OutboundOrderResponse]:
    orders = _service.list_orders()
    return [_to_response(o) for o in orders]


@router.post("/{order_id}/ship", response_model=OutboundOrderResponse)
async def ship_order(order_id: str) -> OutboundOrderResponse:
    try:
        order = _service.ship_order(order_id)
        return _to_response(order)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.post("/{order_id}/deliver", response_model=OutboundOrderResponse)
async def deliver_order(order_id: str) -> OutboundOrderResponse:
    try:
        order = _service.deliver_order(order_id)
        return _to_response(order)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )


@router.post("/{order_id}/complete", response_model=OutboundOrderResponse)
async def complete_order(order_id: str) -> OutboundOrderResponse:
    try:
        order = _service.complete_order(order_id)
        return _to_response(order)
    except ValueError as e:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=str(e),
        )
