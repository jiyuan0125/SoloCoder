from typing import List
from uuid import UUID

from fastapi import APIRouter, HTTPException
from pydantic import UUID4

from shared.models import (
    Order,
    CreateOrderRequest,
    UpdateOrderRequest,
    PackingList,
    PackingSuggestion,
    ApiResponse,
)
from shared import protocols
from server.app.exceptions import PackingSystemException
from server.app.services import order_service
from server.app.services.packing_service import (
    get_packing_list,
    generate_packing_list_text,
)
from server.app.services.optimization_service import suggest_packing

router = APIRouter()


@router.post(protocols.ORDERS_ENDPOINT, response_model=ApiResponse)
async def create_order(request: CreateOrderRequest) -> ApiResponse:
    try:
        order = order_service.create_order(
            order_id=request.order_id,
            items=request.items,
        )
        return ApiResponse(
            success=True,
            data={
                "order": order.model_dump(mode="json"),
            },
            message="订单创建成功",
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.get(protocols.ORDERS_ENDPOINT, response_model=ApiResponse)
async def list_orders() -> ApiResponse:
    orders = order_service.get_all_orders()
    return ApiResponse(
        success=True,
        data={
            "orders": [o.model_dump(mode="json") for o in orders],
        },
    )


@router.get(protocols.ORDER_ENDPOINT, response_model=ApiResponse)
async def get_order(order_id: str) -> ApiResponse:
    try:
        order = order_service.get_order(order_id)
        return ApiResponse(
            success=True,
            data={
                "order": order.model_dump(mode="json"),
            },
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.put(protocols.ORDER_ENDPOINT, response_model=ApiResponse)
async def update_order(order_id: str, request: UpdateOrderRequest) -> ApiResponse:
    try:
        order = order_service.update_order(
            order_id=order_id,
            items=request.items,
        )
        return ApiResponse(
            success=True,
            data={
                "order": order.model_dump(mode="json"),
            },
            message="订单更新成功",
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.get(protocols.PACKING_LISTS_ENDPOINT, response_model=ApiResponse)
async def get_order_packing_list(order_id: str) -> ApiResponse:
    try:
        order = order_service.get_order(order_id)
        packing_list = get_packing_list(order_id)

        return ApiResponse(
            success=True,
            data={
                "order": order.model_dump(mode="json"),
                "packing_list": packing_list.model_dump(mode="json")
                if packing_list
                else None,
            },
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.get(protocols.PACKING_LIST_FORMAT_ENDPOINT, response_model=ApiResponse)
async def get_formatted_packing_list(order_id: str) -> ApiResponse:
    try:
        text = generate_packing_list_text(order_id)
        return ApiResponse(
            success=True,
            data={
                "order_id": order_id,
                "packing_list_text": text,
            },
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )


@router.get(protocols.PACKING_SUGGESTIONS_ENDPOINT, response_model=ApiResponse)
async def get_packing_suggestions(order_id: str) -> ApiResponse:
    try:
        suggestions = suggest_packing(order_id)
        return ApiResponse(
            success=True,
            data={
                "order_id": order_id,
                "suggestions": [s.model_dump(mode="json") for s in suggestions],
            },
        )
    except PackingSystemException as e:
        return ApiResponse(
            success=False,
            error_code=e.error_code,
            message=e.message,
        )
