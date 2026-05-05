from typing import Optional
from uuid import UUID

from fastapi import APIRouter, Depends, Query

from shared.models.enums import OrderStatus
from shared.protocols.common import ApiResponse
from shared.protocols.order import (
    OrderConfirmRequest,
    OrderListResponse,
    OrderResponse,
)
from server.services.exceptions import BusinessException
from server.services.order_service import OrderService
from server.services.supplier_service import SupplierService

router = APIRouter(prefix="/orders", tags=["orders"])


def get_order_service() -> OrderService:
    return OrderService()


def get_supplier_service() -> SupplierService:
    return SupplierService()


@router.post("/from-quote/{quote_id}", response_model=ApiResponse[OrderResponse])
async def create_order_from_quote(
    quote_id: UUID,
    order_service: OrderService = Depends(get_order_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[OrderResponse]:
    try:
        order = order_service.create_order_from_quote(quote_id)
        supplier = supplier_service.get_supplier(order.supplier_id)

        response = OrderResponse.model_validate(order)
        response.supplier_name = supplier.name

        return ApiResponse.create_success(data=response)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("/{order_id}", response_model=ApiResponse[OrderResponse])
async def get_order(
    order_id: UUID,
    order_service: OrderService = Depends(get_order_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[OrderResponse]:
    try:
        order = order_service.get_order(order_id)
        supplier = supplier_service.get_supplier(order.supplier_id)

        response = OrderResponse.model_validate(order)
        response.supplier_name = supplier.name

        return ApiResponse.create_success(data=response)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("/number/{order_number}", response_model=ApiResponse[OrderResponse])
async def get_order_by_number(
    order_number: str,
    order_service: OrderService = Depends(get_order_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[OrderResponse]:
    try:
        order = order_service.get_order_by_number(order_number)
        supplier = supplier_service.get_supplier(order.supplier_id)

        response = OrderResponse.model_validate(order)
        response.supplier_name = supplier.name

        return ApiResponse.create_success(data=response)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.post("/{order_id}/confirm", response_model=ApiResponse[OrderResponse])
async def confirm_order(
    order_id: UUID,
    request: OrderConfirmRequest,
    order_service: OrderService = Depends(get_order_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[OrderResponse]:
    try:
        order = order_service.confirm_order(order_id, request.remarks)
        supplier = supplier_service.get_supplier(order.supplier_id)

        response = OrderResponse.model_validate(order)
        response.supplier_name = supplier.name

        return ApiResponse.create_success(data=response)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.post("/{order_id}/start", response_model=ApiResponse[OrderResponse])
async def start_order(
    order_id: UUID,
    order_service: OrderService = Depends(get_order_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[OrderResponse]:
    try:
        order = order_service.start_order(order_id)
        supplier = supplier_service.get_supplier(order.supplier_id)

        response = OrderResponse.model_validate(order)
        response.supplier_name = supplier.name

        return ApiResponse.create_success(data=response)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.post("/{order_id}/complete", response_model=ApiResponse[OrderResponse])
async def complete_order(
    order_id: UUID,
    order_service: OrderService = Depends(get_order_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[OrderResponse]:
    try:
        order = order_service.complete_order(order_id)
        supplier = supplier_service.get_supplier(order.supplier_id)

        response = OrderResponse.model_validate(order)
        response.supplier_name = supplier.name

        return ApiResponse.create_success(data=response)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.post("/{order_id}/cancel", response_model=ApiResponse[OrderResponse])
async def cancel_order(
    order_id: UUID,
    order_service: OrderService = Depends(get_order_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[OrderResponse]:
    try:
        order = order_service.cancel_order(order_id)
        supplier = supplier_service.get_supplier(order.supplier_id)

        response = OrderResponse.model_validate(order)
        response.supplier_name = supplier.name

        return ApiResponse.create_success(data=response)
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)


@router.get("", response_model=ApiResponse[OrderListResponse])
async def list_orders(
    status: Optional[OrderStatus] = Query(default=None, description="按状态筛选"),
    supplier_id: Optional[UUID] = Query(default=None, description="按供应商筛选"),
    order_service: OrderService = Depends(get_order_service),
    supplier_service: SupplierService = Depends(get_supplier_service),
) -> ApiResponse[OrderListResponse]:
    try:
        if supplier_id is not None:
            orders = order_service.list_orders_by_supplier(supplier_id)
        else:
            orders = order_service.list_orders(status)

        responses: list[OrderResponse] = []
        for order in orders:
            supplier = supplier_service.get_supplier(order.supplier_id)
            response = OrderResponse.model_validate(order)
            response.supplier_name = supplier.name
            responses.append(response)

        return ApiResponse.create_success(
            data=OrderListResponse(orders=responses, total=len(responses))
        )
    except BusinessException as e:
        return ApiResponse.create_error(code=e.code, message=e.message)
